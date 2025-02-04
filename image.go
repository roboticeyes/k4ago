package k4ago

const (
	// ColorImage ...
	ColorImage = iota
	// DepthImage ...
	DepthImage
	// DepthTransformed is the depth image mapped on the RGB image
	DepthTransformed
)

// ImageType for the type
type ImageType int

// Image represents a container for a raw capture image of a type
type Image struct {
	Type   ImageType
	Width  int
	Height int
	Raw    []byte
}
