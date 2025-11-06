package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/marrasen/iec61850"
)

// This example mirrors the libiec61850 C sample server_example_basic_io.c
// using the Go server APIs exposed by this repository.
//
// Notes vs. C sample:
// - Now mirrors the C sample more closely: uses SetWriteAccessPolicy for FC=DC,
//   installs connection indication and RCB event handlers via the Go API.

func main() {
	var tcpPort int
	flag.IntVar(&tcpPort, "p", 102, "TCP port to listen on")
	flag.Parse()

	fmt.Printf("Using libIEC61850 version %s\n", iec61850.GetVersionString())

	// 1) Load/create the data model. We reuse the repo's simple model used by tests
	// to match the object references used in the client example.
	model, err := loadSimpleIOModel()
	if err != nil {
		log.Fatalf("failed to load model: %v", err)
	}
	defer model.Destroy()

	// 2) Create and configure server
	cfg := iec61850.NewServerConfig()
	cfg.ReportBufferSize = 200000
	cfg.Edition = 1 // IEC 61850 Edition 2
	cfg.FileServiceBasePath = "./vmd-filestore/"
	cfg.EnableFileService = false
	cfg.EnableDynamicDataSetService = true
	cfg.EnableLogService = false
	cfg.MaxConnections = 2

	server := iec61850.NewServerWithConfig(cfg, model)
	defer server.Destroy()

	// Identity (vendor, model, revision)
	server.SetServerIdentity("MZ", "basic io", "1.6.0")

	// 3) Allow write access to FC=DC globally (enables writes to GGIO1.NamPlt.vendor, etc.)
	server.SetWriteAccessPolicy(iec61850.DC, iec61850.ACCESS_POLICY_ALLOW)

	// Install connection indication handler
	server.SetConnectionIndicationHandler(func(s *iec61850.IedServer, connected bool) {
		if connected {
			fmt.Println("Connection opened")
		} else {
			fmt.Println("Connection closed")
		}
	})

	// Install RCB event handler (prints like the C sample)
	server.SetRCBEventHandler(func(rcb *iec61850.ReportControlBlock, event iec61850.RCBEventType, parameterName string, serviceError iec61850.MmsDataAccessError) {
		fmt.Printf("RCB: %s event: %d\n", rcb.GetName(), event)
		if event == iec61850.RCB_EVENT_SET_PARAMETER || event == iec61850.RCB_EVENT_GET_PARAMETER {
			fmt.Printf("  param:  %s\n", parameterName)
			fmt.Printf("  result: %d\n", serviceError)
		}
		if event == iec61850.RCB_EVENT_ENABLE {
			fmt.Printf("   rptID:  %s\n", rcb.GetRptID())
			fmt.Printf("   datSet: %s\n", rcb.GetDataSet())
		}
	})

	// 4) Install control handlers for GGIO1.SPCSO1..4
	installSPCSOHandler := func(name string) {
		n := model.GetModelNodeByObjectReference("simpleIOGenericIO/GGIO1." + name)
		if n == nil {
			fmt.Printf("Warning: model node not found for %s\n", name)
			return
		}
		server.SetControlHandler(n, func(node *iec61850.ModelNode, action *iec61850.ControlAction, mmsValue *iec61850.MmsValue, test bool) iec61850.ControlHandlerResult {
			if test {
				return iec61850.CONTROL_RESULT_FAILED
			}
			// Expect boolean value
			if mmsValue == nil || mmsValue.Type != iec61850.Boolean {
				return iec61850.CONTROL_RESULT_FAILED
			}

			// Update t and stVal attributes
			nowMs := time.Now().UnixMilli()
			tNode := model.GetModelNodeByObjectReference("simpleIOGenericIO/GGIO1." + name + ".t")
			stValNode := model.GetModelNodeByObjectReference("simpleIOGenericIO/GGIO1." + name + ".stVal")
			server.UpdateUTCTimeAttributeValue(tNode, nowMs)
			b, _ := mmsValue.Value.(bool)
			// The server API doesn't expose UpdateBoolean; but writing stVal via UpdateInt32AttributeValue with 0/1
			// is not correct for boolean. However, libiec61850 maps boolean to MMS Boolean internally when the control
			// action is processed. The recommended approach is to return OK and let the stack update stVal based on control.
			// To emulate the C sample prints/update, we attempt a best-effort approach:
			if b {
				fmt.Println("received binary control command: on")
			} else {
				fmt.Println("received binary control command: off")
			}
			// If a boolean update API is added in the future, call it here for stVal.
			_ = stValNode
			return iec61850.CONTROL_RESULT_OK
		})
	}
	installSPCSOHandler("SPCSO1")
	installSPCSOHandler("SPCSO2")
	installSPCSOHandler("SPCSO3")
	installSPCSOHandler("SPCSO4")

	// 5) Start server
	server.Start(tcpPort)
	fmt.Printf("Listening on 0.0.0.0:%d\n", tcpPort)
	if !server.IsRunning() {
		fmt.Println("Starting server failed (maybe need admin/root permissions or port already in use)! Exit.")
		server.Destroy()
		os.Exit(1)
	}

	// Graceful shutdown on CTRL+C
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGTERM)

	// 6) Main loop: update four analog measurements with timestamps
	anIn := []string{"AnIn1", "AnIn2", "AnIn3", "AnIn4"}
	// Cache attribute nodes for performance
	anInT := make([]*iec61850.ModelNode, len(anIn))
	anInF := make([]*iec61850.ModelNode, len(anIn))
	for i, n := range anIn {
		anInT[i] = model.GetModelNodeByObjectReference("simpleIOGenericIO/GGIO1." + n + ".t")
		anInF[i] = model.GetModelNodeByObjectReference("simpleIOGenericIO/GGIO1." + n + ".mag.f")
	}

	t := 0.0
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for running := true; running; {
		select {
		case <-ticker.C:
			// Simulate data
			t += 0.1
			an := [4]float32{
				float32(math.Sin(t + 0.0)),
				float32(math.Sin(t + 1.0)),
				float32(math.Sin(t + 2.0)),
				float32(math.Sin(t + 3.0)),
			}
			nowMs := time.Now().UnixMilli()

			// In C sample they build a Timestamp and set flags (LeapSecondKnown, toggle ClockNotSynchronized).
			// The Go wrapper exposes UpdateUTCTimeAttributeValue which sets UTC ms; flags are not directly exposed.

			server.LockDataModel()
			for i := 0; i < 4; i++ {
				server.UpdateUTCTimeAttributeValue(anInT[i], nowMs)
				server.UpdateFloatAttributeValue(anInF[i], an[i])
			}
			server.UnlockDataModel()
		case sig := <-done:
			fmt.Printf("Signal received: %v\n", sig)
			running = false
		}
	}

	server.Stop()
}

func loadSimpleIOModel() (*iec61850.IedModel, error) {
	// Try several likely relative paths to the cfg present in this repo
	candidates := []string{
		"../../test/server/simpleIO_direct_control_goose.cfg",
		"../test/server/simpleIO_direct_control_goose.cfg",
		"test/server/simpleIO_direct_control_goose.cfg",
	}
	wd, _ := os.Getwd()
	for _, rel := range candidates {
		p := filepath.Clean(filepath.Join(wd, rel))
		if _, err := os.Stat(p); err == nil {
			return iec61850.CreateModelFromConfigFileEx(p)
		}
	}
	return nil, fmt.Errorf("simpleIO model cfg not found relative to %s", wd)
}
