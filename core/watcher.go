package core

import "github.com/sevigo/notify/event"

type WatchingOptions struct {
	Rescan        bool
	ActionFilters []event.ActionType
	FileFilters   []string
}

// DirectoryWatcher interface
type DirectoryWatcher interface {
	Event() chan event.Event
	Error() chan event.Error
	Scan(path string) error
	StartWatching(path string, options *WatchingOptions)
	StopWatching(path string)
}
