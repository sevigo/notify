// +build fake

package notify

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sevigo/notify/core"
	"github.com/sevigo/notify/event"
	"github.com/sevigo/notify/watcher"
)

func TestSetup(t *testing.T) {
	tests := []struct {
		name    string
		options *watcher.Options
		path    string
		want    event.Event
	}{
		{
			name:    "case 1",
			options: &watcher.Options{},
			path:    "/foo/bar",
			want: event.Event{
				Action: event.FileAdded,
				Path:   "/foo/bar/test.txt",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := Setup(context.TODO(), tt.options)
			go w.StartWatching(tt.path, &core.WatchingOptions{})
			go func() {
				<-w.Error()
			}()
			e := <-w.Event()
			assert.Equal(t, e, tt.want)
		})
	}
}
