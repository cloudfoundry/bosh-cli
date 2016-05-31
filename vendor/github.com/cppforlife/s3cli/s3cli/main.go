package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"github.com/cppforlife/s3cli/client"
)

var version string

func main() {
	configPath := flag.String("c", "", "configuration path")
	showVer := flag.Bool("v", false, "version")
	flag.Parse()

	if *showVer {
		fmt.Printf("version %s\n", version)
		os.Exit(0)
	}

	configFile, err := os.Open(*configPath)
	if err != nil {
		log.Fatalln(err)
	}

	blobstoreClient, err := client.New(configFile)
	if err != nil {
		log.Fatalln(err)
	}

	nonFlagArgs := flag.Args()
	if len(nonFlagArgs) != 3 {
		log.Fatalf("Expected 3 arguments got %d\n", len(nonFlagArgs))
	}

	cmd, src, dst := nonFlagArgs[0], nonFlagArgs[1], nonFlagArgs[2]

	switch cmd {
	case "put":
		var sourceFile *os.File
		sourceFile, err = os.Open(src)
		if err != nil {
			log.Fatalln(err)
		}

		defer sourceFile.Close()
		err = blobstoreClient.Put(sourceFile, dst)
	case "get":
		var dstFile *os.File
		dstFile, err = os.Create(dst)
		if err != nil {
			log.Fatalln(err)
		}

		defer dstFile.Close()
		err = blobstoreClient.Get(src, dstFile)
	default:
		log.Fatalf("unknown command: '%s'\n", cmd)
	}

	if err != nil {
		log.Fatalf("performing operation %s: %s\n", cmd, err)
	}
}
