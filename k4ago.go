package k4ago

/*
#cgo LDFLAGS: -lk4a
#include <k4a/k4a.h>
*/
import "C"

// AvailableDevices returns the number of devices attached to the computer
func AvailableDevices() uint {
	return uint(C.k4a_device_get_installed_count())
}
