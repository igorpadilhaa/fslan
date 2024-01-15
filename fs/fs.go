package fs

import (
	"io/fs"
	"mime"
	"os"
	"path/filepath"
	"strings"
)

type DynamicFs struct {
	files map[string]dFile
}

type dFile struct {
	Name string `json:"name"`
	Mime string `json:"mime"`
	Path string `json:"-"`
}

func NewDynamicFs() DynamicFs {
	return DynamicFs{map[string]dFile{}}
}

func (dfs DynamicFs) Open(path string) (fs.File, error) {
	path, _ = strings.CutPrefix(path, "/")
	if !fs.ValidPath(path) {
		return nil, &fs.PathError{
			Op:   "open",
			Path: path,
			Err:  fs.ErrInvalid,
		}
	}

	file, exists := dfs.files[path]
	if !exists {
		return nil, &fs.PathError{
			Op:   "open",
			Path: path,
			Err:  fs.ErrNotExist,
		}
	}

	return os.Open(file.Path)
}

func (dfs DynamicFs) Add(name, path string) {
	fileExtension := filepath.Ext(path)
	mimeType := mime.TypeByExtension(fileExtension)

	dfs.files[name] = dFile{
		Name: name,
		Mime: mimeType,
		Path: path,
	}
}

func (dfs DynamicFs) Files() []dFile {
	var files []dFile
	for _, file := range dfs.files {
		files = append(files, file)
	}
	return files
}
