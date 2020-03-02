package notify

import (
	"context"

	"github.com/sevigo/notify/core"
	"github.com/sevigo/notify/event"
	"github.com/sevigo/notify/watcher"
)

// Setup returns a channel for file change notifications and errors
func Setup(ctx context.Context, options *watcher.Options) core.DirectoryWatcher {
	eventCh := make(chan event.Event)
	errorCh := make(chan event.Error)

	return watcher.Create(ctx, eventCh, errorCh, options)
}
