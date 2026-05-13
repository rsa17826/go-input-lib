package input

import (
	"fmt"
	"os"
)

func OpenMouse(path string) (*RealMouse, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open device: %w", err)
	}
	return &RealMouse{RealDev{dev: file}}, nil
}
