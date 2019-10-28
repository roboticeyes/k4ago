package k4ago

/*
#cgo LDFLAGS: -lk4a
#include <k4a/k4a.h>
*/
import "C"

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"unsafe"
)

// Capture encapsulates a Kinect capture
type Capture struct {
	device *Device
	images map[ImageType]Image
}

// NewCapture returns a new capture object
func NewCapture(d *Device) *Capture {
	return &Capture{
		device: d,
		images: make(map[ImageType]Image),
	}
}

// SingleShot makes one capture shot and stores the data.
// It also releases the capture data.
func (c *Capture) SingleShot() error {

	log.Println("Singleshot start")
	timeout := int(1000 / c.device.Fps)
	timeout = 1000
	var capture C.k4a_capture_t
	res := C.k4a_device_get_capture(c.device.GetHandle(), &capture, C.int(timeout))
	defer C.k4a_capture_release(capture)
	log.Println("Singleshot capture done")

	if res != C.K4A_WAIT_RESULT_SUCCEEDED {
		return fmt.Errorf("Cannot get capture: %d", res)
	} else if res == C.K4A_WAIT_RESULT_TIMEOUT {
		return fmt.Errorf("Running into timeout")
	}

	log.Println("Singleshot get color")
	// get color
	color := C.k4a_capture_get_color_image(capture)
	if color == nil {
		fmt.Println("No color image captured")
	}
	log.Println("Singleshot get depth")

	// get depth
	depth := C.k4a_capture_get_depth_image(capture)
	if depth == nil {
		fmt.Println("No depth image captured")
	}

	log.Println("Singleshot read data")
	c.readImageBuffer(color, ColorImage)
	c.readImageBuffer(depth, DepthImage)
	log.Println("Singleshot read data done")

	log.Println(c.images[ColorImage].Width, c.images[ColorImage].Height)
	log.Println(c.images[DepthImage].Width, c.images[DepthImage].Height)

	return nil
}

// ColorImage converts and returns the RGBA data into a Go image.
// If no image could be found, nil is returned
func (c *Capture) ColorImage() *image.RGBA {

	img, ok := c.images[ColorImage]
	if !ok {
		log.Println("No color image found")
		return nil
	}
	raw := img.Raw
	i := 0
	col := image.NewRGBA(image.Rect(0, 0, img.Width, img.Height))
	for y := 0; y < img.Height; y++ {
		for x := 0; x < img.Width; x++ {
			col.Pix[i] = raw[i+2]
			col.Pix[i+1] = raw[i+1]
			col.Pix[i+2] = raw[i]
			col.Pix[i+3] = raw[i+3]
			i += 4
		}
	}
	return col
}

// DepthImage converts and returns the depth value as a gray16 image.
// If no image could be found, nil is returned
func (c *Capture) DepthImage() *image.Gray16 {

	img, ok := c.images[DepthImage]
	if !ok {
		log.Println("No depth image found")
		return nil
	}
	raw := img.Raw
	depth := image.NewGray16(image.Rect(0, 0, img.Width, img.Height))

	i := 0
	for y := 0; y < img.Height; y++ {
		for x := 0; x < img.Width; x++ {
			val := uint16(raw[i])<<8 | uint16(raw[i+1])
			depth.SetGray16(x, y, color.Gray16{Y: val})
			i += 2
		}
	}
	return depth
}

func (c *Capture) readImageBuffer(input C.k4a_image_t, imageType ImageType) {

	ptr := C.k4a_image_get_buffer(input)
	if ptr == nil {
		log.Fatal("Cannot get pointer")
	}
	sz := C.k4a_image_get_size(input)

	c.images[imageType] = Image{
		Type:   imageType,
		Width:  int(C.k4a_image_get_width_pixels(input)),
		Height: int(C.k4a_image_get_height_pixels(input)),
		Raw:    C.GoBytes(unsafe.Pointer(ptr), C.int(sz)),
	}
}
