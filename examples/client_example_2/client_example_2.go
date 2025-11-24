package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/marrasen/iec61850"
)

// This example mirrors the original libiec61850 C example (client_example1.c)
// using the Go wrapper APIs provided in this repository.
func run() error {
	var host string
	var port int
	var localIP string // not used currently (no explicit local bind in Go wrapper)
	var localPort int  // not used currently

	flag.StringVar(&host, "h", "localhost", "Host name or IP address")
	flag.IntVar(&port, "p", 102, "TCP port")
	flag.StringVar(&localIP, "local-ip", "", "Optional local IP to bind (not used)")
	flag.IntVar(&localPort, "local-port", -1, "Optional local TCP port to bind (not used)")
	flag.Parse()

	fmt.Printf("Using libIEC61850 version %s\n", iec61850.GetVersionString())
	fmt.Printf("Connecting to %s:%d\n", host, port)

	client, err := iec61850.NewClient(iec61850.Settings{
		Host:           host,
		Port:           port,
		ConnectTimeout: 10000,
		RequestTimeout: 10000,
	})
	if err != nil {
		return fmt.Errorf("failed to connect to %s:%d: %w", host, port, err)
	}
	defer client.Close()

	fmt.Println("Connected")

	// ReadObject RCB values (Buffered or Unbuffered as per C example name)
	rcbRef := "T11LD0/LLN0.BR.rcbMeasFlt01"
	rcb, err := client.GetRCBValues(rcbRef)
	if err != nil || rcb == nil {
		return fmt.Errorf("failed to read RCB values for %s: %v", rcbRef, err)
	}

	fmt.Printf("RptEna = %v\n", rcb.Ena)

	// Install report handler using the current RptId
	if rcb.RptId == "" {
		return fmt.Errorf("empty RptId for %s", rcbRef)
	}

	fmt.Printf("Installing report handler for %s (rptId=%s)\n", rcbRef, rcb.RptId)
	if err := client.InstallReportHandler(rcbRef, rcb.RptId, func(cr iec61850.ClientReport) {
		fmt.Printf("received report for %s\n", cr.GetRcbReference())
		// Print first 4 elements like the C example (GGIO1.SPCSOi.stVal)
		for i := 0; i < 4; i++ {
			reason := cr.GetReasonForInclusion(i)
			if reason != iec61850.IEC61850_REASON_NOT_INCLUDED {
				val, err := cr.GetElement(i)
				if err != nil {
					fmt.Printf("  element %d: error: %v\n", i, err)
					continue
				}
				fmt.Printf("  GGIO1.SPCSO%d.stVal: %s (included for reason %s)\n", i, val.Value, reason)
			}
		}
	}); err != nil {
		return fmt.Errorf("failed to install report handler: %w", err)
	}
	defer client.UninstallReportHandler(rcbRef)

	// Set trigger options and enable reporting, set IntgPd=5000ms (like C example)
	ops := iec61850.TrgOps{DataUpdate: true, TriggeredPeriodically: true, Gi: true}

	// Following the C example: set TrgOps + Ena + IntgPd together.
	if err := client.SetRCBValues(rcbRef, iec61850.ClientReportControlBlock{Ena: true, IntgPd: 5000, TrgOps: ops}); err != nil {
		fmt.Printf("report activation failed: %v\n", err)
	}

	time.Sleep(1 * time.Second)

	// Trigger GI
	if err := client.TriggerGIReport(rcbRef); err != nil {
		fmt.Printf("Error triggering a GI report: %v\n", err)
	}

	// Wait for a minute to potentially receive reports
	time.Sleep(60 * time.Second)

	// Disable reporting
	if err := client.SetRptEna(rcbRef, false); err != nil {
		fmt.Printf("disable reporting failed: %v\n", err)
	}

	return nil
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
