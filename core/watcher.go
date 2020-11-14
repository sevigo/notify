package core

import "github.com/sevigo/notify/event"

type WatchingOptions struct {
	Rescan        bool
	ActionFilters []event.ActionType
}

// DirectoryWatcher interface
type DirectoryWatcher interface {
	Event() chan event.Event
	Error() chan event.Error
	RescanAll()
	StartWatching(path string, options *WatchingOptions)
	StopWatching(path string)
}
