package iec61850

// #include "iec61850_common.h"
import "C"

// String implements fmt.Stringer for FC. It returns the short IEC 61850
// abbreviation like "ST", "MX", etc.
func (f FC) String() string {
	return C.GoString(C.FunctionalConstraint_toString(C.FunctionalConstraint(f)))
}
