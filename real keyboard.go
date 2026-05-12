package input

import (
	"encoding/binary"
	"fmt"
	"os"
)

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
