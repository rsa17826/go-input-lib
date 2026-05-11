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

	BUS_USB = 0x0003
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

func CreateVirtualMouse(name string, block ...bool) (*VirtualMouse, error) {
	blocking := len(block) > 0 && block[0]

	flags := os.O_WRONLY
	if !blocking {
		flags |= unix.O_NONBLOCK
	}
	fd, err := os.OpenFile("/dev/uinput", flags, 0660)
	if err != nil {
		return nil, err
	}

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
	copy(usetup.Name[:], name)
	usetup.ID = BUS_USB

	// Write the setup struct to the file descriptor
	err = binary.Write(fd, binary.LittleEndian, usetup)
	if err != nil {
		return nil, err
	}

	// Finalize device creation
	unix.IoctlSetInt(int(fd.Fd()), UI_DEV_CREATE, 0)

	return &VirtualMouse{dev: fd}, nil
}
func (self *VirtualMouse) PressButton(button int) error {
}
func (self *VirtualMouse) Click(args ...ClickArg) error {
	var holdFor, afterDelay time.Duration

	for _, arg := range args {
		switch v := arg.(type) {
		case RawButton:
			if err := self.SendEvent(uint16(v), holdFor, afterDelay); err != nil {
				return err
			}

		case InputTiming:
			holdFor = v.HoldFor
			afterDelay = v.AfterDelay

		case DelayNow:
			time.Sleep(time.Duration(v))
		}
	}
	return nil
}

// PressArg is the sealed interface accepted by Press.
type ClickArg interface{ isClickArg() }

// RawKey presses a key directly by its kernel keycode (e.g. KEY_ENTER, KEY_F5).
// No shift handling — whatever code you pass is exactly what gets sent.
type RawButton uint16

func (RawButton) isClickArg() {}

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

type VirtualInput : VirtualKeyboard|VirtualMouse
type VirtualKeyboard struct {
	dev *os.File
}
type VirtualMouse struct {
	dev *os.File
}

func (kbd *VirtualMouse) Close() error {
	return kbd.dev.Close()
}
func (kbd *VirtualKeyboard) Close() error {
	return kbd.dev.Close()
}

// ── Argument types ────────────────────────────────────────────────────────────

// PressArg is the sealed interface accepted by Press.
type PressArg interface{ isPressArg() }

// RawKey presses a key directly by its kernel keycode (e.g. KEY_ENTER, KEY_F5).
// No shift handling — whatever code you pass is exactly what gets sent.
type RawKey uint16

func (RawKey) isPressArg() {}

// KeyString types each character in the string, automatically applying the
// correct keycode and Shift modifier for each rune.
//
// Example: KeyString(`asd"123!@#|\`) types those characters literally.
type KeyString string

func (KeyString) isPressArg() {}

// InputTiming changes the hold-duration and post-key delay for every key that
// follows it in the same Press call. Zero values mean no delay.
type InputTiming struct {
	HoldFor    time.Duration // how long to hold each key before releasing
	AfterDelay time.Duration // pause after each key is released
}

func (InputTiming) isPressArg() {}
func (InputTiming) isClickArg() {}

// DelayNow inserts an immediate pause at this point in the sequence.
// It does not affect the current KeyTiming.
//
// Example: Press(dev, KeyString("hello"), DelayNow(500*time.Millisecond), KeyString("world"))
type DelayNow time.Duration

func (DelayNow) isPressArg() {}
func (DelayNow) isClickArg() {}

// ── Press ─────────────────────────────────────────────────────────────────────

// Press sends a sequence of key events to dev (a virtual keyboard created by
// CreateVirtualKeyboard). Args can be any mix of RawKey, KeyString, KeyTiming,
// and DelayNow in any order.
//
// Example:
//
//	kbd, _ := CreateVirtualKeyboard()
//	kbd.Press(
//	    KeyTiming{HoldFor: 10*time.Millisecond, AfterDelay: 30*time.Millisecond},
//	    KeyString("Hello, World!"),
//	    DelayNow(time.Second),
//	    RawKey(KEY_ENTER),
//	)
func (kbd *VirtualKeyboard) Press(args ...PressArg) error {
	var holdFor, afterDelay time.Duration

	for _, arg := range args {
		switch v := arg.(type) {
		case RawKey:
			if err := kbd.TapKey(uint16(v), holdFor, afterDelay); err != nil {
				return err
			}

		case KeyString:
			for _, ch := range string(v) {
				info, ok := charKeyMap[ch]
				if !ok {
					return fmt.Errorf("press: no keycode mapping for character %q", ch)
				}
				if err := kbd.TapKeyWithShift(info.code, info.shift, holdFor, afterDelay); err != nil {
					return err
				}
			}

		case InputTiming:
			holdFor = v.HoldFor
			afterDelay = v.AfterDelay

		case DelayNow:
			time.Sleep(time.Duration(v))
		}
	}
	return nil
}

// ── Virtual keyboard ──────────────────────────────────────────────────────────

// CreateVirtualKeyboard creates a uinput virtual keyboard device and returns
// its file descriptor. The caller is responsible for closing it.
func CreateVirtualKeyboard(name string, block ...bool) (*VirtualKeyboard, error) {
	blocking := len(block) > 0 && block[0]

	flags := os.O_WRONLY
	if !blocking {
		flags |= unix.O_NONBLOCK
	}
	fd, err := os.OpenFile("/dev/uinput", flags, 0660)
	if err != nil {
		return nil, err
	}

	ifd := int(fd.Fd())

	// Register EV_KEY support
	unix.IoctlSetInt(ifd, UI_SET_EVBIT, EV_KEY)

	// Register every key we know about
	for code := range allKeyCodes {
		unix.IoctlSetInt(ifd, UI_SET_KEYBIT, int(code))
	}

	var usetup uinput_user_dev
	copy(usetup.Name[:], name)
	usetup.ID = BUS_USB //

	if err := binary.Write(fd, binary.LittleEndian, usetup); err != nil {
		fd.Close()
		return nil, err
	}

	unix.IoctlSetInt(int(fd.Fd()), UI_DEV_CREATE, 0)

	return &VirtualKeyboard{dev: fd}, nil
}

func (self *VirtualInput) SendEvent(evType, code uint16, value int32) error {
	ev := InputEvent{Type: evType, Code: code, Value: value}
	return binary.Write(self.dev, binary.LittleEndian, ev)
}

func (kbd *VirtualKeyboard) Sync() error {
	return kbd.SendEvent(EV_SYN, 0, 0)
}

// TapKey presses and releases a single keycode, with optional hold and post delay.
func (kbd *VirtualKeyboard) TapKey(code uint16, holdFor, afterDelay time.Duration) error {
	if err := kbd.SendEvent(EV_KEY, code, 1); err != nil { // key down
		return err
	}
	if err := kbd.Sync(); err != nil {
		return err
	}
	if holdFor > 0 {
		time.Sleep(holdFor)
	}
	if err := kbd.SendEvent(EV_KEY, code, 0); err != nil { // key up
		return err
	}
	if err := kbd.Sync(); err != nil {
		return err
	}
	if afterDelay > 0 {
		time.Sleep(afterDelay)
	}
	return nil
}

// TapKeyWithShift wraps tapKey with an optional Left Shift hold around it.
func (kbd *VirtualKeyboard) TapKeyWithShift(code uint16, shift bool, holdFor, afterDelay time.Duration) error {
	if shift {
		if err := kbd.SendEvent(EV_KEY, KEY_LEFTSHIFT, 1); err != nil {
			return err
		}
		if err := kbd.Sync(); err != nil {
			return err
		}
	}

	if err := kbd.SendEvent(EV_KEY, code, 1); err != nil {
		return err
	}
	if err := kbd.Sync(); err != nil {
		return err
	}
	if holdFor > 0 {
		time.Sleep(holdFor)
	}
	if err := kbd.SendEvent(EV_KEY, code, 0); err != nil {
		return err
	}
	if err := kbd.Sync(); err != nil {
		return err
	}

	if shift {
		if err := kbd.SendEvent(EV_KEY, KEY_LEFTSHIFT, 0); err != nil {
			return err
		}
		if err := kbd.Sync(); err != nil {
			return err
		}
	}

	if afterDelay > 0 {
		time.Sleep(afterDelay)
	}
	return nil
}

// ── Character → keycode map ───────────────────────────────────────────────────

type keyInfo struct {
	code  uint16
	shift bool
}

var allKeyCodes = map[uint16]struct{}{
	KEY_ESC: {}, KEY_1: {}, KEY_2: {}, KEY_3: {}, KEY_4: {}, KEY_5: {},
	KEY_6: {}, KEY_7: {}, KEY_8: {}, KEY_9: {}, KEY_0: {}, KEY_MINUS: {},
	KEY_EQUAL: {}, KEY_BACKSPACE: {}, KEY_TAB: {}, KEY_Q: {}, KEY_W: {},
	KEY_E: {}, KEY_R: {}, KEY_T: {}, KEY_Y: {}, KEY_U: {}, KEY_I: {},
	KEY_O: {}, KEY_P: {}, KEY_LEFTBRACE: {}, KEY_RIGHTBRACE: {}, KEY_ENTER: {},
	KEY_LEFTCTRL: {}, KEY_A: {}, KEY_S: {}, KEY_D: {}, KEY_F: {}, KEY_G: {},
	KEY_H: {}, KEY_J: {}, KEY_K: {}, KEY_L: {}, KEY_SEMICOLON: {},
	KEY_APOSTROPHE: {}, KEY_GRAVE: {}, KEY_LEFTSHIFT: {}, KEY_BACKSLASH: {},
	KEY_Z: {}, KEY_X: {}, KEY_C: {}, KEY_V: {}, KEY_B: {}, KEY_N: {},
	KEY_M: {}, KEY_COMMA: {}, KEY_DOT: {}, KEY_SLASH: {}, KEY_RIGHTSHIFT: {},
	KEY_LEFTALT: {}, KEY_SPACE: {}, KEY_CAPSLOCK: {}, KEY_F1: {}, KEY_F2: {},
	KEY_F3: {}, KEY_F4: {}, KEY_F5: {}, KEY_F6: {}, KEY_F7: {}, KEY_F8: {},
	KEY_F9: {}, KEY_F10: {}, KEY_F11: {}, KEY_F12: {}, KEY_RIGHTCTRL: {},
	KEY_RIGHTALT: {}, KEY_HOME: {}, KEY_UP: {}, KEY_PAGEUP: {}, KEY_LEFT: {},
	KEY_RIGHT: {}, KEY_END: {}, KEY_DOWN: {}, KEY_PAGEDOWN: {}, KEY_INSERT: {},
	KEY_DELETE: {}, KEY_LEFTMETA: {}, KEY_RIGHTMETA: {},
}

// charKeyMap maps every typeable rune to its keycode + whether Shift is needed.
var charKeyMap = map[rune]keyInfo{
	// Lowercase letters
	'a': {KEY_A, false}, 'b': {KEY_B, false}, 'c': {KEY_C, false},
	'd': {KEY_D, false}, 'e': {KEY_E, false}, 'f': {KEY_F, false},
	'g': {KEY_G, false}, 'h': {KEY_H, false}, 'i': {KEY_I, false},
	'j': {KEY_J, false}, 'k': {KEY_K, false}, 'l': {KEY_L, false},
	'm': {KEY_M, false}, 'n': {KEY_N, false}, 'o': {KEY_O, false},
	'p': {KEY_P, false}, 'q': {KEY_Q, false}, 'r': {KEY_R, false},
	's': {KEY_S, false}, 't': {KEY_T, false}, 'u': {KEY_U, false},
	'v': {KEY_V, false}, 'w': {KEY_W, false}, 'x': {KEY_X, false},
	'y': {KEY_Y, false}, 'z': {KEY_Z, false},

	// Uppercase letters (shift)
	'A': {KEY_A, true}, 'B': {KEY_B, true}, 'C': {KEY_C, true},
	'D': {KEY_D, true}, 'E': {KEY_E, true}, 'F': {KEY_F, true},
	'G': {KEY_G, true}, 'H': {KEY_H, true}, 'I': {KEY_I, true},
	'J': {KEY_J, true}, 'K': {KEY_K, true}, 'L': {KEY_L, true},
	'M': {KEY_M, true}, 'N': {KEY_N, true}, 'O': {KEY_O, true},
	'P': {KEY_P, true}, 'Q': {KEY_Q, true}, 'R': {KEY_R, true},
	'S': {KEY_S, true}, 'T': {KEY_T, true}, 'U': {KEY_U, true},
	'V': {KEY_V, true}, 'W': {KEY_W, true}, 'X': {KEY_X, true},
	'Y': {KEY_Y, true}, 'Z': {KEY_Z, true},

	// Digits
	'1': {KEY_1, false}, '2': {KEY_2, false}, '3': {KEY_3, false},
	'4': {KEY_4, false}, '5': {KEY_5, false}, '6': {KEY_6, false},
	'7': {KEY_7, false}, '8': {KEY_8, false}, '9': {KEY_9, false},
	'0': {KEY_0, false},

	// Shifted digits → symbols
	'!': {KEY_1, true}, '@': {KEY_2, true}, '#': {KEY_3, true},
	'$': {KEY_4, true}, '%': {KEY_5, true}, '^': {KEY_6, true},
	'&': {KEY_7, true}, '*': {KEY_8, true}, '(': {KEY_9, true},
	')': {KEY_0, true},

	// Punctuation (unshifted)
	'-': {KEY_MINUS, false}, '=': {KEY_EQUAL, false},
	'[': {KEY_LEFTBRACE, false}, ']': {KEY_RIGHTBRACE, false},
	'\\': {KEY_BACKSLASH, false}, ';': {KEY_SEMICOLON, false},
	'\'': {KEY_APOSTROPHE, false}, '`': {KEY_GRAVE, false},
	',': {KEY_COMMA, false}, '.': {KEY_DOT, false}, '/': {KEY_SLASH, false},

	// Punctuation (shifted)
	'_': {KEY_MINUS, true}, '+': {KEY_EQUAL, true},
	'{': {KEY_LEFTBRACE, true}, '}': {KEY_RIGHTBRACE, true},
	'|': {KEY_BACKSLASH, true}, ':': {KEY_SEMICOLON, true},
	'"': {KEY_APOSTROPHE, true}, '~': {KEY_GRAVE, true},
	'<': {KEY_COMMA, true}, '>': {KEY_DOT, true}, '?': {KEY_SLASH, true},

	// Whitespace
	' ':  {KEY_SPACE, false},
	'\t': {KEY_TAB, false},
	'\n': {KEY_ENTER, false},
}
