package input

import (
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

	// 1. Register basic capability groups
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

	// Modern Linux uinput ioctl definitions
	const (
		UI_DEV_SETUP = 0x405c5503
		UI_ABS_SETUP = 0x401c5504
	)

	type inputId struct {
		Bustype uint16
		Vendor  uint16
		Product uint16
		Version uint16
	}

	type uinputSetup struct {
		ID           inputId
		Name         [80]byte
		FfEffectsMax uint32
	}

	type inputAbsinfo struct {
		Value      int32
		Minimum    int32
		Maximum    int32
		Fuzz       int32
		Flat       int32
		Resolution int32
	}

	type uinputAbsSetup struct {
		Code    uint16
		_       uint16 // Padding
		Absinfo inputAbsinfo
	}

	// 2. Configure Absolute Ranges using modern UI_ABS_SETUP
	absX := uinputAbsSetup{
		Code:    uint16(ABS_X),
		Absinfo: inputAbsinfo{Minimum: 0, Maximum: cfg.maxX},
	}
	if err := ioctlSetPointer(ifd, UI_ABS_SETUP, unsafe.Pointer(&absX)); err != nil {
		fd.Close()
		return nil, err
	}

	absY := uinputAbsSetup{
		Code:    uint16(ABS_Y),
		Absinfo: inputAbsinfo{Minimum: 0, Maximum: cfg.maxY},
	}
	if err := ioctlSetPointer(ifd, UI_ABS_SETUP, unsafe.Pointer(&absY)); err != nil {
		fd.Close()
		return nil, err
	}

	// 3. Setup device identity using modern UI_DEV_SETUP instead of binary.Write
	var devSetup uinputSetup
	devSetup.ID = inputId{Bustype: BUS_USB}
	copy(devSetup.Name[:79], name)

	if err := ioctlSetPointer(ifd, UI_DEV_SETUP, unsafe.Pointer(&devSetup)); err != nil {
		fd.Close()
		return nil, err
	}

	// 4. Finally create the device node
	if err := unix.IoctlSetInt(ifd, UI_DEV_CREATE, 0); err != nil {
		fd.Close()
		return nil, err
	}

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
