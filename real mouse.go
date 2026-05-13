package input

import (
	"fmt"
	"os"
)

func FindAndOpenMouse(id string) (*RealMouse, error) {
	p, err := FindDevice(id)
	if err != nil {
		return nil, err
	}
	mouse, err := OpenMouse(p)
	if err != nil {
		return nil, err
	}
	return mouse, nil
}
func OpenMouse(path string) (*RealMouse, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open device: %w", err)
	}
	return &RealMouse{RealDev{dev: file}}, nil
}
