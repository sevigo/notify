//go:build darwin
// +build darwin

package watcher

import (
	"log/slog"

	"github.com/fsnotify/fsnotify"

	"github.com/sevigo/notify/core"
	"github.com/sevigo/notify/event"
	"github.com/sevigo/notify/fileutil"
)

// exclude these folders from the recursive scan
var ignoreFolders = map[string]bool{}

func (w *DirectoryWatcher) StartWatching(path string, _ *core.WatchingOptions) {
	var err error
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		slog.Error("can't start directory watcher", "error", err, "path", path)
		return
	}
	defer watcher.Close()

	// Processing all events and errors
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}

				slog.Info("processing event:", "operation", event.Op, "file", event.Name)
				mappedEvent, ok := mapEvent(event.Op)
				if !ok {
					continue
				}
				w.notify(event.Name, mappedEvent)

			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				slog.Error("error event", "error", err)
			}
		}
	}()

	slog.Info("adding new path to the wacher list", "path", path)

	if err := watcher.Add(path); err != nil {
		slog.Error("can't add file path to the watcher", "error", err, "path", path)
		return
	}
	<-make(chan struct{}) // Block forever
}

// notify translates fsnotify events to custom notification events
func (dw *DirectoryWatcher) notify(absoluteFilePath string, action event.ActionType) {
	fileInfo, err := fileutil.CheckValidFile(absoluteFilePath, action)
	if err != nil {
		slog.Error("file is invalid", "error", err, "path", absoluteFilePath)
		return
	}

	fileChangeNotifier(absoluteFilePath, action, &event.AdditionalInfo{
		Size:    fileInfo.Size(),
		ModTime: fileInfo.ModTime(),
	})
}

// mapEvent maps fsnotify's events to custom event types
func mapEvent(op fsnotify.Op) (event.ActionType, bool) {
	switch op {
	case fsnotify.Write:
		return event.FileModified, true
	case fsnotify.Create:
		return event.FileAdded, true
	case fsnotify.Remove:
		return event.FileRemoved, true
	case fsnotify.Rename:
		return event.FileRenamedOldName, true
	default:
		return event.Invalid, false
	}
}
