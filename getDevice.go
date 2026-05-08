package main

import (
	"encoding/binary"
	"fmt"
	"os"
	"strings"
)

func getDeviceToUser() {
	// 1. Get all persistent device paths
	files, _ := os.ReadDir("/dev/input/")

	// Channel to receive the ID of the device that was touched
	foundChan := make(chan string)

	fmt.Println("Listening on all devices... Press any key on the target device.")

	for _, f := range files {
		if strings.HasPrefix(f.Name(), "event") {
			path := "/dev/input/" + f.Name()
			println(getDeviceName(path))
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
						id := getPersistentID(p)
						if id != "" {
							foundChan <- "id:" + strings.TrimPrefix(id, "/dev/input/by-id/")
						} else {
							foundChan <- "name:" + getDeviceName(p)
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
}
