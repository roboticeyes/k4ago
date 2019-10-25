package k4ago

/*
#cgo LDFLAGS: -lk4a
#include <stdlib.h>
#include <string.h>
#include <k4a/k4a.h>
*/
import "C"

import (
	"fmt"
	"log"
	"unsafe"
)

const (
	// Default is the default ID for one connected Kinect
	Default = 0

	snLength = 13
)

// DeviceVersions gives information about device sensor versions
type DeviceVersions struct {
	Rgb    string
	Depth  string
	Audio  string
	Sensor string
}

func (v DeviceVersions) String() string {
	return "Versions [rgb: " + v.Rgb + ", depth: " + v.Depth + ", audio: " + v.Audio + ", sensor: " + v.Sensor + "]"
}

// Device represents one Kinect DK device
type Device struct {
	ID uint32

	// deviceHandle stores the native pointer
	handle C.k4a_device_t
}

// NewDevice creates a new k4ago device
func NewDevice(id uint32) *Device {
	return &Device{
		ID: id,
	}
}

// Open opens the device and establishes the connection. Please make sure to close the device after
// usage!
func (d *Device) Open() error {

	res := C.k4a_device_open(C.uint(d.ID), &d.handle)
	if res != 0 {
		return fmt.Errorf("Cannot open device")
	}
	log.Println("Successfully opened device")
	return nil
}

// Close releases all resources
func (d *Device) Close() {
	C.k4a_device_close(d.handle)
	log.Println("Successfully closed device")
}

// Versions returns the stored sensor versions
func (d *Device) Versions() (DeviceVersions, error) {

	var v C.k4a_hardware_version_t
	res := C.k4a_device_get_version(d.handle, &v)
	if res != 0 {
		return DeviceVersions{}, fmt.Errorf("Cannot get versions: %d", res)
	}

	return DeviceVersions{
		Rgb:    fmt.Sprintf("%d.%d.%d", v.rgb.major, v.rgb.minor, v.rgb.iteration),
		Depth:  fmt.Sprintf("%d.%d.%d", v.depth.major, v.depth.minor, v.depth.iteration),
		Audio:  fmt.Sprintf("%d.%d.%d", v.audio.major, v.audio.minor, v.audio.iteration),
		Sensor: fmt.Sprintf("%d.%d.%d", v.depth_sensor.major, v.depth_sensor.minor, v.depth_sensor.iteration),
	}, nil
}

// SerialNumber returns the serial number of the device
func (d *Device) SerialNumber() (string, error) {

	var sz C.size_t = snLength
	ptr := C.malloc(C.sizeof_char * snLength)
	defer C.free(unsafe.Pointer(ptr))

	res := C.k4a_device_get_serialnum(d.handle, (*C.char)(ptr), &sz)
	if res != 0 {
		return "", fmt.Errorf("Cannot read serial number (sz=%d): error %d", sz, res)
	}
	serial := C.GoString((*C.char)(ptr))
	log.Println("Serial number:", serial)
	return serial, nil
}
