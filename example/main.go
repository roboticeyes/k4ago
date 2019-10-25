package main

import (
	"fmt"
	"github.com/roboticeyes/k4ago"
	"image/png"
	"log"
	"os"
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

	// Start camera for capture mode
	err = d.Start()
	if err != nil {
		log.Println(err)
	}

	capture := k4ago.NewCapture(d)
	capture.SingleShot()

	colorImage := capture.ColorImage()
	colorFile, err := os.Create("color.png")
	if err != nil {
		log.Fatal(err)
	}
	defer colorFile.Close()
	png.Encode(colorFile, colorImage)

	// Stop camera
	d.Stop()
}
