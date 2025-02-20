package internal

import (
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	//"path/filepath"
	"unsafe"

	"golang.org/x/sys/unix"
)

type ModificationType int32

const (
	Modified ModificationType = 0
	Deleted  ModificationType = 1
	Created  ModificationType = 2
)

type WatchEvent struct {
	EventType ModificationType
	FileName  string
}

type linuxFileWatcher struct{}

func NewLinuxFileWatcher() *linuxFileWatcher {
	return &linuxFileWatcher{}
}

func (lw *linuxFileWatcher) Watch(dir string, eventChan chan WatchEvent) error {
	if eventChan == nil {
		return fmt.Errorf("an event channel cannot be nil")
	}
	fd, err := unix.InotifyInit()
	if err != nil {
		return err
	}
	defer unix.Close(fd)

	files, err := fs.ReadDir(os.DirFS(dir), ".")
	if err != nil {
		return err
	}
	watchWithPath := make(map[int]string)
	_, err = unix.InotifyAddWatch(fd, dir, unix.IN_CREATE|unix.IN_MODIFY|unix.IN_DELETE)
	if err != nil {
		return err
	}/*
	for _, file := range files {
		watchDescriptor, err := unix.InotifyAddWatch(fd, filepath.Join(dir, file.Name()), unix.IN_CREATE|unix.IN_MODIFY|unix.IN_DELETE)
		if err != nil {
			return err
		}
		watchWithPath[watchDescriptor] = file.Name()
	}*/
	buf := make([]byte, unix.SizeofInotifyEvent*len(files)*500)
	slog.Info("watching", "dir", dir)
	for {
		n, err := unix.Read(fd, buf)
		if err != nil {
			return err
		}
		for offset := 0; offset < n; {
			event := (*unix.InotifyEvent)(unsafe.Pointer(&buf[offset]))
			modType, err := getModificationType(event.Mask)
			if err != nil {
				slog.Error("error transforming watch to event", "error", err)
				return err
			}
			slog.Info("event", "test", event)
			watchEvent := WatchEvent{
				EventType: modType,
				FileName:  watchWithPath[int(event.Wd)],
			}
			eventChan <- watchEvent
			offset += unix.SizeofInotifyEvent + int(event.Len)
		}
	}
}

func getModificationType(mask uint32) (ModificationType, error) {
	if mask&unix.IN_DELETE == unix.IN_DELETE {
		return Deleted, nil
	}
	if mask&unix.IN_CREATE == unix.IN_CREATE {
		return Created, nil
	}
	if mask&unix.IN_MODIFY == unix.IN_MODIFY {
		return Modified, nil
	}
	return -1, fmt.Errorf("unknown modification type")
}
