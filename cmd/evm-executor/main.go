package main

import (
	"flag"
	"log"

	"github.com/rollkit/go-execution-evm/pkg/execution"
)

func main() {
	var (
		address = flag.String("address", "127.0.0.1:8551", "The address to listen on for gRPC server")
	)
	flag.Parse()

	// Initialize EVMExecution
	_, err := execution.NewEVMExecution(*address)
	println(err)
	println(address)
	if err != nil {
		log.Fatalf("Failed to create EVM execution: %v", err)
	}
}
