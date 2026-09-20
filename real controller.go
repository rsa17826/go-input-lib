package input

import (
	"fmt"
	"os"
)

// RealController embeds RealDev to inherit standard device reading capabilities[cite: 4].
type RealController struct {
	RealDev
}

// FindAndOpenController resolves a selector string to a path and opens the device[cite: 1].
func FindAndOpenController(id string) (*RealController, error) {
	p, err := FindDevice(id)
	if err != nil {
		return nil, err
	}
	return OpenController(p)
}

func OpenController(path string) (*RealController, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open device: %w", err)
	}
	return &RealController{RealDev{dev: file}}, nil
}
