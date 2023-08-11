package store

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
)

var (
	InvalidStoreDirectoryErr = errors.New("invalid store directory")
)

type File struct {
	storeDirectory string
	allowWrites    bool

	cacheLock      sync.Mutex
	unwrittenCache map[string]Document
}

func NewFile(storeDirectory string, allowWrites bool) (Type, error) {
	if len(storeDirectory) == 0 {
		return nil, InvalidStoreDirectoryErr
	}
	if _, err := os.Stat(storeDirectory); os.IsNotExist(err) {
		if allowWrites {
			err = os.MkdirAll(storeDirectory, os.ModePerm)
		}
		if err != nil {
			return nil, fmt.Errorf("cannot access store directory: %v", err)
		}
	}
	return &File{
		storeDirectory: storeDirectory,
		allowWrites:    allowWrites,
		unwrittenCache: map[string]Document{},
	}, nil
}

func (f *File) Create(document Document) error {
	return f.Update(document)
}

func (f *File) Update(document Document) error {
	if !f.allowWrites {
		f.cacheLock.Lock()
		f.unwrittenCache[document.ID] = document
		f.cacheLock.Unlock()
		return nil
	}
	filePath := filepath.Join(f.storeDirectory, document.ID)
	fileDir := filepath.Dir(filePath)
	if _, err := os.Stat(fileDir); os.IsNotExist(err) {
		if err = os.MkdirAll(fileDir, os.ModePerm); err != nil {
			return fmt.Errorf("cannot create file path: %v, error: %v", document.ID, err)
		}
	}
	return ioutil.WriteFile(filePath, []byte(document.Content), 0666)
}

func (f *File) Read(id string) (Document, error) {
	if !f.allowWrites {
		f.cacheLock.Lock()
		d, ok := f.unwrittenCache[id]
		f.cacheLock.Unlock()
		if ok {
			return d, nil
		}
	}

	bytes, err := ioutil.ReadFile(filepath.Join(f.storeDirectory, id))
	if err != nil {
		return Document{}, fmt.Errorf("failed to read content from file: %v", err)
	}
	return Document{ID: id, Content: string(bytes)}, nil
}
