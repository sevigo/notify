package watcher

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/sevigo/notify/core"
	"github.com/sevigo/notify/event"
)

// ActionToString maps Action value to string
func ActionToString(action event.ActionType) string {
	switch action {
	case event.FileAdded:
		return "added"
	case event.FileRemoved:
		return "removed"
	case event.FileModified:
		return "modified"
	case event.FileRenamedOldName, event.FileRenamedNewName:
		return "renamed"
	default:
		return "invalid"
	}
}

// DirectoryWatcher ...
type DirectoryWatcher struct {
	ignoreUserFolders map[string]map[string]bool
	ignoreUserFiles   map[string]map[string]bool
	acceptUserFiles   map[string]map[string]bool

	events chan event.Event
	errors chan event.Error

	event.Waiter
}

// Options ...
type Options struct {
}

// Callback holds information about watcher channels
type Callback struct {
	Stop  bool
	Pause bool
}

var watcher *DirectoryWatcher
var once sync.Once

var watchersCallbackMutex sync.Mutex
var watchersCallback = make(map[string]chan Callback)

func RegisterCallback(path string) chan Callback {
	cb := make(chan Callback)
	watchersCallbackMutex.Lock()
	defer watchersCallbackMutex.Unlock()
	watchersCallback[path] = cb
	return cb
}

func UnregisterCallback(path string) {
	watchersCallbackMutex.Lock()
	defer watchersCallbackMutex.Unlock()
	delete(watchersCallback, path)
}

func LookupForCallback(path string) (chan Callback, bool) {
	watchersCallbackMutex.Lock()
	defer watchersCallbackMutex.Unlock()
	data, ok := watchersCallback[path]
	return data, ok
}

// Create new global instance of file watcher
func Create(ctx context.Context, callbackCh chan event.Event, errorCh chan event.Error, options *Options) *DirectoryWatcher {
	once.Do(func() {
		go processContext(ctx)
		watcher = &DirectoryWatcher{
			ignoreUserFiles:   make(map[string]map[string]bool),
			ignoreUserFolders: make(map[string]map[string]bool),
			acceptUserFiles:   make(map[string]map[string]bool),
			events:            callbackCh,
			errors:            errorCh,

			Waiter: event.Waiter{
				EventCh:  callbackCh,
				Timeout:  1 * time.Second,
				MaxCount: 5,
			},
		}
	})
	return watcher
}

func (w *DirectoryWatcher) Event() chan event.Event {
	return w.events
}

func (w *DirectoryWatcher) Error() chan event.Error {
	return w.errors
}

func (w *DirectoryWatcher) setOptions(path string, options *core.WatchingOptions) {
	path = filepath.Clean(path)
	if options != nil {
		if len(options.IgnoreFiles) > 0 {
			w.ignoreUserFiles[filepath.Clean(path)] = make(map[string]bool)
			for _, file := range options.IgnoreFiles {
				w.ignoreUserFiles[path][filepath.Clean(file)] = true
			}
		}

		if len(options.IgnoreFolders) > 0 {
			w.ignoreUserFolders[path] = make(map[string]bool)
			for _, folder := range options.IgnoreFolders {
				w.ignoreUserFolders[path][filepath.Clean(folder)] = true
			}
		}
		// AcceptFiles wins over IgnoreFiles
		if len(options.AcceptFiles) > 0 {
			w.acceptUserFiles[path] = make(map[string]bool)
			for _, file := range options.AcceptFiles {
				w.acceptUserFiles[path][filepath.Clean(file)] = true
			}
		}
	}
}

func (w *DirectoryWatcher) scan(path string) error {
	fileDebug("DEBUG", fmt.Sprintf("scan(): starting recursive scanning from root [%q]", path))
	return filepath.Walk(path, func(absoluteFilePath string, fileInfo os.FileInfo, err error) error {
		if fileInfo.IsDir() {
			dir := fileInfo.Name()
			if ignoreFolders[dir] || w.ignoreUserFolders[path][dir] {
				fileDebug("DEBUG", fmt.Sprintf("dir [%s] is excluded from watching", absoluteFilePath))
				return filepath.SkipDir
			}

			if os.IsPermission(err) {
				fileDebug("DEBUG", fmt.Sprintf("dir [%s] is excluded from watching because of an error: %v", absoluteFilePath, err))
				return filepath.SkipDir
			}
		}

		if err != nil {
			fileError("ERROR", fmt.Errorf("can't scan [%s]: %v", path, err))
			return filepath.SkipDir
		}
		if !fileInfo.IsDir() {
			ext := filepath.Ext(fileInfo.Name())
			// acceptUserFiles wins over ignoreUserFiles
			if len(w.acceptUserFiles[path]) > 0 {
				if w.acceptUserFiles[path][ext] {
					fileChangeNotifier(absoluteFilePath, event.FileAdded, &event.AdditionalInfo{
						Size:    fileInfo.Size(),
						ModTime: fileInfo.ModTime(),
					})
					return nil
				}
			} else {
				if !w.ignoreUserFiles[path][ext] {
					fileChangeNotifier(absoluteFilePath, event.FileAdded, &event.AdditionalInfo{
						Size:    fileInfo.Size(),
						ModTime: fileInfo.ModTime(),
					})
				}
			}
		}
		return nil
	})
}

func (w *DirectoryWatcher) RescanAll() {
	fileDebug("DEBUG", "RescanAll(): event triggerd")
	for path := range watchersCallback {
		err := w.scan(path)
		if err != nil {
			fileError("CRITICAL", fmt.Errorf("cannot scan directory [%s]", path))
		}
	}
}

func processContext(ctx context.Context) {
	<-ctx.Done()
	watchersCallbackMutex.Lock()
	defer watchersCallbackMutex.Unlock()
	for _, ch := range watchersCallback {
		ch <- Callback{
			Stop: true,
		}
	}
}

// StopWatching sends a signal to stop watching a directory
func (w *DirectoryWatcher) StopWatching(watchDirectoryPath string) {
	ch, ok := LookupForCallback(watchDirectoryPath)
	if ok {
		ch <- Callback{
			Stop:  true,
			Pause: false,
		}
	}
}

func fileError(lvl string, err error) {
	// TODO: we can print out here if it is configured
	watcher.errors <- event.FormatError(lvl, err.Error())
}

func fileDebug(lvl string, msg string) {
	// TODO: we can print out here if it is configured
	watcher.errors <- event.FormatError(lvl, msg)
}

func fileChangeNotifier(absoluteFilePath string, action event.ActionType, info *event.AdditionalInfo) {
	fileDebug("DEBUG", fmt.Sprintf("file [%s], action [%s]", absoluteFilePath, ActionToString(action)))
	// notification event is registered for this path, wait for 5 secs
	wait, exists := watcher.LookupForFileNotification(absoluteFilePath)
	if exists {
		wait <- true
		return
	}
	watcher.RegisterFileNotification(absoluteFilePath)

	data := &event.Event{
		Path:   absoluteFilePath,
		Action: action,
	}
	if info != nil {
		data.Size = info.Size
		data.ModTime = info.ModTime
	}

	go watcher.Wait(data)
}
