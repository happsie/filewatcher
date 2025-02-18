package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"
)

type DirectoryScanner interface {
	Scan(baseDir string, subDir string) ([]FileMeta, error)
	Find(path string) (bool, FileMeta)
}

type ModificationType int32

const (
	NoChange ModificationType = 0
	Modified ModificationType = 1
	Removed  ModificationType = 2
	Created  ModificationType = 3
)

type FileMeta struct {
	Type      ModificationType
	FileName  string
	Path      string
	Directory bool
	ModTime   time.Time
}

type CachedDirectoryScanner struct {
	fileMetaMap map[string]FileMeta
	events      []FileMeta
}

func NewCachedDirectoryScanner() *CachedDirectoryScanner {
	return &CachedDirectoryScanner{
		fileMetaMap: map[string]FileMeta{},
		events:      []FileMeta{},
	}
}

func (cds *CachedDirectoryScanner) Scan(baseDir string, subDir string) ([]FileMeta, error) {
	entries, err := fs.ReadDir(os.DirFS(baseDir), subDir)
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			return nil, err
		}
		// skip recursive directory for now
		if info.IsDir() {
			continue
		}
		path := filepath.Join(baseDir, info.Name())
		if cds.isNew(path) {
			meta := FileMeta{
				Type:      Created,
				Path:      path,
				FileName:  info.Name(),
				ModTime:   info.ModTime(),
				Directory: false,
			}
			cds.fileMetaMap[path] = meta
			cds.events = append(cds.events, meta)
			continue
		}
		ok, fileMeta := cds.Find(path)
		if !ok {
			return nil, fmt.Errorf("could not find file")
		}
		if info.ModTime().Compare(fileMeta.ModTime) == 1 {
			meta := FileMeta{
				Type:      Modified,
				Path:      path,
				FileName:  info.Name(),
				ModTime:   info.ModTime(),
				Directory: false,
			}
			cds.fileMetaMap[path] = meta
			cds.events = append(cds.events, meta)
			continue
		}
	}

	defer func() {
		cds.events = nil
	}()

	return cds.events, nil
}

func (cds CachedDirectoryScanner) Find(path string) (bool, FileMeta) {
	meta := cds.fileMetaMap[path]
	if meta == (FileMeta{}) {
		return false, FileMeta{}
	}
	return true, meta
}

func (cds *CachedDirectoryScanner) isNew(path string) bool {
	meta := cds.fileMetaMap[path]
	if meta == (FileMeta{}) {
		return true
	}
	return false
}
