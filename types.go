package iec61850

type MmsType int

type MmsValue struct {
	Type  MmsType
	Value interface{}
}

// data types
const (
	Array MmsType = iota
	Structure
	Boolean
	BitString
	Integer
	Unsigned
	Float
	OctetString
	VisibleString
	GeneralizedTime
	BinaryTime
	Bcd
	ObjId
	String
	UTCTime
	DataAccessError
	Int8
	Int16
	Int32
	Int64
	Uint8
	Uint16
	Uint32
)

type MmsDataAccessError int

const (
	DATA_ACCESS_ERROR_SUCCESS_NO_UPDATE             MmsDataAccessError = -3
	DATA_ACCESS_ERROR_NO_RESPONSE                   MmsDataAccessError = -2
	DATA_ACCESS_ERROR_SUCCESS                       MmsDataAccessError = -1
	DATA_ACCESS_ERROR_OBJECT_INVALIDATED            MmsDataAccessError = 0
	DATA_ACCESS_ERROR_HARDWARE_FAULT                MmsDataAccessError = 1
	DATA_ACCESS_ERROR_TEMPORARILY_UNAVAILABLE       MmsDataAccessError = 2
	DATA_ACCESS_ERROR_OBJECT_ACCESS_DENIED          MmsDataAccessError = 3
	DATA_ACCESS_ERROR_OBJECT_UNDEFINED              MmsDataAccessError = 4
	DATA_ACCESS_ERROR_INVALID_ADDRESS               MmsDataAccessError = 5
	DATA_ACCESS_ERROR_TYPE_UNSUPPORTED              MmsDataAccessError = 6
	DATA_ACCESS_ERROR_TYPE_INCONSISTENT             MmsDataAccessError = 7
	DATA_ACCESS_ERROR_OBJECT_ATTRIBUTE_INCONSISTENT MmsDataAccessError = 8
	DATA_ACCESS_ERROR_OBJECT_ACCESS_UNSUPPORTED     MmsDataAccessError = 9
	DATA_ACCESS_ERROR_OBJECT_NONE_EXISTENT          MmsDataAccessError = 10
	DATA_ACCESS_ERROR_OBJECT_VALUE_INVALID          MmsDataAccessError = 11
	DATA_ACCESS_ERROR_UNKNOWN                       MmsDataAccessError = 12
)

// AccessPolicy maps to libiec61850 AccessPolicy
// ACCESS_POLICY_ALLOW allows writes, ACCESS_POLICY_DENY denies writes for given FC
// Values must match the C enum ordering.
type AccessPolicy int

const (
	ACCESS_POLICY_ALLOW AccessPolicy = iota
	ACCESS_POLICY_DENY
)

type ControlHandlerResult int

const (
	CONTROL_RESULT_FAILED ControlHandlerResult = iota
	CONTROL_RESULT_OK
	CONTROL_RESULT_WAITING
)

type ControlModel int

const (
	// CONTROL_MODEL_STATUS_ONLY No support for control functions. Control object only support status information.
	CONTROL_MODEL_STATUS_ONLY ControlModel = iota
	// CONTROL_MODEL_DIRECT_NORMAL Direct control with normal security: Supports Operate, TimeActivatedOperate (optional), and Cancel (optional).
	CONTROL_MODEL_DIRECT_NORMAL
	// CONTROL_MODEL_SBO_NORMAL Select before operate (SBO) with normal security: Supports Select, Operate, TimeActivatedOperate (optional), and Cancel (optional).
	CONTROL_MODEL_SBO_NORMAL
	// CONTROL_MODEL_DIRECT_ENHANCED Direct control with enhanced security (enhanced security includes the CommandTermination service)
	CONTROL_MODEL_DIRECT_ENHANCED
	// CONTROL_MODEL_SBO_ENHANCED Select before operate (SBO) with enhanced security (enhanced security includes the CommandTermination service)
	CONTROL_MODEL_SBO_ENHANCED
)

type AcseAuthenticationMechanism int

const (
	// ACSE_AUTH_NONE Neither ACSE nor TLS authentication used
	ACSE_AUTH_NONE AcseAuthenticationMechanism = iota

	// ACSE_AUTH_PASSWORD Use ACSE password for client authentication
	ACSE_AUTH_PASSWORD

	// ACSE_AUTH_CERTIFICATE Use ACSE certificate for client authentication
	ACSE_AUTH_CERTIFICATE

	// ACSE_AUTH_TLS Use TLS certificate for client authentication
	ACSE_AUTH_TLS
)

// ACSIClass represents the different ACSI class types as defined in IEC 61850
type ACSIClass int

const (
	ACSI_CLASS_DATA_OBJECT ACSIClass = iota
	ACSI_CLASS_DATA_SET
	ACSI_CLASS_BRCB
	ACSI_CLASS_URCB
	ACSI_CLASS_LCB
	ACSI_CLASS_LOG
	ACSI_CLASS_SGCB
	ACSI_CLASS_GoCB
	ACSI_CLASS_GsCB
	ACSI_CLASS_MSVCB
	ACSI_CLASS_USVCB
)
