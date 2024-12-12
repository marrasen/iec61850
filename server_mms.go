package iec61850

// #include <iec61850_server.h>
import "C"

type MmsError int

type MMSServer struct {
	server C.MmsServer
}

const (
	/* generic error codes */
	MMS_ERROR_NONE                   MmsError = 0
	MMS_ERROR_CONNECTION_REJECTED    MmsError = 1
	MMS_ERROR_CONNECTION_LOST        MmsError = 2
	MMS_ERROR_SERVICE_TIMEOUT        MmsError = 3
	MMS_ERROR_PARSING_RESPONSE       MmsError = 4
	MMS_ERROR_HARDWARE_FAULT         MmsError = 5
	MMS_ERROR_CONCLUDE_REJECTED      MmsError = 6
	MMS_ERROR_INVALID_ARGUMENTS      MmsError = 7
	MMS_ERROR_OUTSTANDING_CALL_LIMIT MmsError = 8
	MMS_ERROR_OTHER                  MmsError = 9

	/* confirmed error PDU codes */
	MMS_ERROR_VMDSTATE_OTHER MmsError = 10

	MMS_ERROR_APPLICATION_REFERENCE_OTHER MmsError = 20

	MMS_ERROR_DEFINITION_OTHER                         MmsError = 30
	MMS_ERROR_DEFINITION_INVALID_ADDRESS               MmsError = 31
	MMS_ERROR_DEFINITION_TYPE_UNSUPPORTED              MmsError = 32
	MMS_ERROR_DEFINITION_TYPE_INCONSISTENT             MmsError = 33
	MMS_ERROR_DEFINITION_OBJECT_UNDEFINED              MmsError = 34
	MMS_ERROR_DEFINITION_OBJECT_EXISTS                 MmsError = 35
	MMS_ERROR_DEFINITION_OBJECT_ATTRIBUTE_INCONSISTENT MmsError = 36

	MMS_ERROR_RESOURCE_OTHER                  MmsError = 40
	MMS_ERROR_RESOURCE_CAPABILITY_UNAVAILABLE MmsError = 41

	MMS_ERROR_SERVICE_OTHER                      MmsError = 50
	MMS_ERROR_SERVICE_OBJECT_CONSTRAINT_CONFLICT MmsError = 55

	MMS_ERROR_SERVICE_PREEMPT_OTHER MmsError = 60

	MMS_ERROR_TIME_RESOLUTION_OTHER MmsError = 70

	MMS_ERROR_ACCESS_OTHER                     MmsError = 80
	MMS_ERROR_ACCESS_OBJECT_NON_EXISTENT       MmsError = 81
	MMS_ERROR_ACCESS_OBJECT_ACCESS_UNSUPPORTED MmsError = 82
	MMS_ERROR_ACCESS_OBJECT_ACCESS_DENIED      MmsError = 83
	MMS_ERROR_ACCESS_OBJECT_INVALIDATED        MmsError = 84
	MMS_ERROR_ACCESS_OBJECT_VALUE_INVALID      MmsError = 85 /* for DataAccessError 11 */
	MMS_ERROR_ACCESS_TEMPORARILY_UNAVAILABLE   MmsError = 86 /* for DataAccessError 2 */

	MMS_ERROR_FILE_OTHER                           MmsError = 90
	MMS_ERROR_FILE_FILENAME_AMBIGUOUS              MmsError = 91
	MMS_ERROR_FILE_FILE_BUSY                       MmsError = 92
	MMS_ERROR_FILE_FILENAME_SYNTAX_ERROR           MmsError = 93
	MMS_ERROR_FILE_CONTENT_TYPE_INVALID            MmsError = 94
	MMS_ERROR_FILE_POSITION_INVALID                MmsError = 95
	MMS_ERROR_FILE_FILE_ACCESS_DENIED              MmsError = 96
	MMS_ERROR_FILE_FILE_NON_EXISTENT               MmsError = 97
	MMS_ERROR_FILE_DUPLICATE_FILENAME              MmsError = 98
	MMS_ERROR_FILE_INSUFFICIENT_SPACE_IN_FILESTORE MmsError = 99

	/* reject codes */
	MMS_ERROR_REJECT_OTHER                    MmsError = 100
	MMS_ERROR_REJECT_UNKNOWN_PDU_TYPE         MmsError = 101
	MMS_ERROR_REJECT_INVALID_PDU              MmsError = 102
	MMS_ERROR_REJECT_UNRECOGNIZED_SERVICE     MmsError = 103
	MMS_ERROR_REJECT_UNRECOGNIZED_MODIFIER    MmsError = 104
	MMS_ERROR_REJECT_REQUEST_INVALID_ARGUMENT MmsError = 105
)

func (is *IedServer) GetMMSServer() *MMSServer {
	return &MMSServer{
		server: C.IedServer_getMmsServer(is.server),
	}
}