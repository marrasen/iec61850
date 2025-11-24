package iec61850

// #include <iec61850_client.h>
import "C"
import (
	"fmt"
	"github.com/spf13/cast"
	"unsafe"
)

const (
	ActDA  = "%s/%s.SGCB.ActSG"
	EditDA = "%s/%s.SGCB.EditSG"
	CnfDA  = "%s/%s.SGCB.CnfEdit"
)

type SettingGroup struct {
	NumOfSG int
	ActSG   int
	EditSG  int
	CnfEdit bool
}

// WriteSG 写入SettingGroup
func (c *Client) WriteSG(ld, ln, objectRef string, fc FC, actSG int, value interface{}) error {
	// Set active setting group
	if err := c.WriteObject(fmt.Sprintf(ActDA, ld, ln), SP, actSG); err != nil {
		return fmt.Errorf("WriteSG set ActSG ld=%s ln=%s actSG=%d: %w", ld, ln, actSG, err)
	}

	// Set edit setting group
	if err := c.WriteObject(fmt.Sprintf(EditDA, ld, ln), SP, actSG); err != nil {
		return fmt.Errorf("WriteSG set EditSG ld=%s ln=%s actSG=%d: %w", ld, ln, actSG, err)
	}

	// Change a setting group value
	if err := c.WriteObject(objectRef, fc, value); err != nil {
		return fmt.Errorf("WriteSG write value %q fc=%s: %w", objectRef, fc, err)
	}

	// Confirm new setting group values
	if err := c.WriteObject(fmt.Sprintf(CnfDA, ld, ln), SP, true); err != nil {
		return fmt.Errorf("WriteSG confirm CnfEdit ld=%s ln=%s: %w", ld, ln, err)
	}
	return nil
}

// GetSG 获取SettingGroup
func (c *Client) GetSG(objectRef string) (*SettingGroup, error) {
	var clientError C.IedClientError
	cObjectRef := C.CString(objectRef)
	defer C.free(unsafe.Pointer(cObjectRef))

	// 获取类型
	sgcbVarSpec := C.IedConnection_getVariableSpecification(c.conn, &clientError, cObjectRef, C.FunctionalConstraint(SP))
	if err := GetIedClientError(clientError); err != nil {
		return nil, fmt.Errorf("GetSG get var spec %q: %w", objectRef, err)
	}
	defer C.MmsVariableSpecification_destroy(sgcbVarSpec)

	// ReadObject SGCB
	sgcbVal := C.IedConnection_readObject(c.conn, &clientError, cObjectRef, C.FunctionalConstraint(SP))
	if err := GetIedClientError(clientError); err != nil {
		return nil, fmt.Errorf("GetSG read object %q: %w", objectRef, err)
	}
	//defer C.MmsValue_delete(sgcbVal)

	numOfSGValue, err := c.getSubElementValue(sgcbVal, sgcbVarSpec, "NumOfSG")
	if err != nil {
		return nil, fmt.Errorf("GetSG read NumOfSG from %q: %w", objectRef, err)
	}

	actSGValue, err := c.getSubElementValue(sgcbVal, sgcbVarSpec, "ActSG")
	if err != nil {
		return nil, fmt.Errorf("GetSG read ActSG from %q: %w", objectRef, err)
	}

	editSGValue, err := c.getSubElementValue(sgcbVal, sgcbVarSpec, "EditSG")
	if err != nil {
		return nil, fmt.Errorf("GetSG read EditSG from %q: %w", objectRef, err)
	}

	cnfEditValue, err := c.getSubElementValue(sgcbVal, sgcbVarSpec, "CnfEdit")
	if err != nil {
		return nil, fmt.Errorf("GetSG read CnfEdit from %q: %w", objectRef, err)
	}

	sg := &SettingGroup{
		NumOfSG: cast.ToInt(numOfSGValue),
		ActSG:   cast.ToInt(actSGValue),
		EditSG:  cast.ToInt(editSGValue),
		CnfEdit: cast.ToBool(cnfEditValue),
	}
	return sg, nil
}
