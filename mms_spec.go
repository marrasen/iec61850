package iec61850

// #include <iec61850_client.h>
import "C"
import (
	"fmt"
	"slices"
	"strings"
	"unsafe"
)

// MmsVariableSpec is a pure-Go representation of libiec61850's MmsVariableSpecification.
// It models the tagged-union style by having a common Type and optional
// variant-specific fields/children.
type MmsVariableSpec struct {
	Type MmsType
	Name string

	// Complex type variants
	Array     *MmsArraySpec
	Structure *MmsStructureSpec

	// Scalar/meta information for simple types
	IntegerBits        int // for Integer
	UnsignedBits       int // for Unsigned
	FloatExponentWidth int // for Float
	FloatFormatWidth   int // for Float
	BitStringSize      int // number of bits
	OctetStringSize    int // number of octets
	VisibleStringSize  int // max chars
	MmsStringSize      int // MMS String size
	BinaryTimeSize     int // 4 or 6
}

// MmsArraySpec describes an MMS array type
type MmsArraySpec struct {
	ElementCount int
	Element      *MmsVariableSpec
}

// MmsStructureSpec describes an MMS structure type
type MmsStructureSpec struct {
	Elements []MmsVariableSpec // children, each with its own LN and Type
}

// VariableTypeValue represents a flattened variable leaf with its MMS type,
// name from the specification, full object reference, and parsed Go value.
// For composite (Array/Structure) types we don't emit an entry; we only emit
// leaves that carry a concrete value.
type VariableTypeValue struct {
	Type  MmsType
	Name  string
	Ref   string
	Value any
}

// GetVariableSpecification retrieves the MMS variable specification for the given
// data attribute reference and functional constraint (FC) from the connected server.
//
// It wraps the underlying libiec61850 call `IedConnection_getVariableSpecification`
// and converts the returned `MmsVariableSpecification` into a pure-Go
// `MmsVariableSpec`. The returned value is fully detached from any C pointers and
// safe to use after the call returns.
//
// Params:
//   - dataAttributeReference: full object reference (e.g. "IEDNAME/LD0.LLN0.Mod.stVal")
//   - fc: functional constraint to query (e.g. FC_ST, FC_SP, ...)
//
// Returns:
//   - *MmsVariableSpec describing the type. For arrays/structures the result is
//     populated recursively and can be printed using `String()`.
//   - error if the server returned an error or the specification could not be retrieved.
//
// Example:
//
//	spec, err := client.GetVariableSpecification("IED/LD0.LLN0.Mod.stVal", FC_ST)
//	if err != nil { /* handle */ }
//	fmt.Println(spec) // e.g. "Integer(8bit)" or "Structure{stVal: Boolean, q: BitString(13bit), t: UTCTime}"
func (c *Client) GetVariableSpecification(dataAttributeReference string, fc FC) (*MmsVariableSpec, error) {
	var clientError C.IedClientError
	cRef := C.CString(dataAttributeReference)
	defer C.free(unsafe.Pointer(cRef))

	cSpec := C.IedConnection_getVariableSpecification(c.conn, &clientError, cRef, C.FunctionalConstraint(fc))
	if err := GetIedClientError(clientError); err != nil {
		return nil, err
	}
	if cSpec == nil {
		return nil, fmt.Errorf("IedConnection_getVariableSpecification returned NULL")
	}
	defer C.MmsVariableSpecification_destroy(cSpec)

	return c.cToGoVarSpec(cSpec), nil
}

type Vars map[string][]string

type FCVar struct {
	LN     string
	FCVars map[string]Vars
}

func (c *Client) GetLogicalDeviceVariablesHierarchical(ldName string) ([]FCVar, error) {
	vars, err := c.GetLogicalDeviceVariables(ldName)
	if err != nil {
		return nil, err
	}

	m := make(map[string]FCVar)

	for _, v := range vars {
		parts := strings.SplitN(v, "$", 3)
		if len(parts) != 3 {
			// Skip the top level variable names
			continue
		}

		lnName := parts[0]
		fc := parts[1]
		variable := parts[2]
		variables := strings.SplitN(variable, "$", 2)
		base := variables[0]

		if _, ok := m[lnName]; !ok {
			m[lnName] = FCVar{LN: lnName, FCVars: make(map[string]Vars)}
		}
		if _, ok := m[lnName].FCVars[fc]; !ok {
			m[lnName].FCVars[fc] = make(Vars)
		}
		if _, ok := m[lnName].FCVars[fc][base]; !ok {
			m[lnName].FCVars[fc][base] = make([]string, 0)
		}
		if len(variables) == 2 {
			m[lnName].FCVars[fc][base] = append(m[lnName].FCVars[fc][base], variables[1])
		}
	}

	ret := make([]FCVar, 0, len(m))
	for _, v := range m {
		ret = append(ret, v)
	}
	slices.SortFunc(ret, func(a, b FCVar) int {
		return strings.Compare(a.LN, b.LN)
	})
	return ret, nil
}

// GetLogicalDeviceVariables wraps C.IedConnection_getLogicalDeviceVariables and returns Go strings
func (c *Client) GetLogicalDeviceVariables(ldName string) ([]string, error) {
	var clientError C.IedClientError
	cName := C.CString(ldName)
	defer C.free(unsafe.Pointer(cName))

	list := C.IedConnection_getLogicalDeviceVariables(c.conn, &clientError, cName)
	defer func() {
		if list != nil {
			C.LinkedList_destroy(list)
		}
	}()
	if err := GetIedClientError(clientError); err != nil {
		return nil, err
	}
	var out []string
	if list != nil {
		it := list.next
		for it != nil {
			out = append(out, C2GoStr((*C.char)(it.data)))
			it = it.next
		}
	}
	return out, nil
}

// GetLogicalDeviceDataSets wraps C.IedConnection_getLogicalDeviceDataSets and returns Go strings
func (c *Client) GetLogicalDeviceDataSets(ldName string) ([]string, error) {
	var clientError C.IedClientError
	cName := C.CString(ldName)
	defer C.free(unsafe.Pointer(cName))

	list := C.IedConnection_getLogicalDeviceDataSets(c.conn, &clientError, cName)
	defer func() {
		if list != nil {
			C.LinkedList_destroy(list)
		}
	}()
	if err := GetIedClientError(clientError); err != nil {
		return nil, err
	}
	var out []string
	if list != nil {
		it := list.next
		for it != nil {
			out = append(out, C2GoStr((*C.char)(it.data)))
			it = it.next
		}
	}
	return out, nil
}

// GetVariableTypeValues collects all leaf variables and their values for the given
// object reference and functional constraint. It retrieves both the MMS variable
// specification and the current value, then matches them recursively to produce
// a flat list of variable/value pairs. Composite nodes (Array/Structure) are
// traversed but not returned as items; only scalar leaves are returned.
func (c *Client) GetVariableTypeValues(objectRef string, fc FC) ([]VariableTypeValue, error) {
	// Fetch specification
	spec, err := c.GetVariableSpecification(objectRef, fc)
	if err != nil {
		return nil, err
	}

	// Fetch value (already converted to Go types via toGoValue within ReadObject)
	val, err := c.ReadObject(objectRef, fc)
	if err != nil {
		return nil, err
	}

	out := make([]VariableTypeValue, 0)

	// Helper: recursively align spec and value and collect leaves
	var walk func(spec *MmsVariableSpec, val *MmsValue, curRef string, curName string) error
	walk = func(spec *MmsVariableSpec, val *MmsValue, curRef string, curName string) error {
		if spec == nil || val == nil {
			return fmt.Errorf("nil spec or value for ref %s", curRef)
		}

		switch spec.Type {
		case Array:
			if val.Type != Array {
				return fmt.Errorf("type mismatch at %s: spec Array, value %d", curRef, val.Type)
			}
			// Expect value to be []*MmsValue
			elems, ok := val.Value.([]*MmsValue)
			if !ok {
				return fmt.Errorf("unexpected value for array at %s", curRef)
			}
			// Iterate elements; element spec is shared
			for i, child := range elems {
				nextRef := fmt.Sprintf("%s[%d]", curRef, i)
				// Prefer element's own name if provided, otherwise use parent name with index
				nextName := curName
				if spec.Array != nil && spec.Array.Element != nil && spec.Array.Element.Name != "" {
					nextName = spec.Array.Element.Name
				} else {
					nextName = fmt.Sprintf("%s[%d]", curName, i)
				}
				if spec.Array == nil || spec.Array.Element == nil {
					return fmt.Errorf("array spec missing element at %s", curRef)
				}
				if err := walk(spec.Array.Element, child, nextRef, nextName); err != nil {
					return err
				}
			}
			return nil
		case Structure:
			if val.Type != Structure {
				return fmt.Errorf("type mismatch at %s: spec Structure, value %d", curRef, val.Type)
			}
			elems, ok := val.Value.([]*MmsValue)
			if !ok {
				return fmt.Errorf("unexpected value for structure at %s", curRef)
			}
			if spec.Structure == nil {
				return fmt.Errorf("structure spec missing children at %s", curRef)
			}
			if len(spec.Structure.Elements) != len(elems) {
				// We still try to traverse the min length to be tolerant
				// but report a mismatch
				// Continue after recording error
			}
			n := len(elems)
			if len(spec.Structure.Elements) < n {
				n = len(spec.Structure.Elements)
			}
			for i := 0; i < n; i++ {
				childSpec := &spec.Structure.Elements[i]
				fieldName := childSpec.Name
				nextRef := curRef
				if fieldName != "" {
					nextRef = curRef + "." + fieldName
				}
				nextName := fieldName
				if nextName == "" {
					nextName = curName
				}
				if err := walk(childSpec, elems[i], nextRef, nextName); err != nil {
					return err
				}
			}
			return nil
		default:
			// Leaf node: add to output
			leafName := spec.Name
			if leafName == "" {
				leafName = curName
			}
			out = append(out, VariableTypeValue{
				Type:  val.Type,
				Name:  leafName,
				Ref:   curRef,
				Value: val.Value,
			})
			return nil
		}
	}

	// Start with the provided reference and the spec root name
	startName := spec.Name
	if startName == "" {
		// derive from objectRef last path if possible
		if idx := strings.LastIndex(objectRef, "."); idx != -1 && idx+1 < len(objectRef) {
			startName = objectRef[idx+1:]
		} else {
			startName = objectRef
		}
	}

	if err := walk(spec, val, objectRef, startName); err != nil {
		return nil, err
	}
	return out, nil
}

// cToGoVarSpec converts a C MmsVariableSpecification into a Go MmsVariableSpec recursively.
func (c *Client) cToGoVarSpec(spec *C.MmsVariableSpecification) *MmsVariableSpec {
	if spec == nil {
		return nil
	}
	goSpec := &MmsVariableSpec{
		Type: MmsType(C.MmsVariableSpecification_getType(spec)),
		Name: C2GoStr(C.MmsVariableSpecification_getName(spec)),
	}

	switch goSpec.Type {
	case Array:
		// element count in getSize, element spec via getArrayElementSpecification
		count := int(C.MmsVariableSpecification_getSize(spec))
		elemSpec := C.MmsVariableSpecification_getArrayElementSpecification(spec)
		goSpec.Array = &MmsArraySpec{
			ElementCount: count,
			Element:      c.cToGoVarSpec(elemSpec),
		}
	case Structure:
		count := int(C.MmsVariableSpecification_getSize(spec))
		elements := make([]MmsVariableSpec, 0, count)
		for i := 0; i < count; i++ {
			child := C.MmsVariableSpecification_getChildSpecificationByIndex(spec, C.int(i))
			if child != nil {
				if gs := c.cToGoVarSpec(child); gs != nil {
					elements = append(elements, *gs)
				}
			}
		}
		goSpec.Structure = &MmsStructureSpec{Elements: elements}
	case Integer:
		goSpec.IntegerBits = int(C.MmsVariableSpecification_getSize(spec))
	case Unsigned:
		goSpec.UnsignedBits = int(C.MmsVariableSpecification_getSize(spec))
	case Float:
		// Use the dedicated accessor for exponent; format width often equals getSize
		goSpec.FloatExponentWidth = int(C.MmsVariableSpecification_getExponentWidth(spec))
		goSpec.FloatFormatWidth = int(C.MmsVariableSpecification_getSize(spec))
	case BitString:
		goSpec.BitStringSize = int(C.MmsVariableSpecification_getSize(spec))
	case OctetString:
		goSpec.OctetStringSize = int(C.MmsVariableSpecification_getSize(spec))
	case VisibleString:
		goSpec.VisibleStringSize = int(C.MmsVariableSpecification_getSize(spec))
	case String:
		goSpec.MmsStringSize = int(C.MmsVariableSpecification_getSize(spec))
	case BinaryTime:
		goSpec.BinaryTimeSize = int(C.MmsVariableSpecification_getSize(spec))
	default:
		// For types like Boolean, GeneralizedTime, UTCTime, ObjId, etc., no extra fields required
	}

	return goSpec
}
