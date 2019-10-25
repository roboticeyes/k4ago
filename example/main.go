package main

import (
	"fmt"
	"github.com/roboticeyes/k4ago"
)

func countDevices() uint {
	deviceCount := k4ago.AvailableDevices()
	fmt.Printf("Found %d devices\n", deviceCount)
	return deviceCount
}

func main() {
	fmt.Println("k4ago examples")

	if countDevices() == 0 {
		panic("No device found.")
	}
	d := k4ago.NewDevice(k4ago.Default)
	d.Open()
	defer d.Close()

	// Get versions
	versions, err := d.Versions()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(versions)

	// Get serial number
	serial, err := d.SerialNumber()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Serial:", serial)
}
