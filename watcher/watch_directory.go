package watcher

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

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
	case event.FileRenamedOldName:
		return "renamedFrom"
	case event.FileRenamedNewName:
		return "renamedTo"
	default:
		return "invalid"
	}
}

// DirectoryWatcher represents a configuration for
// a single directory to watch
type DirectoryWatcher struct {
	events chan event.Event
	errors chan event.Error

	event.Waiter
}

// Options represents global options for the notify
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
	watchersCallbackMutex.Lock()
	defer watchersCallbackMutex.Unlock()
	cb := make(chan Callback)
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
			events: callbackCh,
			errors: errorCh,

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

func (w *DirectoryWatcher) scan(path string) error {
	path = filepath.Clean(path)
	fileDebug("DEBUG", fmt.Sprintf("scan(): starting recursive scanning from root [%q]", path))
	return filepath.Walk(path, func(absoluteFilePath string, fileInfo os.FileInfo, err error) error {
		if fileInfo.IsDir() {
			dir := fileInfo.Name()
			if ignoreFolders[dir] {
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
			fileChangeNotifier(absoluteFilePath, event.FileAdded, &event.AdditionalInfo{
				Size:    fileInfo.Size(),
				ModTime: fileInfo.ModTime(),
			})
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
		UnregisterCallback(watchDirectoryPath)
	}
}

func fileError(lvl string, err error) {
	// TODO: we can print to STDOUT here if this is globaly configured
	// fmt.Printf("fileDebug(): [%s] %s\n", lvl, msg)
	watcher.errors <- event.FormatError(lvl, err.Error())
}

func fileDebug(lvl string, msg string) {
	// TODO: we can print to STDOUT here if this is globaly configured
	// fmt.Printf("fileDebug(): [%s] %s\n", lvl, msg)
	watcher.errors <- event.FormatError(lvl, msg)
}

func fileChangeNotifier(absoluteFilePath string, action event.ActionType, info *event.AdditionalInfo) {
	fileDebug("DEBUG", fmt.Sprintf("file [%s], action [%s]", absoluteFilePath, ActionToString(action)))
	// notification event is registered for this path, wait for 5 secs
	data := &event.Event{
		Path:   absoluteFilePath,
		Action: action,
	}

	fileNotificationKey := absoluteFilePath
	wait, exists := watcher.LookupForFileNotification(fileNotificationKey)
	if exists {
		wait <- *data
		return
	}
	watcher.RegisterFileNotification(fileNotificationKey)
	if info != nil {
		data.Size = info.Size
		data.ModTime = info.ModTime
		data.OldName = info.OldName
	}

	go watcher.Wait(fileNotificationKey, data)
}
