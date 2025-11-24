package iec61850

// #include <iec61850_client.h>
import "C"
import (
	"fmt"
	"strings"
)

// String implements fmt.Stringer for MmsValue.
// It prints a human-readable representation for all MMS data types.
// For scalar types that are supported by libiec61850 constructors in this package,
// it creates a temporary C.MmsValue and reads back the value using the C accessors
// (from mms_value.h). For composite types (Array/Structure) it formats recursively.
func (v MmsValue) String() string {
	var b strings.Builder
	writeMmsValue(&b, v, 0)
	return strings.TrimRight(b.String(), "\n")
}

func writeMmsValue(b *strings.Builder, v MmsValue, level int) {
	//indent := strings.Repeat("  ", level)
	switch v.Type {
	case Array, Structure:
		// Expect Value to be []*MmsValue
		if v.Type == Structure {
			b.WriteString("{")
		} else {
			b.WriteString("[")
		}
		if children, ok := v.Value.([]*MmsValue); ok {
			for _, child := range children {
				if child == nil {
					b.WriteString(strings.Repeat("  ", level+1) + "<nil>")
					continue
				}
				writeMmsValue(b, *child, level+1)
			}
		}
		if v.Type == Structure {
			b.WriteString("}")
		} else {
			b.WriteString("]")
		}
	case Boolean, String, VisibleString, Float, Int8, Int16, Int32, Int64, Uint8, Uint16, Uint32:
		b.WriteString(fmt.Sprintf("%s(%s)", mmsTypeName(v.Type), scalarToStringViaC(v)))
	case Integer, Unsigned:
		// Generic integer families (rare in this package since we refine sizes)
		b.WriteString(fmt.Sprintf("%s(%v)", mmsTypeName(v.Type), v.Value))
	case BitString:
		b.WriteString(fmt.Sprintf("BitString(0b%b)", v.Value))
	case OctetString:
		if bs, ok := v.Value.([]byte); ok {
			b.WriteString(fmt.Sprintf("OctetString(% X)", bs))
		} else {
			b.WriteString(fmt.Sprintf("OctetString(%v)", v.Value))
		}
	case GeneralizedTime:
		b.WriteString(fmt.Sprintf("GeneralizedTime(%v)", v.Value))
	case BinaryTime:
		b.WriteString(fmt.Sprintf("BinaryTime(utcMs=%v)", v.Value))
	case Bcd:
		b.WriteString(fmt.Sprintf("BCD(%v)", v.Value))
	case ObjId:
		b.WriteString(fmt.Sprintf("ObjId(%v)", v.Value))
	case UTCTime:
		b.WriteString(fmt.Sprintf("UTCTime(unix=%v)", v.Value))
	case DataAccessError:
		b.WriteString(fmt.Sprintf("DataAccessError(%v)", v.Value))
	default:
		b.WriteString(fmt.Sprintf("UnknownType(%d:%v)", v.Type, v.Value))
	}
}

// scalarToStringViaC tries to use libiec61850 C accessors to render a scalar value.
// Falls back to Go formatting if construction is not supported.
func scalarToStringViaC(v MmsValue) string {
	// Attempt to create a temporary C.MmsValue using helpers from mms.go
	makeAndRead := func(t MmsType, value interface{}) (string, bool) {
		m, err := toMmsValue(t, value)
		if err != nil || m == nil {
			return "", false
		}
		defer C.MmsValue_delete(m)
		switch t {
		case Boolean:
			return fmt.Sprintf("%t", bool(C.MmsValue_getBoolean(m))), true
		case String, VisibleString:
			return C.GoString(C.MmsValue_toString(m)), true
		case Float:
			return fmt.Sprintf("%g", float32(C.MmsValue_toFloat(m))), true
		case Int8, Int16, Int32, Int64:
			return fmt.Sprintf("%d", int64(C.MmsValue_toInt64(m))), true
		case Uint8, Uint16, Uint32:
			return fmt.Sprintf("%d", uint32(C.MmsValue_toUint32(m))), true
		default:
			return "", false
		}
	}

	if s, ok := makeAndRead(v.Type, v.Value); ok {
		return s
	}
	// Fallback: Go-native formatting
	return fmt.Sprintf("%v", v.Value)
}

func mmsTypeName(t MmsType) string {
	switch t {
	case Array:
		return "Array"
	case Structure:
		return "Structure"
	case Boolean:
		return "Boolean"
	case BitString:
		return "BitString"
	case Integer:
		return "Integer"
	case Unsigned:
		return "Unsigned"
	case Float:
		return "Float"
	case OctetString:
		return "OctetString"
	case VisibleString:
		return "VisibleString"
	case GeneralizedTime:
		return "GeneralizedTime"
	case BinaryTime:
		return "BinaryTime"
	case Bcd:
		return "Bcd"
	case ObjId:
		return "ObjId"
	case String:
		return "String"
	case UTCTime:
		return "UTCTime"
	case DataAccessError:
		return "DataAccessError"
	case Int8:
		return "Int8"
	case Int16:
		return "Int16"
	case Int32:
		return "Int32"
	case Int64:
		return "Int64"
	case Uint8:
		return "Uint8"
	case Uint16:
		return "Uint16"
	case Uint32:
		return "Uint32"
	default:
		return fmt.Sprintf("MmsType(%d)", int(t))
	}
}

func (mt MmsType) String() string {
	return mmsTypeName(mt)
}
