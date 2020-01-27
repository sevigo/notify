package main

import (
	"context"
	"log"
	"runtime"

	othernotify "github.com/rjeczalik/notify"

	"github.com/sevigo/notify"
	"github.com/sevigo/notify/watcher"
)

var dirsWin = []string{"C:\\Users\\Igor\\Files"}
var dirsLin = []string{"/home/igor/Downloads", "/home/igor/Documents"}

func main() {
	log.Println("Starting the service ...")
	ctx := context.TODO()

	var dirs []string
	switch runtime.GOOS {
	case "windows":
		dirs = dirsWin
	case "linux":
		dirs = dirsLin
	default:
		panic("not supported OS")
	}

	w := notify.Setup(
		ctx,
		&watcher.Options{
			Rescan:      true,
			FileFilters: []string{".crdownload", ".lock", ".snapshot"},
		})

	for _, dir := range dirs {
		w.StartWatching(dir)
	}
	defer func() {
		for _, dir := range dirs {
			w.StopWatching(dir)
		}
	}()

	// wait for file change events
	for {
		select {
		case ev := <-w.Events:
			log.Printf("[EVENT] %s: %q", watcher.ActionToString(ev.Action), ev.Path)
		case err := <-w.Errors:
			if err.Level == "ERROR" {
				log.Printf("[%s] %s", err.Level, err.Message)
			}
		}
	}

}

func watchX(dir string) {
	c := make(chan othernotify.EventInfo, 1)
	defer othernotify.Stop(c)

	if err := othernotify.Watch(dir, c, othernotify.Create, othernotify.Remove); err != nil {
		log.Fatal(err)
	}

	for otherfile := range c {
		log.Println("[EVENT]", otherfile)
	}
}

// 2020/01/20 00:49:58 [X-EVENT] notify.Create: "C:\Users\Igor\Files\7b2528e7-8c97-4172-be9d-8f8cd757da70.jfif"
// 2020/01/20 00:49:58 [X-EVENT] notify.Remove: "C:\Users\Igor\Files\7b2528e7-8c97-4172-be9d-8f8cd757da70.jfif"
// 2020/01/20 00:49:58 [X-EVENT] notify.Create: "C:\Users\Igor\Files\7b2528e7-8c97-4172-be9d-8f8cd757da70.jfif.crdownload"
