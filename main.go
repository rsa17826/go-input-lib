package input

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"

	argparse "github.com/rsa17826/go-arg-lib"
)

var getDeviceId bool

func init() {
	argparse.ParseArgs([]argparse.ArgumentData{{Keys: []string{"selectDevice"}, AfterCount: 0, Target: &getDeviceId, Description: "interactive device selection to get device id", VarArgs: false, AllowDupes: false}})
	go func() {
		argparse.EnsureParsed() // called from goroutine, not goroutine 1 — waitForInits resolves
		println(getDeviceId, "getDeviceId")
		if getDeviceId {
			GetDeviceToUser()
		}
	}()
}

func GetDeviceToUser() string {
	// 1. Get all persistent device paths
	files, _ := os.ReadDir("/dev/input/")

	// Channel to receive the ID of the device that was touched
	foundChan := make(chan string)

	fmt.Println("Listening on all devices... Press any key on the target device.")

	for _, f := range files {
		if strings.HasPrefix(f.Name(), "event") {
			path := "/dev/input/" + f.Name()
			println(GetDeviceName(path))
			go func(p string) {
				f, err := os.Open(p)
				if err != nil {
					return
				}
				defer f.Close()

				var ev InputEvent
				for {
					err := binary.Read(f, binary.LittleEndian, &ev)
					if err != nil {
						return
					}

					// Type 1 = EV_KEY, Value 1 = Key Down
					if ev.Type == 1 && ev.Value == 1 {
						id := GetPersistentID(p)
						if id != "" {
							foundChan <- "id:" + strings.TrimPrefix(id, "/dev/input/by-id/")
						} else {
							foundChan <- "name:" + GetDeviceName(p)
						}
						return
					}
				}
			}(path)
		}
	}

	// Wait for the first device to send a keypress
	winningID := <-foundChan
	fmt.Printf("\nTarget Device Identified!\n")
	fmt.Printf("Persistent ID: \"%s\"\n", winningID)
	fmt.Println("Use this path in your code to ensure you get the same device every time.")
	return winningID
}
func GetPersistentID(eventPath string) string {
	// eventPath is like "/dev/input/event4"
	absPath, _ := filepath.Abs(eventPath)

	matches, _ := filepath.Glob("/dev/input/by-id/*")
	for _, idPath := range matches {
		evalPath, _ := filepath.EvalSymlinks(idPath)
		if evalPath == absPath {
			return idPath // Found the persistent ID
		}
	}
	return "" // No persistent ID found (likely a virtual device)
}

func GetDeviceName(path string) string {
	f, err := os.Open(path)
	if err != nil {
		return "Unknown (Permission Denied)"
	}
	defer f.Close()

	// Create a buffer to hold the name (up to 256 chars)
	name := make([]byte, 256)

	// Perform the ioctl syscall
	_, _, errno := syscall.Syscall(
		syscall.SYS_IOCTL,
		f.Fd(),
		uintptr(EVIOCGNAME),
		uintptr(unsafe.Pointer(&name[0])),
	)

	if errno != 0 {
		return "Unknown Device"
	}

	// Trim the null characters from the buffer
	return string(bytes.Trim(name, "\x00"))
}
