package notify

// 保管

import (
	"context"

	"github.com/sevigo/notify/event"
)

// Watch ...
type Watch struct {
	ctx    context.Context
	Events chan event.Event
	Errors chan event.Error

	watcher *DirectoryWatcher
}

// Setup returns a channel for file change notifications and errors
func Setup(ctx context.Context, options *Options) *Watch {
	eventCh := make(chan event.Event)
	errorCh := make(chan event.Error)
	if options == nil {
		options = &Options{IgnoreDirectoies: true}
	}

	watcher := Create(ctx, eventCh, errorCh, options)
	w := &Watch{
		ctx:     ctx,
		Errors:  errorCh,
		Events:  eventCh,
		watcher: watcher,
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
