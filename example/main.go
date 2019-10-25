package main

import (
	"fmt"
	"github.com/roboticeyes/k4ago"
	"log"
	"time"
)

func countDevices() uint {
	deviceCount := k4ago.AvailableDevices()
	fmt.Printf("Found %d devices\n", deviceCount)
	return deviceCount
}

func main() {
	log.Println("k4ago")

	if countDevices() == 0 {
		panic("No device found.")
	}
	d := k4ago.NewDevice(k4ago.Default)
	if err := d.Open(); err != nil {
		panic(err)
	}
	defer d.Close()

	// Get versions
	versions, err := d.Versions()
	if err != nil {
		log.Fatal(err)
	}
	log.Println(versions)

	// Get serial number
	serial, err := d.SerialNumber()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Serial:", serial)

	d.Start()
	time.Sleep(1 * time.Second)
	d.Stop()
}
