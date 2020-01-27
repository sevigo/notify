package notify

// 保管

import (
	"context"

	"github.com/sevigo/notify/event"
	"github.com/sevigo/notify/watcher"
)

// Watch ...
type Watch struct {
	ctx    context.Context
	Events chan event.Event
	Errors chan event.Error

	watcher *watcher.DirectoryWatcher
}

// Setup returns a channel for file change notifications and errors
func Setup(ctx context.Context, options *watcher.Options) *Watch {
	eventCh := make(chan event.Event)
	errorCh := make(chan event.Error)
	if options == nil {
		options = &watcher.Options{Rescan: true}
	}

	w := &Watch{
		ctx:     ctx,
		Errors:  errorCh,
		Events:  eventCh,
		watcher: watcher.Create(ctx, eventCh, errorCh, options),
	}

	return w
}

// StopWatching removes a watcher from a dir
func (w *Watch) StopWatching(dir string) {
	w.watcher.StopWatching(dir)
}

// StartWatching adds a watcher for a dir
func (w *Watch) StartWatching(dir string) {
	go w.watcher.StartWatching(dir)
}
