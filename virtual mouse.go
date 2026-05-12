package input

import (
	"time"

	"golang.org/x/sys/unix"
)

// ── VirtualMouse ──────────────────────────────────────────────────────────────

// VirtualMouse is a uinput virtual mouse device.
type VirtualMouse struct{ virtualDev }

// CreateVirtualMouse creates a uinput virtual mouse.
// Pass block=true to open the fd in blocking mode (default: non-blocking).
func CreateVirtualMouse(name string, block ...bool) (*VirtualMouse, error) {
	blocking := len(block) > 0 && block[0]

	fd, err := openUinput(blocking)
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

	if err := writeUinputSetup(fd, name); err != nil {
		fd.Close()
		return nil, err
	}
	unix.IoctlSetInt(int(fd.Fd()), UI_DEV_CREATE, 0)

	return &VirtualMouse{virtualDev{fd}}, nil
}

// Click sends a sequence of button events. Args can be any mix of RawButton,
// InputTiming, and DelayNow in any order.
//
//	mouse.Click(RawButton(BTN_LEFT))
//	mouse.Click(InputTiming{HoldFor: 50*time.Millisecond}, RawButton(BTN_RIGHT))
func (m *VirtualMouse) Click(args ...ClickArg) error {
	var holdFor, afterDelay time.Duration

	for _, arg := range args {
		switch v := arg.(type) {
		case RawButton:
			if err := m.tapButton(uint16(v), holdFor, afterDelay); err != nil {
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

// Move moves the mouse cursor by dx, dy relative to its current position.
func (m *VirtualMouse) Move(dx, dy int32) error {
	if err := m.SendEvent(EV_REL, REL_X, dx); err != nil {
		return err
	}
	if err := m.SendEvent(EV_REL, REL_Y, dy); err != nil {
		return err
	}
	return m.Sync()
}

// Scroll scrolls vertically by clicks (positive = up) and horizontally by hClicks.
func (m *VirtualMouse) Scroll(clicks, hClicks int32) error {
	if clicks != 0 {
		if err := m.SendEvent(EV_REL, REL_WHEEL, clicks); err != nil {
			return err
		}
	}
	if hClicks != 0 {
		if err := m.SendEvent(EV_REL, REL_HWHEEL, hClicks); err != nil {
			return err
		}
	}
	return m.Sync()
}

func (m *VirtualMouse) tapButton(code uint16, holdFor, afterDelay time.Duration) error {
	if err := m.SendEvent(EV_KEY, code, 1); err != nil {
		return err
	}
	if err := m.Sync(); err != nil {
		return err
	}
	if holdFor > 0 {
		time.Sleep(holdFor)
	}
	if err := m.SendEvent(EV_KEY, code, 0); err != nil {
		return err
	}
	if err := m.Sync(); err != nil {
		return err
	}
	if afterDelay > 0 {
		time.Sleep(afterDelay)
	}
	return nil
}
