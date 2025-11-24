package iec61850

/*
#include <iec61850_client.h>

extern void connectionClosedHandlerBridge(void* parameter, IedConnection connection);
*/
import "C"

import (
	"sync"
	"unsafe"
)

// ConnectionClosedHandler is called by the stack when the connection is lost/closed.
// Note: This uses the deprecated libiec61850 IedConnection_installConnectionClosedHandler API
// because some users/devices rely on it. Consider migrating to state-changed handler in future.
type ConnectionClosedHandler func()

var (
	connectionClosedCallbacksMu sync.RWMutex
	connectionClosedCallbacks   = make(map[int32]ConnectionClosedHandler)
)

//export connectionClosedHandlerBridge
func connectionClosedHandlerBridge(parameter unsafe.Pointer, _ C.IedConnection) {
	cbID := int32(uintptr(parameter))
	connectionClosedCallbacksMu.RLock()
	cb := connectionClosedCallbacks[cbID]
	connectionClosedCallbacksMu.RUnlock()
	if cb != nil {
		cb()
	}
}

// InstallConnectionClosedHandler registers a callback invoked when the connection is closed.
// Only one handler per Client is supported; calling again overwrites the previous one.
func (c *Client) InstallConnectionClosedHandler(handler ConnectionClosedHandler) error {
	if handler == nil {
		return nil
	}

	// allocate callback id
	cbID := callbackIdGen.Add(1)
	c.closedHandlerId = cbID

	// store handler
	connectionClosedCallbacksMu.Lock()
	connectionClosedCallbacks[cbID] = handler
	connectionClosedCallbacksMu.Unlock()

	// pass callback id as void* parameter to C
	cPtr := intToPointerBug58625(cbID)
	C.IedConnection_installConnectionClosedHandler(c.conn, (*[0]byte)(C.connectionClosedHandlerBridge), cPtr)
	return nil
}
