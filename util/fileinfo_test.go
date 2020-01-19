package util

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestFileInformation_Checksum(t *testing.T) {
	tmpFile, err := ioutil.TempFile(os.TempDir(), "Checksum-test-*")
	if err != nil {
		t.Errorf("Cannot create temporary file %v", err)
	}
	defer os.Remove(tmpFile.Name())
	text := []byte("Lorem ipsum")
	if _, err = tmpFile.Write(text); err != nil {
		t.Errorf("Failed to write to temporary file %v", err)
	}

	type fields struct {
		absoluteFilePath string
	}
	tests := []struct {
		name    string
		fields  fields
		want    string
		wantErr bool
	}{
		{
			name: "case 1: valid file",
			fields: fields{
				absoluteFilePath: tmpFile.Name(),
			},
			wantErr: false,
			want:    "bcebaf1ff7595840715bb574e77af4f37284fca9eb19e20195d3b7dca0fa52d2",
		},
		{
			name: "case 2: not valid file path",
			fields: fields{
				absoluteFilePath: "/foo/bar",
			},
			wantErr: true,
			want:    "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &FileInformation{
				absoluteFilePath: tt.fields.absoluteFilePath,
			}
			got, err := i.Checksum()
			if (err != nil) != tt.wantErr {
				t.Errorf("FileInformation.Checksum() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("FileInformation.Checksum() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFileInformation_ContentType(t *testing.T) {
	tmpFile, err := ioutil.TempFile(os.TempDir(), "ContentType-test-*")
	if err != nil {
		t.Errorf("Cannot create temporary file %v", err)
	}
	defer os.Remove(tmpFile.Name())
	text := []byte("Lorem ipsum")
	if _, err = tmpFile.Write(text); err != nil {
		t.Errorf("Failed to write to temporary file %v", err)
	}

	type fields struct {
		absoluteFilePath string
	}
	tests := []struct {
		name    string
		fields  fields
		want    string
		wantErr bool
	}{
		{
			name: "case 1: valid file",
			fields: fields{
				absoluteFilePath: tmpFile.Name(),
			},
			wantErr: false,
			want:    "application/octet-stream",
		},
		{
			name: "case 2, invalid file",
			fields: fields{
				absoluteFilePath: "/foo/bar",
			},
			wantErr: true,
			want:    "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &FileInformation{
				absoluteFilePath: tt.fields.absoluteFilePath,
			}
			got, err := i.ContentType()
			if (err != nil) != tt.wantErr {
				t.Errorf("FileInformation.ContentType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("FileInformation.ContentType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetFileInformation(t *testing.T) {
	tmpFile, err := ioutil.TempFile(os.TempDir(), "GetFileInformation-test-*")
	if err != nil {
		t.Errorf("Cannot create temporary file %v", err)
	}
	defer os.Remove(tmpFile.Name())
	text := []byte("Lorem ipsum")
	if _, err = tmpFile.Write(text); err != nil {
		t.Errorf("Failed to write to temporary file %v", err)
	}

	type args struct {
		absoluteFilePath string
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "case 1: valid file",
			args: args{
				absoluteFilePath: tmpFile.Name(),
			},
			wantErr: false,
		},
		{
			name: "case 2: invalid file",
			args: args{
				absoluteFilePath: "/foo/bar",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetFileInformation(tt.args.absoluteFilePath)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetFileInformation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil {
				if got.Name() != filepath.Base(tmpFile.Name()) {
					t.Errorf("GetFileInformation.Name() got = %v, want %v", got.Name(), filepath.Base(tmpFile.Name()))
				}
				if got.Size() != 11 {
					t.Errorf("GetFileInformation.Size() got = %v, want %v", got.Size(), 11)
				}
			}
		})
	}
}
