package watcher

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/sevigo/notify/event"
	"github.com/sevigo/notify/util"
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
	ActionFilters []event.ActionType
	FileFilters   []string
	Options       *Options
	Events        chan event.Event
	Errors        chan event.Error
	StopWatchCh   chan string

	NotificationWaiter event.Waiter
}

// DirectoryWatcherImplementer ...
type DirectoryWatcherImplementer interface {
	Scan(path string) error
	StartWatching(path string)
	StopWatching(path string)
}

// Options ...
type Options struct {
	ActionFilters    []event.ActionType
	FileFilters      []string
	IgnoreDirectoies bool
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
			Options: options,
			Events:  callbackCh,
			Errors:  errorCh,
			NotificationWaiter: event.Waiter{
				EventCh:  callbackCh,
				Timeout:  time.Duration(5 * time.Second),
				MaxCount: 10,
			},
		}
	})
	return watcher
}

func (w *DirectoryWatcher) Scan(path string) error {
	return filepath.Walk(path, func(absoluteFilePath string, fileInfo os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !fileInfo.IsDir() {
			relativePath, err := filepath.Rel(path, absoluteFilePath)
			if err != nil {
				return err
			}
			w.CreateFileAddedNotification(path, relativePath)
		}
		return nil
	})
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
	watcher.Errors <- event.FormatError(lvl, err.Error())
}

func fileDebug(lvl string, msg string) {
	// TODO: we can print out here if it is configured
	watcher.Errors <- event.FormatError(lvl, msg)
}

// CreateFileAddedNotification creates a notifaction
func (w *DirectoryWatcher) CreateFileAddedNotification(watchDirectoryPath, relativeFilePath string) {
	absoluteFilePath := filepath.Join(watchDirectoryPath, relativeFilePath)
	fi, err := util.GetFileInformation(absoluteFilePath)

	if err != nil {
		fileError("WARN", err)
		return
	}

	fileChangeNotifier(absoluteFilePath, event.FileAdded)
}

func fileChangeNotifier(absoluteFilePath string, action event.ActionType) {
	for _, fileFilter := range watcher.Options.FileFilters {
		if strings.Contains(absoluteFilePath, fileFilter) {
			fileDebug("DEBUG", fmt.Sprintf("file [%s] is filtered", fileFilter))
			return
		}
	}

	for _, actionFilter := range watcher.ActionFilters {
		if action == actionFilter {
			fileDebug("DEBUG", fmt.Sprintf("action [%s] is filtered\n", ActionToString(actionFilter)))
			return
		}
	}

	fileDebug("DEBUG", fmt.Sprintf("file [%s], action [%s]\n", absoluteFilePath, ActionToString(action)))
	// notification event is registered for this path, wait for 5 secs
	wait, exists := watcher.NotificationWaiter.LookupForFileNotification(absoluteFilePath)
	if exists {
		wait <- true
		return
	}

	watcher.NotificationWaiter.RegisterFileNotification(absoluteFilePath)

	host, _ := os.Hostname()
	mimeType, err := util.ContentType(absoluteFilePath)
	if err != nil {
		fileError("WARNING", fmt.Errorf("can't get ContentType from the file [%s]: %v", absoluteFilePath, err))
		watcher.NotificationWaiter.UnregisterFileNotification(absoluteFilePath)
		return
	}
	checksum, err := util.Checksum(absoluteFilePath)
	if err != nil {
		fileError("WARNING", fmt.Errorf("can't get Checksum from the file [%s]: %v", absoluteFilePath, err))
		watcher.NotificationWaiter.UnregisterFileNotification(absoluteFilePath)
		return
	}

	data := &event.Event{
		UUID:     uuid.New(),
		MimeType: mimeType,
		Checksum: checksum,
		Path:     absoluteFilePath,
		Action:   action,
	}

	go watcher.NotificationWaiter.Wait(data)
}
