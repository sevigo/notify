//go:build darwin
// +build darwin

package watcher

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"

	"github.com/sevigo/notify/core"
	"github.com/sevigo/notify/event"
	"github.com/sevigo/notify/fileutil"
)

// exclude these folders from the recursive scan
var ignoreFolders = map[string]bool{}

func (w *DirectoryWatcher) initializeWatcher() (*fsnotify.Watcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		slog.Error("can't start directory watcher", "error", err)
		return nil, err
	}

	return watcher, nil
}

func (w *DirectoryWatcher) addDirectoriesRecursively(watcher *fsnotify.Watcher, path string) error {
	return filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			slog.Info("adding new path to the watcher list", "path", p)
			if err := watcher.Add(p); err != nil {
				slog.Error("can't add file path to the watcher", "error", err, "path", p)
				return err
			}
		}
		return nil
	})
}

func (w *DirectoryWatcher) handleEvents(watcher *fsnotify.Watcher) {
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
}

func (w *DirectoryWatcher) StartWatching(path string, opt *core.WatchingOptions) {
	watcher, err := w.initializeWatcher()
	if err != nil {
		return
	}
	defer watcher.Close()

	// Start processing events in a separate goroutine
	go w.handleEvents(watcher)

	if opt.Recursive {
		err = w.addDirectoriesRecursively(watcher, path)
	} else {
		slog.Info("adding new path to the watcher list", "path", path)
		err = watcher.Add(path)
	}

	if err != nil {
		slog.Error("can't add path to the watcher", "error", err, "path", path)
		return
	}

	if opt.Rescan {
		err := w.scan(path)
		if err != nil {
			fileError("CRITICAL", fmt.Errorf("can't scan [%s]: %v", path, err))
			return
		}
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
