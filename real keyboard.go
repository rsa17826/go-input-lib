package input

import (
	"encoding/binary"
	"fmt"
	"os"
	"syscall"
)

func (k *RealKeyboard) Grab() error {
	_, _, err := syscall.Syscall(SYS_IOCTL, k.dev.Fd(), uintptr(EVIOCGRAB), uintptr(1))
	if err != 0 {
		return fmt.Errorf("failed to grab device: %w", err)
	}
	return nil
}

func (k *RealKeyboard) Ungrab() error {
	_, _, err := syscall.Syscall(SYS_IOCTL, k.dev.Fd(), uintptr(EVIOCGRAB), uintptr(0))
	if err != 0 {
		return fmt.Errorf("failed to grab device: %w", err)
	}
	return nil
}
func OpenKeyboard(path string) (*RealKeyboard, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open device: %w", err)
	}
	return &RealKeyboard{RealDev{dev: file}}, nil
}

func (k *RealKeyboard) Close() error {
	return k.dev.Close()
}

func (k *RealKeyboard) Write(data []byte) (int, error) {
	return k.dev.Write(data)
}

func (k *RealKeyboard) ReadNextInput() (InputEvent, error) {
	var ev InputEvent
	err := binary.Read(k.dev, binary.NativeEndian, &ev)
	return ev, err
}
