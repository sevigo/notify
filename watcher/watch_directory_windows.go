// +build windows,!integration,!fake

package watcher

// #include "watch_windows.h"
import "C"
import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unsafe"

	"github.com/sevigo/notify/core"
	"github.com/sevigo/notify/event"
)

var eventCache chan event.Event

func init() {
	C.Setup()
	eventCache = make(chan event.Event, 1)
}

// exclude these folders from the recursive scan
var ignoreFolders = map[string]bool{
	// $ sign indicates that the folder is hidden
	`$Recycle.Bin`: true,
	`$WinREAgent`:  true,
	`$SysReset`:    true,
}

// StartWatching starts a CGO function for getting the notifications
func (w *DirectoryWatcher) StartWatching(path string, options *core.WatchingOptions) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		fileError("CRITICAL", fmt.Errorf("can't start watching [%s]: %v", path, err))
		return
	}

	_, found := LookupForCallback(path)
	if found {
		fileDebug("INFO", fmt.Sprintf("directory [%s] is already watched", path))
		return
	}

	ch := RegisterCallback(path)
	fileDebug("INFO", fmt.Sprintf("start watching [%s]", path))
	cpath := C.CString(path)
	defer func() {
		C.free(unsafe.Pointer(cpath))
	}()

	go func() {
		for p := range ch {
			if p.Stop {
				C.StopWatching(cpath)
			}
		}
	}()

	if options.Rescan {
		err := w.scan(path)
		if err != nil {
			fileError("CRITICAL", fmt.Errorf("can't scan [%s]: %v", path, err))
			return
		}
	}

	C.WatchDirectory(cpath)
	fileDebug("INFO", fmt.Sprintf("[%s] is not watched anymore", path))
}

//export goCallbackFileChange
func goCallbackFileChange(cpath, cfile *C.char, caction C.int) {
	path := strings.TrimSpace(C.GoString(cpath))
	file := strings.TrimSpace(C.GoString(cfile))
	action := event.ActionType(int(caction))

	absoluteFilePath := filepath.Join(path, file)
	switch action {
	case event.FileRenamedOldName:
		go waitForRenameToEvent(absoluteFilePath)
		return
	case event.FileRenamedNewName:
		eventCache <- event.Event{
			Path:   absoluteFilePath,
			Action: action,
		}
		return
	default:
		if ok := checkValidFile(absoluteFilePath, action); ok {
			fileChangeNotifier(absoluteFilePath, action, nil)
		}
	}
}

func checkValidFile(absoluteFilePath string, action event.ActionType) bool {
	// if the file is removed we are good and the event is valid
	if action == event.FileRemoved {
		return true
	}
	// we are checking this because windows tend to create some tmp files if this is a download files
	_, err := os.Stat(absoluteFilePath)
	return err == nil
}

// we assuming that the FileRenamedOldName and FileRenamedNewName are fired together by win api
func waitForRenameToEvent(oldPath string) {
	for {
		select {
		case e := <-eventCache:
			if e.Action == event.FileRenamedNewName {
				newPath := e.Path
				if ok := checkValidFile(newPath, e.Action); ok {
					fileChangeNotifier(newPath, e.Action, &event.AdditionalInfo{OldName: oldPath})
				}
			}
		case <-time.After(time.Second):
			return
		}
	}
}
