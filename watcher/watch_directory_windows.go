// +build windows,!integration

package watcher

// #include "watch_windows.h"
import "C"
import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unsafe"

	// fileinfo "github.com/bakkuappu/helium/pkg/util"
	"github.com/sevigo/notify/event"
)

func init() {
	C.Setup()
}

// StartWatching starts a CGO function for getting the notifications
func (w *DirectoryWatcher) StartWatching(path string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		fileError("CRITICAL", err)
		return
	}
	_, found := LookupForCallback(path)
	if found {
		fileDebug("INFO", fmt.Sprintf("directory [%s] is already watched", path))
		return
	}
	ch := RegisterCallback(path)

	fileDebug("INFO", fmt.Sprintf("start watching [%s]\n", path))
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

	C.WatchDirectory(cpath)
	fileDebug("INFO", fmt.Sprintf("[%s] is not watched anymore", path))
}

//export goCallbackFileChange
func goCallbackFileChange(cpath, cfile *C.char, caction C.int) {
	path := strings.TrimSpace(C.GoString(cpath))
	file := strings.TrimSpace(C.GoString(cfile))
	action := event.ActionType(int(caction))

	absoluteFilePath := filepath.Join(path, file)
	// fi, err := fileinfo.GetFileInformation(absoluteFilePath)

	// if err != nil {
	// 	fileError("WARN", err)
	// 	return
	// }

	fileChangeNotifier(absoluteFilePath, action)
}
