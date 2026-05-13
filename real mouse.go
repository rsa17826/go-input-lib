package input

import (
	"encoding/binary"
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

func (k *RealMouse) Grab() error {
	_, _, err := syscall.Syscall(syscall.SYS_IOCTL, k.dev.Fd(), uintptr(EVIOCGRAB), uintptr(1))
	if err != 0 {
		return fmt.Errorf("failed to grab device: %w", err)
	}
	return nil
}

func (k *RealMouse) Ungrab() error {
	_, _, err := syscall.Syscall(syscall.SYS_IOCTL, k.dev.Fd(), uintptr(EVIOCGRAB), uintptr(0))
	if err != 0 {
		return fmt.Errorf("failed to grab device: %w", err)
	}
	return nil
}
func OpenMouse(path string) (*RealMouse, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open device: %w", err)
	}
	return &RealMouse{RealDev{dev: file}}, nil
}

// func OpenKeyboardForWrite(path string) (*RealMouse, error) {
// 	file, err := os.OpenFile(path, os.O_RDWR, 0600)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to open device: %w", err)
// 	}
// 	return &RealMouse{RealDev{dev: file}}, nil
// }

func (k *RealMouse) Close() error {
	return k.dev.Close()
}

// func (k *RealMouse) Write(data []byte) (int, error) {
// 	return k.dev.Write(data)
// }

func (k *RealMouse) ReadNextInput() (InputEvent, error) {
	var ev InputEvent
	err := binary.Read(k.dev, binary.NativeEndian, &ev)
	return ev, err
}
func (k *RealMouse) KeyState(keyCode int) (int, error) {
	// 1. Prepare a buffer to hold the bitmask of all keys
	// (KEY_MAX + 7) / 8 gives us enough bytes to hold one bit per key
	buffer := make([]byte, (KEY_MAX+7)/8)

	// 2. Call ioctl to fill the buffer with current key states
	_, _, err := syscall.Syscall(
		syscall.SYS_IOCTL,
		k.dev.Fd(),
		uintptr(EVIOCGKEY),
		uintptr(unsafe.Pointer(&buffer[0])),
	)

	if err != 0 {
		return 0, fmt.Errorf("failed to get key state: %v", err)
	}

	// 3. Check the specific bit for the requested keyCode
	byteIdx := keyCode / 8
	bitIdx := keyCode % 8

	if byteIdx >= len(buffer) {
		return 0, fmt.Errorf("key code %d out of range", keyCode)
	}

	// Returns 1 if bit is set (Down), 0 if not (Up)
	if (buffer[byteIdx] & (1 << uint(bitIdx))) != 0 {
		return 1, nil
	}
	return 0, nil
}

// func (d *RealMouse) SendEvent(evType, code uint16, value int32) error {
// 	ev := InputEvent{Type: evType, Code: code, Value: value}
// 	return binary.Write(d.dev, binary.LittleEndian, ev)
// }

// func (d *RealMouse) Sync() error {
// 	return d.SendEvent(EV_SYN, 0, 0)
// }

func (k *RealMouse) GetPressedKeys() ([]uint16, error) {
	// Create a buffer large enough for the bitmask (1 bit per key)
	buffer := make([]byte, (KEY_MAX+7)/8)

	// EVIOCGKEY(len) ioctl
	_, _, err := syscall.Syscall(
		syscall.SYS_IOCTL,
		k.dev.Fd(),
		uintptr(EVIOCGKEY), // This is the EVIOCGKEY(len) value for most 64-bit systems
		uintptr(unsafe.Pointer(&buffer[0])),
	)

	if err != 0 {
		return nil, fmt.Errorf("failed to get key state: %v", err)
	}

	var pressed []uint16
	for code := 0; code <= KEY_MAX; code++ {
		byteIdx := code / 8
		bitIdx := code % 8

		if (buffer[byteIdx] & (1 << uint(bitIdx))) != 0 {
			pressed = append(pressed, uint16(code))
		}
	}

	return pressed, nil
}
