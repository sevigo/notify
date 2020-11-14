// +build linux,!integration,!fake

package watcher

// #include "watch_linux.h"
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

var ignoreFolders = map[string]bool{}

// #define IN_ACCESS		0x00000001	/* File was accessed */
// #define IN_MODIFY		0x00000002	/* File was modified */
// #define IN_ATTRIB		0x00000004	/* Metadata changed */
// #define IN_CLOSE_WRITE	0x00000008	/* Writtable file was closed */
// #define IN_CLOSE_NOWRITE	0x00000010	/* Unwrittable file closed */
// #define IN_OPEN			0x00000020	/* File was opened */
// #define IN_MOVED_FROM	0x00000040	/* File was moved from X */
// #define IN_MOVED_TO		0x00000080	/* File was moved to Y */
// #define IN_CREATE		0x00000100	/* Subfile was created */
// #define IN_DELETE		0x00000200	/* Subfile was deleted */
// #define IN_DELETE_SELF	0x00000400	/* Self was deleted */
func convertMaskToAction(mask int) event.ActionType {
	switch mask {
	case 2, 8: // File was modified
		return event.ActionType(event.FileModified)
	case 256: // Subfile was created
		return event.ActionType(event.FileAdded)
	case 512: // Subfile was deleted
		return event.ActionType(event.FileRemoved)
	case 64: // File was moved from X
		return event.ActionType(event.FileRenamedOldName)
	case 128: // File was moved to Y
		return event.ActionType(event.FileRenamedNewName)
	default:
		return event.ActionType(event.Invalid)
	}
}

// StartWatching starts a CGO function for getting the notifications
func (w *DirectoryWatcher) StartWatching(root string, options *core.WatchingOptions) {
	if _, err := os.Stat(root); os.IsNotExist(err) {
		fileError("CRITICAL", fmt.Errorf("cannot start watching [%s]: no such directory", root))
		return
	}
	err := filepath.Walk(root, func(path string, f os.FileInfo, err error) error {
		if f.IsDir() {
			_, found := LookupForCallback(path)
			if found {
				fileDebug("INFO", fmt.Sprintf("directory [%s] is already watched", path))
				return nil
			}
			ch := RegisterCallback(path)
			fileDebug("INFO", fmt.Sprintf("start watching [%s]", path))
			go watchDir(root, path, ch)
		}
		return nil
	})

	if options.Rescan {
		err := w.scan(root)
		if err != nil {
			fileError("CRITICAL", err)
			return
		}
	}
	if err != nil {
		fileError("ERROR", err)
	}
}

func watchDir(rootDirToWatch string, subDir string, ch chan Callback) {
	croot := C.CString(rootDirToWatch)
	cdir := C.CString(subDir)
	defer func() {
		C.free(unsafe.Pointer(croot))
		C.free(unsafe.Pointer(cdir))
	}()
	go C.WatchDirectory(croot, cdir)
	for p := range ch {
		if p.Stop {
			fileError("INFO", fmt.Errorf("linux.watchDir() stop for %s", subDir))
			return
		}
	}
}

//export goCallbackFileChange
func goCallbackFileChange(croot, cpath, cfile *C.char, caction C.int) {
	path := strings.TrimSpace(C.GoString(cpath))
	file := strings.TrimSpace(C.GoString(cfile))
	action := convertMaskToAction(int(caction))
	absoluteFilePath := filepath.Join(path, file)
	fileChangeNotifier(absoluteFilePath, action, nil)
}
