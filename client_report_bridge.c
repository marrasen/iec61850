#include <stdio.h>
#include <iec61850_client.h>

extern void reportCallbackFunctionBridge(void* parameter, ClientReport report);

void reportCallbackLogging(void* parameter, ClientReport report) {
    printf("[bridge] report from %s rptId=%s\n", ClientReport_getRcbReference(report), ClientReport_getRptId(report));
    reportCallbackFunctionBridge(parameter, report);
}
