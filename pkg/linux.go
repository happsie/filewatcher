package pkg

import (
	"fmt"
	"log/slog"
	"unsafe"

	"golang.org/x/sys/unix"
)

type linuxFileWatcher struct {
	fd int
	wd int
}

func newLinuxFileWatcher() *linuxFileWatcher {
	return &linuxFileWatcher{}
}

func (lw *linuxFileWatcher) watch(dir string, eventChan chan WatchEvent) error {
	if eventChan == nil {
		return fmt.Errorf("an event channel cannot be nil")
	}
	fd, err := unix.InotifyInit()
	if err != nil {
		return err
	}
	defer unix.Close(fd)

	wd, err := unix.InotifyAddWatch(fd, dir, unix.IN_CREATE|unix.IN_MODIFY|unix.IN_DELETE)
	if err != nil {
		return err
	}
	lw.wd = wd
	buf := make([]byte, unix.SizeofInotifyEvent*500)
	slog.Info("watching directory", "dir", dir)
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
				ModificationType: modType,
				Name:             name,
			}
			eventChan <- watchEvent
			offset += unix.SizeofInotifyEvent + int(event.Len)
		}
	}
}

func (lw *linuxFileWatcher) unwatch() error {
	_, err := unix.InotifyRmWatch(lw.fd, uint32(lw.wd))
	return err
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
