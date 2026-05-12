package input

import (
	"fmt"
	"time"

	"golang.org/x/sys/unix"
)

// ── VirtualKeyboard ───────────────────────────────────────────────────────────

// VirtualKeyboard is a uinput virtual keyboard device.
type VirtualKeyboard struct{ VirtualDev }

// CreateVirtualKeyboard creates a uinput virtual keyboard.
// Pass block=true to open the fd in blocking mode (default: non-blocking).
func CreateVirtualKeyboard(name string, block ...bool) (*VirtualKeyboard, error) {
	blocking := len(block) > 0 && block[0]

	fd, err := openUinput(blocking)
	if err != nil {
		return nil, err
	}

	ifd := int(fd.Fd())
	unix.IoctlSetInt(ifd, UI_SET_EVBIT, EV_KEY)
	for code := range allKeyCodes {
		unix.IoctlSetInt(ifd, UI_SET_KEYBIT, int(code))
	}

	if err := writeUinputSetup(fd, name); err != nil {
		fd.Close()
		return nil, err
	}
	unix.IoctlSetInt(int(fd.Fd()), UI_DEV_CREATE, 0)

	return &VirtualKeyboard{VirtualDev{fd}}, nil
}

// Press sends a sequence of key events. Args can be any mix of RawKey,
// KeyString, InputTiming, and DelayNow in any order.
//
//	kbd.Press(
//	    InputTiming{HoldFor: 10*time.Millisecond, AfterDelay: 20*time.Millisecond},
//	    KeyString("Hello!"),
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
					return fmt.Errorf("press: no keycode mapping for %q", ch)
				}
				if err := kbd.TapKeyMaybeShift(info.code, info.shift, holdFor, afterDelay); err != nil {
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

func (kbd *VirtualKeyboard) TapKey(code uint16, holdFor, afterDelay time.Duration) error {
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
	if afterDelay > 0 {
		time.Sleep(afterDelay)
	}
	return nil
}

func (kbd *VirtualKeyboard) TapKeyMaybeShift(code uint16, shift bool, holdFor, afterDelay time.Duration) error {
	if shift {
		if err := kbd.SendEvent(EV_KEY, KEY_LEFTSHIFT, 1); err != nil {
			return err
		}
		if err := kbd.Sync(); err != nil {
			return err
		}
	}
	if err := kbd.TapKey(code, holdFor, 0); err != nil {
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
