package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	iec "github.com/marrasen/iec61850"
)

// findFirstStVal tries to find a suitable stVal within a DO structure by
// returning the first boolean/integer-like field.
func findFirstStVal(doStruct *iec.MmsValue) *iec.MmsValue {
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

// findTimestampMs scans the DO structure and returns a timestamp in ms since epoch
// if a UTCTime or BinaryTime element is present.
func findTimestampMs(doStruct *iec.MmsValue) (uint64, bool) {
	if doStruct == nil {
		return 0, false
	}
	if doStruct.Type != iec.Structure && doStruct.Type != iec.Array {
		return 0, false
	}
	elems, ok := doStruct.Value.([]*iec.MmsValue)
	if !ok {
		return 0, false
	}
	for _, el := range elems {
		if el == nil {
			continue
		}
		switch el.Type {
		case iec.UTCTime:
			if sec, ok := el.Value.(uint32); ok {
				return uint64(sec) * 1000, true
			}
		case iec.BinaryTime:
			if ms, ok := el.Value.(uint64); ok {
				return ms, true
			}
		}
	}
	return 0, false
}

func main() {
	var host string
	var port int
	var ld string
	var ln string
	var ds string
	var idxERcdStored int

	flag.StringVar(&host, "h", "192.0.2.10", "IED host or IP")
	flag.IntVar(&port, "p", 102, "IED TCP port")
	flag.StringVar(&ld, "ld", "T11LD0", "Logical device name")
	flag.StringVar(&ln, "ln", "LLN0", "Logical node name")
	flag.StringVar(&ds, "ds", "T11LD0/LLN0$StatDR", "DataSet reference")
	flag.IntVar(&idxERcdStored, "idx", 2, "Index of ERcdStored DO within the StatDR DataSet")
	flag.Parse()

	fmt.Printf("libIEC61850: %s\n", iec.GetVersionString())
	fmt.Printf("Connecting to %s:%d...\n", host, port)

	client, err := iec.NewClient(iec.Settings{
		Host:           host,
		Port:           port,
		ConnectTimeout: 10000,
		RequestTimeout: 10000,
	})
	if err != nil {
		log.Fatalf("connect failed: %v", err)
	}
	defer client.Close()

	// Abort application if the client connection is lost
	_ = client.InstallConnectionClosedHandler(func() {
		log.Fatalf("Connection to IED lost/closed - aborting application")
	})

	// Select and enable a StatDR BRCB
	rcbRef, cleanup, err := client.PickAndEnableStatDRBRCB(ld, ln, ds)
	if err != nil {
		log.Fatalf("PickAndEnableStatDRBRCB: %v", err)
	}
	log.Printf("Using RCB: %s", rcbRef)

	// Ensure cleanup disables reporting
	defer func() {
		if cleanup != nil {
			if err := cleanup(); err != nil {
				log.Printf("cleanup failed: %v", err)
			}
		}
	}()

	// ReadObject RCB to get rptId for handler registration
	rcb, err := client.GetRCBValues(rcbRef)
	if err != nil || rcb == nil || rcb.RptId == "" {
		log.Fatalf("failed to read RptId for %s: %v", rcbRef, err)
	}

	// Install report handler
	if err := client.InstallReportHandler(rcbRef, rcb.RptId, func(cr iec.ClientReport) {
		log.Printf("Report handler callback received!")
		dsVals, err := cr.GetDataSetValues()
		if err != nil {
			return
		}
		if dsVals.Type != iec.Array && dsVals.Type != iec.Structure {
			log.Printf("Note: Unknown type %v", dsVals.Type)
			return
		}
		arr, ok := dsVals.Value.([]*iec.MmsValue)
		if !ok || idxERcdStored < 0 || idxERcdStored >= len(arr) {
			log.Printf("Wrong index for idxERcdStored: %d, expected max: %d", idxERcdStored, len(arr))
			return
		}
		doStruct := arr[idxERcdStored]
		stVal := findFirstStVal(doStruct)
		if stVal == nil {
			log.Printf("Could not find stVal for %s", doStruct)
			return
		}
		// Evaluate stVal as integer/boolean
		triggered := false
		switch stVal.Type {
		case iec.Boolean:
			if b, ok := stVal.Value.(bool); ok {
				triggered = b
			}
		case iec.Integer, iec.Int32, iec.Int16, iec.Int8:
			if v, ok := stVal.Value.(int32); ok {
				triggered = (v == 1)
			}
		case iec.Unsigned, iec.Uint32, iec.Uint16, iec.Uint8:
			if v, ok := stVal.Value.(uint32); ok {
				triggered = (v == 1)
			}
		}
		if triggered {
			if ms, ok := findTimestampMs(doStruct); ok {
				ts := time.Unix(0, int64(ms)*int64(time.Millisecond)).UTC()
				log.Printf("Ny störning lagrad! ts=%s (ms=%d)\n", ts.Format(time.RFC3339), ms)
			} else {
				log.Printf("Ny störning lagrad!\n")
			}
			// Here you could trigger COMTRADE retrieval
		} else {
			log.Printf("Did not detect trigger on %+v: %+v", doStruct, stVal)
		}
	}); err != nil {
		log.Fatalf("InstallReportHandler: %v", err)
	}

	log.Printf("Waiting for reports on %s (ds=%s)... press Ctrl+C to exit", rcbRef, ds)

	// Wait for Ctrl+C
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	<-sig
	log.Printf("Exiting")
}
