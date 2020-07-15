// +build windows,!integration,!fake

package watcher

// #include "watch_windows.h"
import "C"
import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unsafe"

	"github.com/sevigo/notify/core"
	"github.com/sevigo/notify/event"
)

func init() {
	C.Setup()
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
		UnregisterCallback(path)
		C.free(unsafe.Pointer(cpath))
	}()

	go func() {
		for p := range ch {
			if p.Stop {
				C.StopWatching(cpath)
			}
		}
	}()

	w.setOptions(path, options)

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

	if ok := checkValidFile(absoluteFilePath, action); ok {
		fileChangeNotifier(absoluteFilePath, action, nil)
	}
}

func checkValidFile(absoluteFilePath string, action event.ActionType) bool {
	// if the file is removed we are good and the event is valid
	if action == event.FileRemoved {
		return true
	}
	//we are checking this because windows tend to create some tmp files if this is a download files
	_, err := os.Stat(absoluteFilePath)
	return err == nil
}
