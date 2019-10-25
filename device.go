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
	"time"
	"unsafe"
)

const (
	// Default is the default ID for one connected Kinect
	Default = 0

	snLength           = 13
	maxCaptureAttempts = 10
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

// DeviceConfig configures the device parameters
type DeviceConfig struct {
	Fps             int
	DepthMode       int
	ColorFormat     int
	ColorResolution int
	SyncDepthAndRgb bool
}

// Device represents one Kinect DK device
type Device struct {
	ID  uint32 // device ID
	Fps uint32 // actual FPS

	handle      C.k4a_device_t               // stores the native pointer
	config      C.k4a_device_configuration_t // stores the device config parameters
	calibration C.k4a_calibration_t          // stores the calibration
}

// NewDevice creates a new k4ago device
func NewDevice(id uint32) *Device {
	return &Device{
		ID:     id,
		config: getDefaultConfig(),
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

	err := d.UpdateConfig(DeviceConfig{
		Fps:             C.K4A_FRAMES_PER_SECOND_15,
		DepthMode:       C.K4A_DEPTH_MODE_WFOV_UNBINNED, // C.K4A_DEPTH_MODE_NFOV_UNBINNED
		ColorFormat:     C.K4A_IMAGE_FORMAT_COLOR_BGRA32,
		ColorResolution: C.K4A_COLOR_RESOLUTION_720P,
		SyncDepthAndRgb: true,
	})
	// if calibration cannot be retrieved, close the device
	if err != nil {
		d.Close()
		return err
	}
	return nil
}

// Start starts the camera stream
func (d *Device) Start() error {

	res := C.k4a_device_start_cameras(d.handle, &d.config)
	if res != 0 {
		return fmt.Errorf("Cannot start camera: %d", res)
	}

	// camera requires some time to stabilize itself
	var capture C.k4a_capture_t
	attempts := 0
	for {
		waitRes := C.k4a_device_get_capture(d.handle, &capture, 100)
		if waitRes == C.K4A_WAIT_RESULT_SUCCEEDED {
			log.Println("Yeah, device is now ready ...")
			C.k4a_capture_release(capture)
			break
		}
		if attempts > maxCaptureAttempts {
			return fmt.Errorf("Capture timed out")
		}
		attempts++
		time.Sleep(60 * time.Millisecond)
	}
	return nil
}

// GetHandle returns the native handle
func (d *Device) GetHandle() C.k4a_device_t {
	return d.handle
}

// Stop stops the camera stream
func (d *Device) Stop() {
	C.k4a_device_stop_cameras(d.handle)
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
	return serial, nil
}

// UpdateConfig sets the configuration for the Kinect sensor and reads the new calibration data
func (d *Device) UpdateConfig(config DeviceConfig) error {
	d.config.camera_fps = (C.k4a_fps_t)(config.Fps)
	d.config.depth_mode = (C.k4a_depth_mode_t)(config.DepthMode)
	d.config.color_format = (C.k4a_image_format_t)(config.ColorFormat)
	d.config.color_resolution = (C.k4a_color_resolution_t)(config.ColorResolution)
	d.config.synchronized_images_only = (C.bool)(config.SyncDepthAndRgb)

	d.Fps = convertConfigFpsToRealFps(config.Fps)
	return d.getCalibration()
}

// Get the calibration information for the given config
func (d *Device) getCalibration() error {

	res := C.k4a_device_get_calibration(d.handle, d.config.depth_mode, d.config.color_resolution, &d.calibration)
	if res != 0 {
		return fmt.Errorf("Cannot read calibration data: %d", res)
	}
	return nil
}

func getDefaultConfig() C.k4a_device_configuration_t {
	// those are the defaults from the k4a library
	var config C.k4a_device_configuration_t
	config.color_format = C.K4A_IMAGE_FORMAT_COLOR_MJPG
	config.color_resolution = C.K4A_COLOR_RESOLUTION_OFF
	config.depth_mode = C.K4A_DEPTH_MODE_OFF
	config.camera_fps = C.K4A_FRAMES_PER_SECOND_30
	config.synchronized_images_only = false
	config.depth_delay_off_color_usec = 0
	config.wired_sync_mode = C.K4A_WIRED_SYNC_MODE_STANDALONE
	config.subordinate_delay_off_master_usec = 0
	config.disable_streaming_indicator = false
	return config
}

func convertConfigFpsToRealFps(fps int) uint32 {

	switch fps {
	case C.K4A_FRAMES_PER_SECOND_5:
		return 5
	case C.K4A_FRAMES_PER_SECOND_15:
		return 15
	case C.K4A_FRAMES_PER_SECOND_30:
		return 30
	}
	return 0
}
