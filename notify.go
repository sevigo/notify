package notify

import (
	"context"

	"github.com/sevigo/notify/event"
	"github.com/sevigo/notify/watcher"
)

// DirectoryWatcher interface
type DirectoryWatcher interface {
	Event() chan event.Event
	Error() chan event.Error
	Scan(path string) error
	StartWatching(path string)
	StopWatching(path string)
}

// Setup returns a channel for file change notifications and errors
func Setup(ctx context.Context, options *watcher.Options) DirectoryWatcher {
	eventCh := make(chan event.Event)
	errorCh := make(chan event.Error)
	if options == nil {
		options = &watcher.Options{Rescan: true}
	}

	return watcher.Create(ctx, eventCh, errorCh, options)
}
