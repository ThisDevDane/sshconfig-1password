package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

var op Op
var outputPath string

func init() {
	flag.StringVar(&op.Vault, "vault", "", "Vault to fetch servers from")
	flag.StringVar(&op.Tag, "tag", "ssh-gen", "Tag to lookup specific servers to add to the config")
	flag.StringVar(&outputPath, "out", "", "Path of output file (defaults to stdout)")
}

func main() {
	flag.Parse()
	if err := op.whoAmI(); err != nil {
		log.Fatalln(err)
	}

	items, err := op.listItems()
	if err != nil {
		log.Fatalln(err)
	}

	handle := getOuptutHandle()
	outputGenerateHeader(handle)

	for _, item := range items {
		details, err := op.getItem(item.Id)
		if err != nil {
			log.Println(err)
			continue
		}

		outputHostConfig(handle, details)

		handle.WriteString("\n")
	}
}

func outputHostConfig(handle *os.File, item OpItemDetails) {
	fmt.Fprintf(handle, "Host %s\n", item.getHost())
	fmt.Fprintf(handle, "\tHostname %s\n", item.getHostname())
	fmt.Fprintf(handle, "\tUser %s\n", item.getUser())

	if section, ok := item.SectionMap["SSH Config"]; ok {
		for _, f := range section {
			fmt.Fprintf(handle, "\t%s %s\n", f.Label, strings.TrimSpace(f.Value))
		}
	}
}

func outputGenerateHeader(handle *os.File) {
	fmt.Fprintf(handle, "# Generated from sshconfig-1password on %v\n", time.Now().Format(time.RFC3339))
}

func getOuptutHandle() *os.File {
	var handle *os.File
	var err error
	if outputPath != "" {
		handle, err = os.OpenFile(outputPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Println("Unable to open output file", err)
			os.Exit(1)
		}
	} else {
		handle = os.Stdout
	}

	return handle
}
