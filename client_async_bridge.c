#include <stdint.h>
#include <stdbool.h>
#include <iec61850_client.h>

/*
 * Bridge C callbacks from libiec61850 to Go (cgo //export functions).
 * We use small C wrappers so we can pass function pointers to the lib
 * and still call into Go code safely.
 */

/* extern Go functions implemented in client_async.go */
extern void nameListCallbackFunctionBridge(uint32_t invokeId, void* parameter, IedClientError err, LinkedList nameList, bool moreFollows);
extern void varSpecCallbackFunctionBridge(uint32_t invokeId, void* parameter, IedClientError err, MmsVariableSpecification* spec);

void nameListCallbackBridge(uint32_t invokeId, void* parameter, IedClientError err, LinkedList nameList, bool moreFollows) {
    nameListCallbackFunctionBridge(invokeId, parameter, err, nameList, moreFollows);
}

void varSpecCallbackBridge(uint32_t invokeId, void* parameter, IedClientError err, MmsVariableSpecification* spec) {
    varSpecCallbackFunctionBridge(invokeId, parameter, err, spec);
}
