package input

import (
	"os"
	"syscall"
	"time"
)

// ── Low-level kernel types ────────────────────────────────────────────────────

// InputEvent matches the input_event struct in linux/input.h.
type InputEvent struct {
	Time  syscall.Timeval
	Type  uint16
	Code  uint16
	Value int32
}

// uinputUserDev is the setup struct written to /dev/uinput before UI_DEV_CREATE.
type uinputUserDev struct {
	Name      [80]byte // UINPUT_MAX_NAME_SIZE
	ID        uint16   // bus type
	Vendor    uint16
	Product   uint16
	Version   uint16
	FFEffects uint32
	AbsMax    [64]int32 // ABS_CNT = 64
	AbsMin    [64]int32
	AbsFuzz   [64]int32
	AbsFlat   [64]int32
}

// ── Constants ─────────────────────────────────────────────────────────────────
const (
	// Event types
	EV_SYN = 0x00
	EV_KEY = 0x01
	EV_REL = 0x02

	// Relative axes
	REL_X      = 0x00
	REL_Y      = 0x01
	REL_HWHEEL = 0x06
	REL_WHEEL  = 0x08

	// Mouse buttons
	BTN_LEFT   = 0x110
	BTN_RIGHT  = 0x111
	BTN_MIDDLE = 0x112

	// uinput ioctls
	UI_SET_EVBIT  = 0x40045564
	UI_SET_KEYBIT = 0x40045565
	UI_SET_RELBIT = 0x40045566
	UI_DEV_CREATE = 0x5501

	// Device ioctls
	EVIOCGRAB  = 0x40044590
	EVIOCGNAME = 0x80ff4506

	// Bus type
	BUS_USB = 0x0003

	// Key codes
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

	EVIOCGKEY = 0x80404518 // ioctl code to get key state
	KEY_MAX   = 0x2ff
)

// ── Argument types (shared by Press and Click) ────────────────────────────────

// PressArg is the sealed interface for VirtualKeyboard.Press arguments.
type PressArg interface{ isPressArg() }

// ClickArg is the sealed interface for VirtualMouse.Click arguments.
type ClickArg interface{ isClickArg() }

// RawKey presses a key directly by its kernel keycode (e.g. KEY_ENTER).
// No shift handling is applied.
type RawKey uint16

func (RawKey) isPressArg() {}

// RawButton clicks a mouse button directly by its kernel code (e.g. BTN_LEFT).
type RawButton uint16

func (RawButton) isClickArg() {}

// KeyString types each character, automatically applying the correct keycode
// and Shift modifier for each rune.
//
//	kbd.Press(KeyString(`Hello, World!`))
type KeyString string

func (KeyString) isPressArg() {}

// InputTiming sets the hold-duration and post-event delay for all subsequent
// events in the same Press or Click call. Zero values mean no delay.
type InputTiming struct {
	HoldFor    time.Duration // how long to hold the key/button before releasing
	AfterDelay time.Duration // pause after releasing
}

func (InputTiming) isPressArg() {}
func (InputTiming) isClickArg() {}

// DelayNow inserts an immediate sleep at this point in the sequence without
// affecting the current InputTiming.
type DelayNow time.Duration

func (DelayNow) isPressArg() {}
func (DelayNow) isClickArg() {}

type VirtualDev struct {
	dev *os.File
}

// ── real devices ───────────────────────────────────────────────────

type RealDev struct {
	dev *os.File
}

type RealKeyboard struct {
	RealDev
}

// ── Character → keycode map ───────────────────────────────────────────────────

type keyInfo struct {
	code  uint16
	shift bool
}

// allKeyCodes is every keycode registered with uinput on keyboard creation.
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

// charKeyMap maps every typeable rune to its keycode and whether Shift is needed.
var charKeyMap = map[rune]keyInfo{
	// Lowercase
	'a': {KEY_A, false}, 'b': {KEY_B, false}, 'c': {KEY_C, false},
	'd': {KEY_D, false}, 'e': {KEY_E, false}, 'f': {KEY_F, false},
	'g': {KEY_G, false}, 'h': {KEY_H, false}, 'i': {KEY_I, false},
	'j': {KEY_J, false}, 'k': {KEY_K, false}, 'l': {KEY_L, false},
	'm': {KEY_M, false}, 'n': {KEY_N, false}, 'o': {KEY_O, false},
	'p': {KEY_P, false}, 'q': {KEY_Q, false}, 'r': {KEY_R, false},
	's': {KEY_S, false}, 't': {KEY_T, false}, 'u': {KEY_U, false},
	'v': {KEY_V, false}, 'w': {KEY_W, false}, 'x': {KEY_X, false},
	'y': {KEY_Y, false}, 'z': {KEY_Z, false},

	// Uppercase
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
	'0': {KEY_0, false}, '1': {KEY_1, false}, '2': {KEY_2, false},
	'3': {KEY_3, false}, '4': {KEY_4, false}, '5': {KEY_5, false},
	'6': {KEY_6, false}, '7': {KEY_7, false}, '8': {KEY_8, false},
	'9': {KEY_9, false},

	// Shifted digits
	')': {KEY_0, true}, '!': {KEY_1, true}, '@': {KEY_2, true},
	'#': {KEY_3, true}, '$': {KEY_4, true}, '%': {KEY_5, true},
	'^': {KEY_6, true}, '&': {KEY_7, true}, '*': {KEY_8, true},
	'(': {KEY_9, true},

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
