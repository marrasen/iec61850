package iec61850

type DataModel struct {
	LDs []LD
}

// LD is a Logical Device
type LD struct {
	Data string
	LNs  []LN
}

// LN is a Logical Node
type LN struct {
	Data      string
	Ref       string
	DOs       []DO
	DSs       []DS
	URReports []URReport
	BRReports []BRReport
}

// URReport are Unbuffer Reports
type URReport struct {
	Data string
	Ref  string
}

// BRReport are Buffered Reports
type BRReport struct {
	Data string
	Ref  string
}

// DS represents a DataSet
type DS struct {
	Data        string
	DSRefs      []DSRef
	IsDeletable bool
}

type DSRef struct {
	Data string
}

// DO represents a Data Object
type DO struct {
	Data string
	DAs  []DA
}

// DA represents a Data Attribute
type DA struct {
	Data string
	DAs  []DA
	Ref  string
	FC   FC
}
