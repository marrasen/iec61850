package iec61850

// #include <iec61850_client.h>
// #include <iec61850_common.h>
// #include <mms_value.h>
import "C"

import (
	"encoding/hex"
	"unsafe"
)

type TrgOps struct {
	DataChange            bool // Value change
	QualityChange         bool // Quality change
	DataUpdate            bool // Data update
	TriggeredPeriodically bool // Periodic trigger (integrity)
	Gi                    bool // GI (general interrogation) trigger
	Transient             bool // Transient
}
type OptFlds struct {
	SequenceNumber     bool // Sequence number
	TimeOfEntry        bool // Report timestamp
	ReasonForInclusion bool // Reason code (reason for inclusion)
	DataSetName        bool // Data set
	DataReference      bool // Data reference
	BufferOverflow     bool // Buffer overflow indicator
	EntryID            bool // Report entry identifier
	ConfigRevision     bool // Configuration revision
}

type ClientReportControlBlock struct {
	Ena     bool    // Enable
	IntgPd  int     // Integrity period (ms)
	Resv    bool    // Reservation for URCB
	TrgOps  TrgOps  // Trigger options
	OptFlds OptFlds // Report options
	RptId   string  // RCB report ID
	DatSet  string  // Data set reference
	Owner   string  // Current owner (IP:port) if enabled
}

func (c *Client) GetRCBValues(objectReference string) (*ClientReportControlBlock, error) {
	var clientError C.IedClientError
	cObjectRef := C.CString(objectReference)
	defer C.free(unsafe.Pointer(cObjectRef))
	rcb := C.IedConnection_getRCBValues(c.conn, &clientError, cObjectRef, nil)
	if rcb == nil {
		return nil, GetIedClientError(clientError)
	}
	defer C.ClientReportControlBlock_destroy(rcb)
	// Convert Owner from MMS octet string to a hex string (may contain binary data)
	ownerMms := C.ClientReportControlBlock_getOwner(rcb)
	ownerStr := ""
	if ownerMms != nil {
		sz := C.MmsValue_getOctetStringSize(ownerMms)
		buf := C.MmsValue_getOctetStringBuffer(ownerMms)
		if sz > 0 && buf != nil {
			b := C.GoBytes(unsafe.Pointer(buf), C.int(sz))
			ownerStr = hex.EncodeToString(b)
		}
	}

	info := &ClientReportControlBlock{
		Ena:     c.getRCBEnable(rcb),
		IntgPd:  int(c.getRCBIntgPd(rcb)),
		Resv:    c.getRCBResv(rcb),
		TrgOps:  c.getTrgOps(rcb),
		OptFlds: c.getOptFlds(rcb),
		RptId:   C.GoString(C.ClientReportControlBlock_getRptId(rcb)),
		DatSet:  C.GoString(C.ClientReportControlBlock_getDataSetReference(rcb)),
		Owner:   ownerStr,
	}
	return info, nil
}

func (c *Client) getRCBEnable(rcb C.ClientReportControlBlock) bool {
	enable := C.ClientReportControlBlock_getRptEna(rcb)
	return bool(enable)
}

func (c *Client) getRCBIntgPd(rcb C.ClientReportControlBlock) uint32 {
	intgPd := C.ClientReportControlBlock_getIntgPd(rcb)
	return uint32(intgPd)
}

func (c *Client) getRCBResv(rcb C.ClientReportControlBlock) bool {
	resv := C.ClientReportControlBlock_getResv(rcb)
	return bool(resv)
}

func (c *Client) getOptFlds(rcb C.ClientReportControlBlock) OptFlds {
	optFlds := C.ClientReportControlBlock_getOptFlds(rcb)
	g := int(optFlds)
	return OptFlds{
		SequenceNumber:     IsBitSet(g, 0),
		TimeOfEntry:        IsBitSet(g, 1),
		ReasonForInclusion: IsBitSet(g, 2),
		DataSetName:        IsBitSet(g, 3),
		DataReference:      IsBitSet(g, 4),
		BufferOverflow:     IsBitSet(g, 5),
		EntryID:            IsBitSet(g, 6),
		ConfigRevision:     IsBitSet(g, 7),
	}
}

func (c *Client) getTrgOps(rcb C.ClientReportControlBlock) TrgOps {
	trgOps := C.ClientReportControlBlock_getTrgOps(rcb)
	g := int(trgOps)
	return TrgOps{
		DataChange:            IsBitSet(g, 0),
		QualityChange:         IsBitSet(g, 1),
		DataUpdate:            IsBitSet(g, 2),
		TriggeredPeriodically: IsBitSet(g, 3),
		Gi:                    IsBitSet(g, 4),
		Transient:             IsBitSet(g, 5),
	}
}

func (c *Client) SetRCBValues(objectReference string, settings ClientReportControlBlock) error {
	// NOTE: This combined write is convenient but may lead to IED_ERROR_TEMPORARILY_UNAVAILABLE
	// on Buffered RCBs when enabling and configuring in a single call. Prefer the granular
	// setters (SetRptEna/SetTrgOps/SetDataSetReference) with correct ordering for BRCB.
	var clientError C.IedClientError
	cObjectRef := C.CString(objectReference)
	defer C.free(unsafe.Pointer(cObjectRef))
	rcb := C.ClientReportControlBlock_create(cObjectRef)
	defer C.ClientReportControlBlock_destroy(rcb)
	var trgOps, optFlds C.int
	// trgOps
	if settings.TrgOps.DataChange {
		trgOps = trgOps | C.TRG_OPT_DATA_CHANGED
	}
	if settings.TrgOps.QualityChange {
		trgOps = trgOps | C.TRG_OPT_QUALITY_CHANGED
	}
	if settings.TrgOps.DataUpdate {
		trgOps = trgOps | C.TRG_OPT_DATA_UPDATE
	}
	if settings.TrgOps.TriggeredPeriodically {
		trgOps = trgOps | C.TRG_OPT_INTEGRITY
	}
	if settings.TrgOps.Gi {
		trgOps = trgOps | C.TRG_OPT_GI
	}
	if settings.TrgOps.Transient {
		trgOps = trgOps | C.TRG_OPT_TRANSIENT
	}
	// optFlds
	if settings.OptFlds.SequenceNumber {
		optFlds = optFlds | C.RPT_OPT_SEQ_NUM
	}
	if settings.OptFlds.TimeOfEntry {
		optFlds = optFlds | C.RPT_OPT_TIME_STAMP
	}
	if settings.OptFlds.ReasonForInclusion {
		optFlds = optFlds | C.RPT_OPT_REASON_FOR_INCLUSION
	}
	if settings.OptFlds.DataSetName {
		optFlds = optFlds | C.RPT_OPT_DATA_SET
	}
	if settings.OptFlds.DataReference {
		optFlds = optFlds | C.RPT_OPT_DATA_REFERENCE
	}
	if settings.OptFlds.BufferOverflow {
		optFlds = optFlds | C.RPT_OPT_BUFFER_OVERFLOW
	}
	if settings.OptFlds.EntryID {
		optFlds = optFlds | C.RPT_OPT_ENTRY_ID
	}
	if settings.OptFlds.ConfigRevision {
		optFlds = optFlds | C.RPT_OPT_CONF_REV
	}

	C.ClientReportControlBlock_setTrgOps(rcb, trgOps)               // Trigger options
	C.ClientReportControlBlock_setRptEna(rcb, C.bool(settings.Ena)) // Report enable
	C.ClientReportControlBlock_setResv(rcb, C.bool(settings.Resv))
	C.ClientReportControlBlock_setIntgPd(rcb, C.uint32_t(settings.IntgPd)) // Integrity period (ms)
	C.ClientReportControlBlock_setOptFlds(rcb, optFlds)

	if bool(C.ClientReportControlBlock_isBuffered(rcb)) {
		C.IedConnection_setRCBValues(c.conn, &clientError, rcb, C.RCB_ELEMENT_RPT_ENA|C.RCB_ELEMENT_TRG_OPS|C.RCB_ELEMENT_INTG_PD, true)
	} else {
		C.IedConnection_setRCBValues(c.conn, &clientError, rcb, C.RCB_ELEMENT_RESV|C.RCB_ELEMENT_RPT_ENA|C.RCB_ELEMENT_TRG_OPS|C.RCB_ELEMENT_INTG_PD, true)
	}

	if err := GetIedClientError(clientError); err != nil {
		return err
	}
	return nil
}

// SetRptEna writes only the RptEna flag of an RCB (enable/disable reporting).
func (c *Client) SetRptEna(objectReference string, enable bool) error {
	var clientError C.IedClientError
	cObjectRef := C.CString(objectReference)
	defer C.free(unsafe.Pointer(cObjectRef))
	rcb := C.ClientReportControlBlock_create(cObjectRef)
	defer C.ClientReportControlBlock_destroy(rcb)
	C.ClientReportControlBlock_setRptEna(rcb, C.bool(enable))
	C.IedConnection_setRCBValues(c.conn, &clientError, rcb, C.RCB_ELEMENT_RPT_ENA, true)
	return GetIedClientError(clientError)
}

// SetTrgOps writes only the trigger options of an RCB.
func (c *Client) SetTrgOps(objectReference string, ops TrgOps) error {
	var clientError C.IedClientError
	cObjectRef := C.CString(objectReference)
	defer C.free(unsafe.Pointer(cObjectRef))
	rcb := C.ClientReportControlBlock_create(cObjectRef)
	defer C.ClientReportControlBlock_destroy(rcb)
	var trgOps C.int
	if ops.DataChange {
		trgOps = trgOps | C.TRG_OPT_DATA_CHANGED
	}
	if ops.QualityChange {
		trgOps = trgOps | C.TRG_OPT_QUALITY_CHANGED
	}
	if ops.DataUpdate {
		trgOps = trgOps | C.TRG_OPT_DATA_UPDATE
	}
	if ops.TriggeredPeriodically {
		trgOps = trgOps | C.TRG_OPT_INTEGRITY
	}
	if ops.Gi {
		trgOps = trgOps | C.TRG_OPT_GI
	}
	if ops.Transient {
		trgOps = trgOps | C.TRG_OPT_TRANSIENT
	}
	C.ClientReportControlBlock_setTrgOps(rcb, trgOps)
	C.IedConnection_setRCBValues(c.conn, &clientError, rcb, C.RCB_ELEMENT_TRG_OPS, true)
	return GetIedClientError(clientError)
}

// SetBufTm writes only the BufTm (buffer time in ms) of an RCB.
func (c *Client) SetBufTm(objectReference string, bufTm uint32) error {
	var clientError C.IedClientError
	cObjectRef := C.CString(objectReference)
	defer C.free(unsafe.Pointer(cObjectRef))
	rcb := C.ClientReportControlBlock_create(cObjectRef)
	defer C.ClientReportControlBlock_destroy(rcb)
	C.ClientReportControlBlock_setBufTm(rcb, C.uint32_t(bufTm))
	C.IedConnection_setRCBValues(c.conn, &clientError, rcb, C.RCB_ELEMENT_BUF_TM, true)
	return GetIedClientError(clientError)
}

// SetIntgPd writes only the IntgPd (integrity period in ms) of an RCB.
func (c *Client) SetIntgPd(objectReference string, intgPd uint32) error {
	var clientError C.IedClientError
	cObjectRef := C.CString(objectReference)
	defer C.free(unsafe.Pointer(cObjectRef))
	rcb := C.ClientReportControlBlock_create(cObjectRef)
	defer C.ClientReportControlBlock_destroy(rcb)
	C.ClientReportControlBlock_setIntgPd(rcb, C.uint32_t(intgPd))
	C.IedConnection_setRCBValues(c.conn, &clientError, rcb, C.RCB_ELEMENT_INTG_PD, true)
	return GetIedClientError(clientError)
}

// SetGI sets the GI flag of an RCB (whether a GI is created on enable).
func (c *Client) SetGI(objectReference string, gi bool) error {
	var clientError C.IedClientError
	cObjectRef := C.CString(objectReference)
	defer C.free(unsafe.Pointer(cObjectRef))
	rcb := C.ClientReportControlBlock_create(cObjectRef)
	defer C.ClientReportControlBlock_destroy(rcb)
	C.ClientReportControlBlock_setGI(rcb, C.bool(gi))
	C.IedConnection_setRCBValues(c.conn, &clientError, rcb, C.RCB_ELEMENT_GI, true)
	return GetIedClientError(clientError)
}

// Note: Buffered vs Unbuffered is determined by selecting the appropriate RCB object (BRCB/URCB)
// in the server model. There is no client-side API to toggle this property.

// SetDataSetReference writes only the dataset reference (DatSet) of an RCB.
// When writing from the client, pass the fully-qualified MMS object reference including the IED name,
// e.g. "IEDLD0/LLN0$FileEvts".
func (c *Client) SetDataSetReference(objectReference string, dataSetRef string) error {
	var clientError C.IedClientError
	cObjectRef := C.CString(objectReference)
	defer C.free(unsafe.Pointer(cObjectRef))
	cDs := C.CString(dataSetRef)
	defer C.free(unsafe.Pointer(cDs))
	rcb := C.ClientReportControlBlock_create(cObjectRef)
	defer C.ClientReportControlBlock_destroy(rcb)
	C.ClientReportControlBlock_setDataSetReference(rcb, cDs)
	C.IedConnection_setRCBValues(c.conn, &clientError, rcb, C.RCB_ELEMENT_DATSET, true)
	return GetIedClientError(clientError)
}

func IsBitSet(val int, pos int) bool {
	return (val & (1 << pos)) != 0
}
