package iec61850

// #include <stdint.h>
// #include <stdbool.h>
// #include <stdlib.h>
// #include <iec61850_client.h>
// #include <mms_type_spec.h>
//
// // Callback bridge functions implemented in C (client_async_bridge.c)
// extern void nameListCallbackBridge(uint32_t invokeId, void* parameter, IedClientError err, LinkedList nameList, bool moreFollows);
// extern void varSpecCallbackBridge(uint32_t invokeId, void* parameter, IedClientError err, MmsVariableSpecification* spec);
import "C"

import (
	"runtime/cgo"
	"unsafe"
)

// NameListHandler is invoked for asynchronous name-list responses.
// names contains the page of names received with this callback.
// If moreFollows is true, the server indicates that more elements are available
// and the caller can issue another async call with continueAfter set to the last name.
// When moreFollows is false, this request is completed.
type NameListHandler func(invokeID uint32, names []string, moreFollows bool, err error)

// VarSpecHandler is invoked for asynchronous variable specification responses.
type VarSpecHandler func(invokeID uint32, spec *MmsVariableSpec, err error)

// internal contexts stored as cgo.Handle to survive C roundtrip
type nameListCtx struct {
	handler NameListHandler
}

type varSpecCtx struct {
	handler VarSpecHandler
}

//export nameListCallbackFunctionBridge
func nameListCallbackFunctionBridge(invokeId C.uint32_t, parameter unsafe.Pointer, err C.IedClientError, nameList C.LinkedList, moreFollows C.bool) {
	// Recover handler from parameter handle
	var h cgo.Handle
	if parameter != nil {
		h = cgo.Handle(uintptr(parameter))
	}

	var handler NameListHandler
	if h != 0 {
		if ctx, ok := h.Value().(nameListCtx); ok && ctx.handler != nil {
			handler = ctx.handler
		}
	}

	// Convert error
	goErr := GetIedClientError(err)

	// Convert LinkedList -> []string
	names := make([]string, 0)
	if nameList != nil {
		it := nameList.next
		for it != nil {
			names = append(names, C2GoStr((*C.char)(it.data)))
			it = it.next
		}
		// free C list after conversion
		C.LinkedList_destroy(nameList)
	}

	// Invoke handler if present
	if handler != nil {
		handler(uint32(invokeId), names, bool(moreFollows), goErr)
	}

	// Release handle when finished (no more pages expected or on error)
	if h != 0 && (!bool(moreFollows) || goErr != nil) {
		h.Delete()
	}
}

//export varSpecCallbackFunctionBridge
func varSpecCallbackFunctionBridge(invokeId C.uint32_t, parameter unsafe.Pointer, err C.IedClientError, spec *C.MmsVariableSpecification) {
	var h cgo.Handle
	if parameter != nil {
		h = cgo.Handle(uintptr(parameter))
	}
	var handler VarSpecHandler
	if h != 0 {
		if ctx, ok := h.Value().(varSpecCtx); ok && ctx.handler != nil {
			handler = ctx.handler
		}
	}

	goErr := GetIedClientError(err)

	var goSpec *MmsVariableSpec
	if spec != nil {
		goSpec = cToGoVarSpecStandalone(spec)
		C.MmsVariableSpecification_destroy(spec)
	}

	if handler != nil {
		handler(uint32(invokeId), goSpec, goErr)
	}

	if h != 0 {
		h.Delete()
	}
}

// helper: convert variable spec without requiring a Client receiver (re-using existing logic)
func cToGoVarSpecStandalone(spec *C.MmsVariableSpecification) *MmsVariableSpec {
	if spec == nil {
		return nil
	}
	goSpec := &MmsVariableSpec{
		Type: MmsType(C.MmsVariableSpecification_getType(spec)),
		Name: C2GoStr(C.MmsVariableSpecification_getName(spec)),
	}
	switch goSpec.Type {
	case Array:
		count := int(C.MmsVariableSpecification_getSize(spec))
		elemSpec := C.MmsVariableSpecification_getArrayElementSpecification(spec)
		goSpec.Array = &MmsArraySpec{
			ElementCount: count,
			Element:      cToGoVarSpecStandalone(elemSpec),
		}
	case Structure:
		count := int(C.MmsVariableSpecification_getSize(spec))
		elements := make([]MmsVariableSpec, 0, count)
		for i := 0; i < count; i++ {
			child := C.MmsVariableSpecification_getChildSpecificationByIndex(spec, C.int(i))
			if child != nil {
				if gs := cToGoVarSpecStandalone(child); gs != nil {
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
	}
	return goSpec
}

// Now implement the async initiation methods on Client

// GetServerDirectoryAsync starts an asynchronous request for server directory (LD names).
// continueAfter: empty for first page, or the last received element for continuation calls.
// handler: will be invoked on response or timeout.
// Returns the invoke ID.
func (c *Client) GetServerDirectoryAsync(continueAfter string, handler NameListHandler) (uint32, error) {
	var clientError C.IedClientError
	var cContinue *C.char
	if continueAfter != "" {
		cContinue = Go2CStr(continueAfter)
		defer C.free(unsafe.Pointer(cContinue))
	}

	// store handler in handle so we avoid race before registration
	h := cgo.NewHandle(nameListCtx{handler: handler})
	invokeId := C.IedConnection_getServerDirectoryAsync(c.conn, &clientError, cContinue, nil, (C.IedConnection_GetNameListHandler)(C.nameListCallbackBridge), unsafe.Pointer(uintptr(h)))
	if err := GetIedClientError(clientError); err != nil {
		h.Delete()
		return 0, err
	}
	return uint32(invokeId), nil
}

// GetLogicalDeviceVariablesAsync starts an async request for MMS variable names of a logical device.
func (c *Client) GetLogicalDeviceVariablesAsync(ldName, continueAfter string, handler NameListHandler) (uint32, error) {
	var clientError C.IedClientError
	cLd := Go2CStr(ldName)
	defer C.free(unsafe.Pointer(cLd))
	var cContinue *C.char
	if continueAfter != "" {
		cContinue = Go2CStr(continueAfter)
		defer C.free(unsafe.Pointer(cContinue))
	}
	h := cgo.NewHandle(nameListCtx{handler: handler})
	invokeId := C.IedConnection_getLogicalDeviceVariablesAsync(c.conn, &clientError, cLd, cContinue, nil, (C.IedConnection_GetNameListHandler)(C.nameListCallbackBridge), unsafe.Pointer(uintptr(h)))
	if err := GetIedClientError(clientError); err != nil {
		h.Delete()
		return 0, err
	}
	return uint32(invokeId), nil
}

// GetLogicalDeviceDataSetsAsync starts an async request for dataset names of a logical device.
func (c *Client) GetLogicalDeviceDataSetsAsync(ldName, continueAfter string, handler NameListHandler) (uint32, error) {
	var clientError C.IedClientError
	cLd := Go2CStr(ldName)
	defer C.free(unsafe.Pointer(cLd))
	var cContinue *C.char
	if continueAfter != "" {
		cContinue = Go2CStr(continueAfter)
		defer C.free(unsafe.Pointer(cContinue))
	}
	h := cgo.NewHandle(nameListCtx{handler: handler})
	invokeId := C.IedConnection_getLogicalDeviceDataSetsAsync(c.conn, &clientError, cLd, cContinue, nil, (C.IedConnection_GetNameListHandler)(C.nameListCallbackBridge), unsafe.Pointer(uintptr(h)))
	if err := GetIedClientError(clientError); err != nil {
		h.Delete()
		return 0, err
	}
	return uint32(invokeId), nil
}

// GetVariableSpecificationAsync starts an async request to get variable specification.
func (c *Client) GetVariableSpecificationAsync(dataAttributeReference string, fc FC, handler VarSpecHandler) (uint32, error) {
	var clientError C.IedClientError
	cRef := Go2CStr(dataAttributeReference)
	defer C.free(unsafe.Pointer(cRef))

	h := cgo.NewHandle(varSpecCtx{handler: handler})
	invokeId := C.IedConnection_getVariableSpecificationAsync(c.conn, &clientError, cRef, C.FunctionalConstraint(fc), (C.IedConnection_GetVariableSpecificationHandler)(C.varSpecCallbackBridge), unsafe.Pointer(uintptr(h)))
	if err := GetIedClientError(clientError); err != nil {
		h.Delete()
		return 0, err
	}
	return uint32(invokeId), nil
}
