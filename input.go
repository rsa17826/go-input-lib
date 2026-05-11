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

	// UInput ioctls
	UI_SET_EVBIT  = 0x40045564
	UI_SET_KEYBIT = 0x40045565
	UI_SET_RELBIT = 0x40045566
	UI_DEV_CREATE = 0x5501

	EVIOCGRAB = 0x40044590

	KEY_ESC        = 1
	KEY_1          = 2
	KEY_2          = 3
	KEY_3          = 4
	KEY_4          = 5
	KEY_5          = 6
	KEY_6          = 7
	KEY_7          = 8
	KEY_8          = 9
	KEY_9          = 10
	KEY_0          = 11
	KEY_MINUS      = 12
	KEY_EQUAL      = 13
	KEY_BACKSPACE  = 14
	KEY_TAB        = 15
	KEY_Q          = 16
	KEY_W          = 17
	KEY_E          = 18
	KEY_R          = 19
	KEY_T          = 20
	KEY_Y          = 21
	KEY_U          = 22
	KEY_I          = 23
	KEY_O          = 24
	KEY_P          = 25
	KEY_LEFTBRACE  = 26
	KEY_RIGHTBRACE = 27
	KEY_ENTER      = 28
	KEY_LEFTCTRL   = 29
	KEY_A          = 30
	KEY_S          = 31
	KEY_D          = 32
	KEY_F          = 33
	KEY_G          = 34
	KEY_H          = 35
	KEY_J          = 36
	KEY_K          = 37
	KEY_L          = 38
	KEY_SEMICOLON  = 39
	KEY_APOSTROPHE = 40
	KEY_GRAVE      = 41
	KEY_LEFTSHIFT  = 42
	KEY_BACKSLASH  = 43
	KEY_Z          = 44
	KEY_X          = 45
	KEY_C          = 46
	KEY_V          = 47
	KEY_B          = 48
	KEY_N          = 49
	KEY_M          = 50
	KEY_COMMA      = 51
	KEY_DOT        = 52
	KEY_SLASH      = 53
	KEY_RIGHTSHIFT = 54
	KEY_KPASTERISK = 55
	KEY_LEFTALT    = 56
	KEY_SPACE      = 57
	KEY_CAPSLOCK   = 58
	KEY_F1         = 59
	KEY_F2         = 60
	KEY_F3         = 61
	KEY_F4         = 62
	KEY_F5         = 63
	KEY_F6         = 64
	KEY_F7         = 65
	KEY_F8         = 66
	KEY_F9         = 67
	KEY_F10        = 68
	KEY_NUMLOCK    = 69
	KEY_SCROLLLOCK = 70
	KEY_KP7        = 71
	KEY_KP8        = 72
	KEY_KP9        = 73
	KEY_KPMINUS    = 74
	KEY_KP4        = 75
	KEY_KP5        = 76
	KEY_KP6        = 77
	KEY_KPPLUS     = 78
	KEY_KP1        = 79
	KEY_KP2        = 80
	KEY_KP3        = 81
	KEY_KP0        = 82
	KEY_KPDOT      = 83
	KEY_F11        = 87
	KEY_F12        = 88
	KEY_KPENTER    = 96
	KEY_RIGHTCTRL  = 97
	KEY_KPSLASH    = 98
	KEY_SYSRQ      = 99
	KEY_RIGHTALT   = 100
	KEY_HOME       = 102
	KEY_UP         = 103
	KEY_PAGEUP     = 104
	KEY_LEFT       = 105
	KEY_RIGHT      = 106
	KEY_END        = 107
	KEY_DOWN       = 108
	KEY_PAGEDOWN   = 109
	KEY_INSERT     = 110
	KEY_DELETE     = 111
	KEY_PAUSE      = 119
	KEY_LEFTMETA   = 125
	KEY_RIGHTMETA  = 126
	KEY_COMPOSE    = 127
)
const EVIOCGNAME = 0x80ff4506

func GetDeviceName(path string) string {
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

func GetPersistentID(eventPath string) (string, error) {
	// eventPath is like "/dev/input/event4"
	absPath, err := filepath.Abs(eventPath)
	if err != nil {
		return "", err
	}

	matches, err := filepath.Glob("/dev/input/by-id/*")
	if err != nil {
		return "", err
	}
	for _, idPath := range matches {
		evalPath, err := filepath.EvalSymlinks(idPath)
		if err != nil {
			return "", err
		}
		if evalPath == absPath {
			return idPath, nil // Found the persistent ID
		}
	}
	return "", nil // No persistent ID found (likely a virtual device)
}

func GetDeviceFromIdOrName(input string) (string, error) {
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
				if GetDeviceName(path) == targetName {
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

func CreateVirtualMouse() (*os.File, error) {
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

// type input_event struct {
// 	Time  syscall.Timeval
// 	Type  uint16
// 	Code  uint16
// 	Value int32
// }

func GetDeviceToUser() {
	// 1. Get all persistent device paths
	files, err := os.ReadDir("/dev/input/")
	if err != nil {
		panic(err)
	}

	// Channel to receive the ID of the device that was touched
	foundChan := make(chan string)

	fmt.Println("Listening on all devices... Press any key on the target device.")

	for _, f := range files {
		if strings.HasPrefix(f.Name(), "event") {
			path := "/dev/input/" + f.Name()
			println(GetDeviceName(path))
			go func(p string) {
				f, err := os.Open(p)
				if err != nil {
					return
				}
				defer f.Close()

				var ev InputEvent
				for {
					err := binary.Read(f, binary.LittleEndian, &ev)
					if err != nil {
						return
					}

					// Type 1 = EV_KEY, Value 1 = Key Down
					if ev.Type == 1 && ev.Value == 1 {
						id, err := GetPersistentID(p)
						if err != nil {
							panic(err)
						}
						if id != "" {
							foundChan <- "id:" + strings.TrimPrefix(id, "/dev/input/by-id/")
						} else {
							foundChan <- "name:" + GetDeviceName(p)
						}
						return
					}
				}
			}(path)
		}
	}

	// Wait for the first device to send a keypress
	winningID := <-foundChan
	fmt.Printf("\nTarget Device Identified!\n")
	fmt.Printf("Persistent ID: \"%s\"\n", winningID)
	fmt.Println("Use this path in your code to ensure you get the same device every time.")
}
func WaitForDevice(idOrName string) string {
	for {
		path, err := GetDeviceFromIdOrName(idOrName)
		if err == nil && path != "" {
			return path
		}
		println("waiting for device:", idOrName)
		time.Sleep(500 * time.Millisecond)
	}
}
