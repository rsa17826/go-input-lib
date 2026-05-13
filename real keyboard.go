package input

import (
	"fmt"
	"os"
)

func FindAndOpenKeyboard(id string) (*RealKeyboard, error) {
	p, err := FindDevice(id)
	if err != nil {
		return nil, err
	}
	kbd, err := OpenKeyboard(p)
	if err != nil {
		return nil, err
	}
	return kbd, nil
}
func OpenKeyboard(path string) (*RealKeyboard, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open device: %w", err)
	}
	return &RealKeyboard{RealDev{dev: file}}, nil
}
