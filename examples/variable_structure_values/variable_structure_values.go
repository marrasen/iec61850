package main

import (
	"flag"
	"fmt"
	"log"
	"slices"
	"strings"
	"time"

	"github.com/marrasen/iec61850"
)

func run() error {
	var host string
	var port int

	flag.StringVar(&host, "h", "127.0.0.1", "Host name or IP address")
	flag.IntVar(&port, "p", 102, "Port number")

	flag.Parse()

	fmt.Printf("Using libIEC61850 version %s\n\n", iec61850.GetVersionString())

	fmt.Println("Connecting to server...")
	client, err := iec61850.NewClient(iec61850.Settings{
		Host:           host,
		Port:           port,
		ConnectTimeout: 10000,
		RequestTimeout: 10000,
	})
	if err != nil {
		return fmt.Errorf("failed to connect to server: %w", err)
	}
	defer client.Close()

	fmt.Println("Connected")

	start := time.Now()
	list, err := client.GetVariableValues()
	if err != nil {
		return fmt.Errorf("failed to GetDataModel: %w", err)
	}
	fmt.Printf("GetVariableValues took %v\n", time.Since(start))

	slices.SortFunc(list, func(a, b iec61850.VariableTypeValue) int {
		return strings.Compare(a.Ref, b.Ref)
	})

	for _, v := range list {
		fmt.Printf("%s: %v\n", v.Ref, v.Value)
	}

	return nil
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
