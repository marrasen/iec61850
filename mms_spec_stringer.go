package iec61850

import (
	"fmt"
	"strings"
)

// String returns a human-readable description of an MMS variable specification.
// It takes the type into account and prints details for Arrays and Structures
// recursively. For scalar types with size parameters, the relevant size fields
// are included when available.
func (s MmsVariableSpec) String() string {
	var b strings.Builder
	writeVarSpec(&b, s)
	return b.String()
}

func writeVarSpec(b *strings.Builder, s MmsVariableSpec) {
	switch s.Type {
	case Array:
		// Array[count] of <elem>
		if s.Array == nil || s.Array.Element == nil {
			b.WriteString("Array[?]")
			return
		}
		fmt.Fprintf(b, "Array[%d] of ", s.Array.ElementCount)
		writeVarSpec(b, *s.Array.Element)
	case Structure:
		// Structure{name: <spec>, ...}
		b.WriteString("Structure{")
		if s.Structure != nil && len(s.Structure.Elements) > 0 {
			for i, el := range s.Structure.Elements {
				if i > 0 {
					b.WriteString(", ")
				}
				// Each element may carry its own name
				if el.Name != "" {
					fmt.Fprintf(b, "%s: ", el.Name)
				}
				writeVarSpec(b, el)
			}
		}
		b.WriteString("}")
	case Integer:
		if s.IntegerBits != 0 {
			fmt.Fprintf(b, "Integer(%dbit)", s.IntegerBits)
		} else {
			b.WriteString("Integer")
		}
	case Unsigned:
		if s.UnsignedBits != 0 {
			fmt.Fprintf(b, "Unsigned(%dbit)", s.UnsignedBits)
		} else {
			b.WriteString("Unsigned")
		}
	case Float:
		if s.FloatFormatWidth != 0 || s.FloatExponentWidth != 0 {
			if s.FloatFormatWidth != 0 && s.FloatExponentWidth != 0 {
				fmt.Fprintf(b, "Float(fmt=%d, exp=%d)", s.FloatFormatWidth, s.FloatExponentWidth)
			} else if s.FloatFormatWidth != 0 {
				fmt.Fprintf(b, "Float(fmt=%d)", s.FloatFormatWidth)
			} else {
				fmt.Fprintf(b, "Float(exp=%d)", s.FloatExponentWidth)
			}
		} else {
			b.WriteString("Float")
		}
	case BitString:
		if s.BitStringSize != 0 {
			fmt.Fprintf(b, "BitString(%dbit)", s.BitStringSize)
		} else {
			b.WriteString("BitString")
		}
	case OctetString:
		if s.OctetStringSize != 0 {
			fmt.Fprintf(b, "OctetString(%d)", s.OctetStringSize)
		} else {
			b.WriteString("OctetString")
		}
	case VisibleString:
		if s.VisibleStringSize != 0 {
			fmt.Fprintf(b, "VisibleString(%d)", s.VisibleStringSize)
		} else {
			b.WriteString("VisibleString")
		}
	case String:
		if s.MmsStringSize != 0 {
			fmt.Fprintf(b, "String(%d)", s.MmsStringSize)
		} else {
			b.WriteString("String")
		}
	case BinaryTime:
		if s.BinaryTimeSize != 0 {
			fmt.Fprintf(b, "BinaryTime(%d)", s.BinaryTimeSize)
		} else {
			b.WriteString("BinaryTime")
		}
	default:
		// For all other scalar/meta types just render the type name.
		b.WriteString(mmsTypeName(s.Type))
	}
}
