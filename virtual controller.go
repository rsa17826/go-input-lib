package input

import (
	"encoding/binary"

	"golang.org/x/sys/unix"
)

// VirtualController embeds VirtualDev to share the underlying file handle and send/sync logic[cite: 8].
type VirtualController struct {
	VirtualDev
}

// CreateVirtualController creates a uinput virtual gamepad.
func CreateVirtualController(name string, blocking ...bool) (*VirtualController, error) {
	isBlocking := len(blocking) > 0 && blocking[0]

	// openUinput is called to establish the file descriptor for /dev/uinput[cite: 8].
	fd, err := openUinput(isBlocking)
	if err != nil {
		return nil, err
	}

	ifd := int(fd.Fd())

	// Enable Controller Buttons (EV_KEY)[cite: 7]
	unix.IoctlSetInt(ifd, UI_SET_EVBIT, EV_KEY)
	buttons := []int{
		BTN_SOUTH, BTN_EAST, BTN_NORTH, BTN_WEST,
		BTN_TL, BTN_TR, BTN_TL2, BTN_TR2,
		BTN_SELECT, BTN_START, BTN_MODE,
		BTN_THUMBL, BTN_THUMBR,
	}
	for _, btn := range buttons {
		unix.IoctlSetInt(ifd, UI_SET_KEYBIT, btn)
	}

	// Enable Absolute Axes (EV_ABS) for Joysticks and D-Pad[cite: 7]
	unix.IoctlSetInt(ifd, UI_SET_EVBIT, EV_ABS)
	axes := []int{ABS_X, ABS_Y, ABS_Z, ABS_RX, ABS_RY, ABS_RZ, ABS_HAT0X, ABS_HAT0Y}
	for _, axis := range axes {
		unix.IoctlSetInt(ifd, UI_SET_ABSBIT, axis)
	}

	var setup uinputUserDev
	copy(setup.Name[:], name)
	setup.ID = BUS_USB // Sets the bus type to USB[cite: 7]

	// Define limits and deadzones for absolute axes
	for _, axis := range axes {
		setup.AbsMin[axis] = -32768
		setup.AbsMax[axis] = 32767
		setup.AbsFuzz[axis] = 16
		setup.AbsFlat[axis] = 128
	}

	if err := binary.Write(fd, binary.LittleEndian, setup); err != nil {
		fd.Close()
		return nil, err
	}
	unix.IoctlSetInt(ifd, UI_DEV_CREATE, 0)

	return &VirtualController{VirtualDev: VirtualDev{dev: fd}}, nil
}

// PressButton sends a digital button state (1 for down, 0 for up).
func (c *VirtualController) PressButton(code uint16, value int32) error {
	if err := c.SendEvent(EV_KEY, code, value); err != nil {
		return err
	}
	return c.Sync()
}

// MoveAxis sends an absolute value to a specified axis (e.g., analog sticks).
func (c *VirtualController) MoveAxis(axis uint16, value int32) error {
	if err := c.SendEvent(EV_ABS, axis, value); err != nil {
		return err
	}
	return c.Sync() // Synchronizes the event sequence[cite: 8]
}
