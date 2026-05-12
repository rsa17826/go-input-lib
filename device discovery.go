package input

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"
	"unsafe"
)

// ── Device discovery ──────────────────────────────────────────────────────────

// DeviceName returns the kernel name of the input device at path.
func DeviceName(path string) string {
	f, err := os.Open(path)
	if err != nil {
		return "Unknown (Permission Denied)"
	}
	defer f.Close()

	name := make([]byte, 256)
	_, _, errno := syscall.Syscall(
		syscall.SYS_IOCTL,
		f.Fd(),
		uintptr(EVIOCGNAME),
		uintptr(unsafe.Pointer(&name[0])),
	)
	if errno != 0 {
		return "Unknown Device"
	}
	return string(bytes.Trim(name, "\x00"))
}

// PersistentID resolves an event path (e.g. /dev/input/event4) to its stable
// /dev/input/by-id symlink, if one exists.
func PersistentID(eventPath string) (string, error) {
	absPath, err := filepath.Abs(eventPath)
	if err != nil {
		return "", err
	}
	matches, err := filepath.Glob("/dev/input/by-id/*")
	if err != nil {
		return "", err
	}
	for _, idPath := range matches {
		resolved, err := filepath.EvalSymlinks(idPath)
		if err != nil {
			return "", err
		}
		if resolved == absPath {
			return idPath, nil
		}
	}
	return "", nil // no persistent ID (likely a virtual device)
}

// FindDevice resolves an "id:<name>" or "name:<name>" selector to an event path.
func FindDevice(selector string) (string, error) {
	switch {
	case strings.HasPrefix(selector, "id:"):
		id := strings.TrimPrefix(selector, "id:")
		return filepath.Join("/dev/input/by-id", id), nil

	case strings.HasPrefix(selector, "name:"):
		target := strings.TrimPrefix(selector, "name:")
		entries, err := os.ReadDir("/dev/input/")
		if err != nil {
			return "", err
		}
		for _, e := range entries {
			if strings.HasPrefix(e.Name(), "event") {
				path := filepath.Join("/dev/input", e.Name())
				if DeviceName(path) == target {
					return path, nil
				}
			}
		}
	}
	return "", fmt.Errorf("device not found: %s", selector)
}

// WaitForDevice blocks until the device matching selector becomes available,
// then returns its path.
func WaitForDevice(selector string) string {
	for {
		path, err := FindDevice(selector)
		if err == nil && path != "" {
			return path
		}
		fmt.Println("waiting for device:", selector)
		time.Sleep(500 * time.Millisecond)
	}
}

// PickDevice listens on all input devices and prints the selector string for
// the first one that receives a keypress. Use this to identify a device to
// pass to FindDevice or WaitForDevice.
func PickDevice() string {
	entries, err := os.ReadDir("/dev/input/")
	if err != nil {
		panic(err)
	}

	found := make(chan string, 1)
	fmt.Println("Listening on all devices — press any key on the target device.")

	for _, e := range entries {
		if !strings.HasPrefix(e.Name(), "event") {
			continue
		}
		path := filepath.Join("/dev/input", e.Name())
		fmt.Println(" ", DeviceName(path))

		go func(p string) {
			f, err := os.Open(p)
			if err != nil {
				return
			}
			defer f.Close()

			var ev InputEvent
			for {
				if err := binary.Read(f, binary.LittleEndian, &ev); err != nil {
					return
				}
				if ev.Type == EV_KEY && ev.Value == 1 {
					id, err := PersistentID(p)
					if err != nil || id == "" {
						found <- "name:" + DeviceName(p)
					} else {
						found <- "id:" + strings.TrimPrefix(id, "/dev/input/by-id/")
					}
					return
				}
			}
		}(path)
	}

	selector := <-found
	fmt.Printf("\nDevice identified: %q\n", selector)
	fmt.Println("Pass this string to WaitForDevice or FindDevice.")
	return selector
}
