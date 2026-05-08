package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"

	"golang.org/x/sys/unix"
)

// InputEvent matches the 'input_event' struct in linux/input.h
type InputEvent struct {
	Time  syscall.Timeval
	Type  uint16
	Code  uint16
	Value int32
}

const (
	EV_KEY     = 0x01
	EV_REL     = 0x02
	EV_SYN     = 0x00
	BTN_LEFT   = 0x110
	BTN_RIGHT  = 0x111
	BTN_MIDDLE = 0x112
	REL_X      = 0x00
	REL_Y      = 0x01
	REL_WHEEL  = 0x08
	REL_HWHEEL = 0x06
	KEY_Z      = 44
	KEY_SCROLL = 70

	// UInput ioctls
	UI_SET_EVBIT  = 0x40045564
	UI_SET_KEYBIT = 0x40045565
	UI_SET_RELBIT = 0x40045566
	UI_DEV_CREATE = 0x5501

	EVIOCGRAB = 0x40044590
)
const EVIOCGNAME = 0x80ff4506

func getDeviceName(path string) string {
	f, err := os.Open(path)
	if err != nil {
		return "Unknown (Permission Denied)"
	}
	defer f.Close()

	// Create a buffer to hold the name (up to 256 chars)
	name := make([]byte, 256)

	// Perform the ioctl syscall
	_, _, errno := syscall.Syscall(
		syscall.SYS_IOCTL,
		f.Fd(),
		uintptr(EVIOCGNAME),
		uintptr(unsafe.Pointer(&name[0])),
	)

	if errno != 0 {
		return "Unknown Device"
	}

	// Trim the null characters from the buffer
	return string(bytes.Trim(name, "\x00"))
}

func getPersistentID(eventPath string) string {
	// eventPath is like "/dev/input/event4"
	absPath, _ := filepath.Abs(eventPath)

	matches, _ := filepath.Glob("/dev/input/by-id/*")
	for _, idPath := range matches {
		evalPath, _ := filepath.EvalSymlinks(idPath)
		if evalPath == absPath {
			return idPath // Found the persistent ID
		}
	}
	return "" // No persistent ID found (likely a virtual device)
}

func getDeviceFromIdOrName(input string) (string, error) {
	var devicePath string

	if strings.HasPrefix(input, "id:") {
		// Use the persistent symlink in /dev/input/by-id/
		idPart := strings.TrimPrefix(input, "id:")
		devicePath = filepath.Join("/dev/input/by-id", idPart)

	} else if strings.HasPrefix(input, "name:") {
		// Scan all events to find the one with the matching Name
		targetName := strings.TrimPrefix(input, "name:")

		files, err := os.ReadDir("/dev/input/")
		if err != nil {
			return "", err
		}

		for _, f := range files {
			if strings.HasPrefix(f.Name(), "event") {
				path := filepath.Join("/dev/input/", f.Name())
				if getDeviceName(path) == targetName {
					devicePath = path
					break
				}
			}
		}
	}

	if devicePath == "" {
		return "", fmt.Errorf("device not found for input: %s", input)
	}

	// Open the file and return the file pointer
	return devicePath, nil
}

type uinput_user_dev struct {
	Name      [80]byte // UINPUT_MAX_NAME_SIZE is usually 80
	ID        uint16   // Bus type
	Vendor    uint16
	Product   uint16
	Version   uint16
	FFEffects uint32
	AbsMax    [64]int32 // ABS_CNT is 64
	AbsMin    [64]int32
	AbsFuzz   [64]int32
	AbsFlat   [64]int32
}

func createVirtualMouse() (*os.File, error) {
	// Use unix.O_NONBLOCK or syscall.O_NONBLOCK
	fd, err := os.OpenFile("/dev/uinput", os.O_WRONLY|unix.O_NONBLOCK, 0660)
	if err != nil {
		return nil, err
	}

	// Use your local constants (removed "unix." prefix)
	ifd := int(fd.Fd())
	unix.IoctlSetInt(ifd, UI_SET_EVBIT, EV_KEY)
	unix.IoctlSetInt(ifd, UI_SET_KEYBIT, BTN_LEFT)
	unix.IoctlSetInt(ifd, UI_SET_KEYBIT, BTN_RIGHT)
	unix.IoctlSetInt(ifd, UI_SET_KEYBIT, BTN_MIDDLE)
	unix.IoctlSetInt(ifd, UI_SET_EVBIT, EV_REL)
	unix.IoctlSetInt(ifd, UI_SET_RELBIT, REL_X)
	unix.IoctlSetInt(ifd, UI_SET_RELBIT, REL_Y)
	unix.IoctlSetInt(ifd, UI_SET_RELBIT, REL_WHEEL)
	unix.IoctlSetInt(ifd, UI_SET_RELBIT, REL_HWHEEL)

	// Setup the virtual device metadata
	var usetup uinput_user_dev
	copy(usetup.Name[:], "Turbo-Mouse")
	usetup.ID = 0x0003 // BUS_USB

	// Write the setup struct to the file descriptor
	err = binary.Write(fd, binary.LittleEndian, usetup)
	if err != nil {
		return nil, err
	}

	// Finalize device creation
	unix.IoctlSetInt(int(fd.Fd()), UI_DEV_CREATE, 0)

	return fd, nil
}

type input_event struct {
	Time  syscall.Timeval
	Type  uint16
	Code  uint16
	Value int32
}
