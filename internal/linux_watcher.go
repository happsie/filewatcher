package internal

import (
	"fmt"
	"io/fs"
	"log/slog"
	"os"
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
	Name      string
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
	_, err = unix.InotifyAddWatch(fd, dir, unix.IN_CREATE|unix.IN_MODIFY|unix.IN_DELETE)
	if err != nil {
		return err
	}
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
				slog.Error("error transforming modifcation type to event", "error", err)
				return err
			}
			var name string
			if event.Len > 0 {
				name = unix.ByteSliceToString(buf[offset+unix.SizeofInotifyEvent : offset+unix.SizeofInotifyEvent+int(event.Len)])
			}
			watchEvent := WatchEvent{
				EventType: modType,
				Name:      name,
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
