package input

import (
	"encoding/binary"
	"time"
	"unsafe"

	"golang.org/x/sys/unix"
)

// ── VirtualMouse ──────────────────────────────────────────────────────────────

// VirtualMouse is a uinput virtual mouse device.
type VirtualMouse struct {
	VirtualDev
	MaxX, MaxY int32
}

// CreateVirtualMouse creates a uinput virtual mouse.
// Pass block=true to open the fd in blocking mode (default: non-blocking).
// MouseOption is a functional option for CreateVirtualMouse.
type MouseOption func(*mouseConfig)

type mouseConfig struct {
	blocking bool
	maxX     int32
	maxY     int32
}

func defaultMouseConfig() mouseConfig {
	return mouseConfig{maxX: 1920, maxY: 1080}
}

// Blocking opens the uinput fd in blocking mode.
func Blocking() MouseOption { return func(c *mouseConfig) { c.blocking = true } }

// WithAbsRange sets the coordinate space for absolute positioning.
func WithAbsRange(maxX, maxY int32) MouseOption {
	return func(c *mouseConfig) { c.maxX = maxX; c.maxY = maxY }
}

// Helper function to handle pointer-based ioctls
func ioctlSetPointer(fd int, req uint, ptr unsafe.Pointer) error {
	_, _, errno := unix.Syscall(
		unix.SYS_IOCTL,
		uintptr(fd),
		uintptr(req),
		uintptr(ptr),
	)
	if errno != 0 {
		return errno
	}
	return nil
}
func CreateVirtualMouse(name string, opts ...MouseOption) (*VirtualMouse, error) {
	cfg := defaultMouseConfig()
	for _, o := range opts {
		o(&cfg)
	}

	fd, err := openUinput(cfg.blocking)
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

	unix.IoctlSetInt(ifd, UI_SET_EVBIT, EV_ABS)
	unix.IoctlSetInt(ifd, UI_SET_ABSBIT, ABS_X)
	unix.IoctlSetInt(ifd, UI_SET_ABSBIT, ABS_Y)

	// UI_ABS_SETUP ioctl value
	const UI_ABS_SETUP = 0x401c5504

	type uinputAbsSetup struct {
		Code uint16
		_    uint16 // Padding
		Val  int32
		Min  int32
		Max  int32
		Fuzz int32
		Flat int32
		Res  int32
	}

	// Setup X Axis limits using our unsafe pointer helper
	absX := uinputAbsSetup{Code: uint16(ABS_X), Min: 0, Max: cfg.maxX}
	if err := ioctlSetPointer(ifd, UI_ABS_SETUP, unsafe.Pointer(&absX)); err != nil {
		fd.Close()
		return nil, err
	}

	// Setup Y Axis limits using our unsafe pointer helper
	absY := uinputAbsSetup{Code: uint16(ABS_Y), Min: 0, Max: cfg.maxY}
	if err := ioctlSetPointer(ifd, UI_ABS_SETUP, unsafe.Pointer(&absY)); err != nil {
		fd.Close()
		return nil, err
	}

	var setup uinputUserDev
	copy(setup.Name[:], name)
	setup.ID = BUS_USB
	setup.AbsMax[ABS_X] = cfg.maxX
	setup.AbsMax[ABS_Y] = cfg.maxY
	if err := binary.Write(fd, binary.LittleEndian, setup); err != nil {
		fd.Close()
		return nil, err
	}
	unix.IoctlSetInt(int(fd.Fd()), UI_DEV_CREATE, 0)

	return &VirtualMouse{VirtualDev: VirtualDev{fd}, MaxX: cfg.maxX, MaxY: cfg.maxY}, nil
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

func (m *VirtualMouse) MoveTo(x, y int32) error {
	x = clamp(x, 0, m.MaxX)
	y = clamp(y, 0, m.MaxY)
	if err := m.SendEvent(EV_ABS, ABS_X, x); err != nil {
		return err
	}
	if err := m.SendEvent(EV_ABS, ABS_Y, y); err != nil {
		return err
	}
	return m.Sync()
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
	return m.TapButton(code, holdFor, afterDelay)
}

func clamp(v, lo, hi int32) int32 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
