package main

import (
	"flag"
	"fmt"
	"log"

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

	// ReadObject a DataSet: simpleIOGenericIO/LLN0.Events
	ref := "T11DR/RDRE1.NamPlt.swRev"
	read, err := client.ReadObject(ref, iec61850.DC)
	if err != nil {
		fmt.Println("failed to read", ref)
		log.Println(err)
	} else {
		fmt.Printf("ReadObject '%s' = %v\n", ref, read)
	}

	return nil
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
