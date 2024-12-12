package iec61850

// #include <iec61850_common.h>
import "C"

// GetVersionString retrieves the version string of the underlying libIEC61850 library.
func GetVersionString() string {
	value := C.LibIEC61850_getVersionString()
	return C.GoString(value)
}
