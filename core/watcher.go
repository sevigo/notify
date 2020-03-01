package core

import "github.com/sevigo/notify/event"

// DirectoryWatcher interface
type DirectoryWatcher interface {
	Event() chan event.Event
	Error() chan event.Error
	Scan(path string) error
	StartWatching(path string)
	StopWatching(path string)
}
