package event

import "time"

// ActionType represents what happens with the file
type ActionType int

// MetaInfo is additional info for any type of event
type MetaInfo map[string]string

// Event ...
type Event struct {
	Action ActionType
	Path   string
	AdditionalInfo
}

type AdditionalInfo struct {
	Size    int64
	ModTime time.Time
	// used if file was renamed
	OldName string
}

const (
	// Invalid action is 0
	Invalid ActionType = iota
	// FileAdded - the file was added to the directory.
	FileAdded // 1
	// FileRemoved - the file was removed from the directory.
	FileRemoved // 2
	// FileModified - the file was modified. This can be a change in the time stamp or attributes.
	FileModified // 3
	// FileRenamedOldName - the file was renamed and this is the old name.
	FileRenamedOldName // 4
	// FileRenamedNewName - the file was renamed and this is the new name.
	FileRenamedNewName // 5
)
