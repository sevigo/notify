package fileutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestChecksum(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		want     string
		wantErr  bool
	}{
		{
			name:     "case 1",
			filePath: "testdata/test.txt",
			want:     "06ae84e5d26e5537ee9b7b732fb2c091f72884b920102e5ecf3f2b13a6dd1933",
			wantErr:  false,
		},
		{
			name:     "case 2",
			filePath: "testdata/wrong.txt",
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Checksum(tt.filePath)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestContentType(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		want     string
		wantErr  bool
	}{
		{
			name:     "case 1",
			filePath: "testdata/test.txt",
			want:     "application/octet-stream",
			wantErr:  false,
		},
		{
			name:     "case 2",
			filePath: "testdata/wrong.txt",
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ContentType(tt.filePath)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
