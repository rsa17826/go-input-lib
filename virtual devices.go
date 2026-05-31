package input

import (
	"encoding/binary"
	"os"
	"time"

	"golang.org/x/sys/unix"
)

// ── Shared virtual device base ────────────────────────────────────────────────

// virtualDev is embedded in VirtualKeyboard and VirtualMouse to share the
// underlying file handle and low-level send/sync logic.

func (d *VirtualDev) SendEvent(evType, code uint16, value int32) error {
	ev := InputEvent{Type: evType, Code: code, Value: value}
	return binary.Write(d.dev, binary.LittleEndian, ev)
}

func (d *VirtualDev) Sync() error {
	return d.SendEvent(EV_SYN, 0, 0)
}

// Close releases the uinput device.
func (d *VirtualDev) Close() error {
	return d.dev.Close()
}

func openUinput(blocking bool) (*os.File, error) {
	flags := os.O_WRONLY
	if !blocking {
		flags |= unix.O_NONBLOCK
	}
	return os.OpenFile("/dev/uinput", flags, 0660)
}

func writeUinputSetup(fd *os.File, name string) error {
	var setup uinputUserDev
	copy(setup.Name[:], name)
	setup.ID = BUS_USB
	return binary.Write(fd, binary.LittleEndian, setup)
}

// TapButton sends a button-down, optional hold, button-up, optional delay.
// Used by both VirtualMouse and VirtualAbsMouse.
func (d *VirtualDev) TapButton(code uint16, holdFor, afterDelay time.Duration) error {
	if err := d.SendEvent(EV_KEY, code, 1); err != nil {
		return err
	}
	if err := d.Sync(); err != nil {
		return err
	}
	if holdFor > 0 {
		time.Sleep(holdFor)
	}
	if err := d.SendEvent(EV_KEY, code, 0); err != nil {
		return err
	}
	if err := d.Sync(); err != nil {
		return err
	}
	if afterDelay > 0 {
		time.Sleep(afterDelay)
	}
	return nil
}
