package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	iec "github.com/marrasen/iec61850"
)

// DataSetMap indexes for LLN0/StatDR according to your IED's DataSet order
type DataSetMap struct {
	IdxRcdMade    int // 0
	IdxRcdStr     int // 1
	IdxERcdStored int // 2
	IdxERcdDelete int // 3
	IdxEMemFull   int // 4
	IdxEOwRcd     int // 5
	IdxEPerTrg    int // 6
	IdxEManTrg    int // 7
}

type AppContext struct {
	Map DataSetMap
}

// printUtcIfPresent tries to find a UTC timestamp in the DO structure and prints it (ms since epoch)
func printUtcIfPresent(doStruct *iec.MmsValue) {
	if doStruct == nil {
		return
	}
	if doStruct.Type != iec.Structure && doStruct.Type != iec.Array {
		return
	}
	elems, ok := doStruct.Value.([]*iec.MmsValue)
	if !ok {
		return
	}
	for _, el := range elems {
		if el == nil {
			continue
		}
		switch el.Type {
		case iec.UTCTime:
			// toGoValue for UTCTime returns uint32 seconds since epoch
			if sec, ok := el.Value.(uint32); ok {
				ms := uint64(sec) * 1000
				fmt.Printf("  t(ms since epoch): %d\n", ms)
				return
			}
		case iec.BinaryTime:
			// BinaryTime returns uint64 milliseconds since epoch
			if ms, ok := el.Value.(uint64); ok {
				fmt.Printf("  t(ms since epoch): %d\n", ms)
				return
			}
		}
	}
}

// findStValInDO returns the first boolean/integer/unsigned element as stVal (heuristic)
func findStValInDO(doStruct *iec.MmsValue) *iec.MmsValue {
	if doStruct == nil {
		return nil
	}
	if doStruct.Type != iec.Structure && doStruct.Type != iec.Array {
		return nil
	}
	elems, ok := doStruct.Value.([]*iec.MmsValue)
	if !ok {
		return nil
	}
	for _, el := range elems {
		if el == nil {
			continue
		}
		switch el.Type {
		case iec.Boolean, iec.Integer, iec.Unsigned, iec.Int32, iec.Int16, iec.Int8, iec.Uint32, iec.Uint16, iec.Uint8:
			return el
		}
	}
	return nil
}

// enableBrcb configures and enables a buffered RCB following a safe ordering
// Note: Buffered vs Unbuffered is selected by choosing an appropriate RCB reference (BRCB vs URCB) on the server.
func enableBrcb(client *iec.Client, brcbRef, datasetRef string, bufTmMs, intgPdMs uint32, giOnEnable bool) error {
	// Read RCB to get RptId and check existence
	rcb, err := client.GetRCBValues(brcbRef)
	if err != nil || rcb == nil {
		return fmt.Errorf("get RCB values failed for %s: %v", brcbRef, err)
	}

	// Disable before reconfiguration
	if err := client.SetRptEna(brcbRef, false); err != nil {
		return fmt.Errorf("disable RCB failed: %w", err)
	}

	// Set dataset reference
	if datasetRef != "" {
		if err := client.SetDataSetReference(brcbRef, datasetRef); err != nil {
			return fmt.Errorf("set datasetRef failed: %w", err)
		}
	}

	// Trigger options: DATA_CHANGED + INTEGRITY, commonly used for StatDR
	ops := iec.TrgOps{DataChange: true, TriggeredPeriodically: true, Gi: giOnEnable}
	if err := client.SetTrgOps(brcbRef, ops); err != nil {
		return fmt.Errorf("set TrgOps failed: %w", err)
	}

	// BufTm & IntgPd
	if bufTmMs > 0 {
		if err := client.SetBufTm(brcbRef, bufTmMs); err != nil {
			return fmt.Errorf("set BufTm failed: %w", err)
		}
	}
	if intgPdMs > 0 {
		if err := client.SetIntgPd(brcbRef, intgPdMs); err != nil {
			return fmt.Errorf("set IntgPd failed: %w", err)
		}
	}

	// GI behavior
	if err := client.SetGI(brcbRef, giOnEnable); err != nil {
		fmt.Printf("warning: SetGI failed: %v\n", err)
	}

	// Finally enable reporting
	if err := client.SetRptEna(brcbRef, true); err != nil {
		return fmt.Errorf("enable RCB failed: %w", err)
	}
	_ = rcb // rcb kept to ensure RCB exists; rptId used later by InstallReportHandler
	return nil
}

func run() error {
	var host string
	var port int
	var brcbRef string
	var datasetRef string

	// Defaults: adjust to your IED
	flag.StringVar(&host, "h", "192.0.2.10", "Host name or IP address")
	flag.IntVar(&port, "p", 102, "TCP port")
	flag.StringVar(&brcbRef, "rcb", "T11DR/LLN0.BR01", "RCB reference (e.g. T11DR/LLN0.BR01)")
	flag.StringVar(&datasetRef, "ds", "T11DR/LLN0.StatDR", "DataSet reference (e.g. T11DR/LLN0.StatDR)")
	flag.Parse()

	fmt.Printf("Using libIEC61850 version %s\n", iec.GetVersionString())
	fmt.Printf("Connecting to %s:%d\n", host, port)

	client, err := iec.NewClient(iec.Settings{
		Host:           host,
		Port:           port,
		ConnectTimeout: 10000,
		RequestTimeout: 10000,
	})
	if err != nil {
		return fmt.Errorf("failed to connect to %s:%d: %w", host, port, err)
	}
	defer client.Close()

	// Map indexes of DataSet elements (adjust to your DataSet order)
	app := &AppContext{Map: DataSetMap{
		IdxRcdMade:    0,
		IdxRcdStr:     1,
		IdxERcdStored: 2,
		IdxERcdDelete: 3,
		IdxEMemFull:   4,
		IdxEOwRcd:     5,
		IdxEPerTrg:    6,
		IdxEManTrg:    7,
	}}

	// Read RCB to get rptId for handler installation
	rcb, err := client.GetRCBValues(brcbRef)
	if err != nil || rcb == nil {
		return fmt.Errorf("failed to read RCB values for %s: %v", brcbRef, err)
	}
	if rcb.RptId == "" {
		return fmt.Errorf("empty RptId for %s", brcbRef)
	}

	// Install report handler before enabling, so early reports aren't missed
	if err := client.InstallReportHandler(brcbRef, rcb.RptId, func(cr iec.ClientReport) {
		dsVals, err := cr.GetDataSetValues()
		if err != nil {
			return
		}
		if dsVals.Type != iec.Array && dsVals.Type != iec.Structure {
			return
		}
		arr, ok := dsVals.Value.([]*iec.MmsValue)
		if !ok {
			return
		}
		// Access ERcdStored DO by index from the dataset
		if app.Map.IdxERcdStored < 0 || app.Map.IdxERcdStored >= len(arr) {
			return
		}
		doERcdStored := arr[app.Map.IdxERcdStored]
		stVal := findStValInDO(doERcdStored)
		if stVal == nil {
			return
		}
		// Convert stVal to integer 0/1
		val := 0
		switch v := stVal.Value.(type) {
		case bool:
			if v {
				val = 1
			}
		case int64:
			if v != 0 {
				val = 1
			}
		case uint32:
			if v != 0 {
				val = 1
			}
		case int32:
			if v != 0 {
				val = 1
			}
		case int16:
			if v != 0 {
				val = 1
			}
		case int8:
			if v != 0 {
				val = 1
			}
		case uint16:
			if v != 0 {
				val = 1
			}
		case uint8:
			if v != 0 {
				val = 1
			}
		}

		if val == 1 {
			fmt.Printf("\n=== New disturbance stored (ERcdStored=1) ===\n")
			fmt.Printf("RCB: %s\n", cr.GetRcbReference())
			if cr.HasDataSetName() {
				fmt.Printf("DataSet: %s\n", cr.GetDataSetName())
			}
			printUtcIfPresent(doERcdStored)
			fmt.Println("(Trigger your COMTRADE download here.)")
		}
	}); err != nil {
		return fmt.Errorf("failed to install report handler: %w", err)
	}
	defer client.UninstallReportHandler(brcbRef)

	// Configure and enable buffered reporting to LLN0.StatDR
	// Select a BRCB by passing a buffered RCB reference in brcbRef.
	if err := enableBrcb(client, brcbRef, datasetRef, 50, 10000, true); err != nil {
		return fmt.Errorf("failed to enable BRCB: %w", err)
	}

	fmt.Printf("Waiting for reports (ERcdStored=1) on %s (DataSet=%s)...\n", brcbRef, datasetRef)
	for {
		time.Sleep(500 * time.Millisecond)
	}
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
