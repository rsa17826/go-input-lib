package input

import (
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
