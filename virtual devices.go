package input

import (
	"encoding/binary"
	"os"

	"golang.org/x/sys/unix"
)

// ── Shared virtual device base ────────────────────────────────────────────────

// virtualDev is embedded in VirtualKeyboard and VirtualMouse to share the
// underlying file handle and low-level send/sync logic.

func (d *virtualDev) sendEvent(evType, code uint16, value int32) error {
	ev := InputEvent{Type: evType, Code: code, Value: value}
	return binary.Write(d.dev, binary.LittleEndian, ev)
}

func (d *virtualDev) sync() error {
	return d.sendEvent(EV_SYN, 0, 0)
}

// Close releases the uinput device.
func (d *virtualDev) Close() error {
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
