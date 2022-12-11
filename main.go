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

const VERSION = "0.0.0"

var gitHash = "N/A"
var printVersion = false

func init() {
	flag.StringVar(&op.Vault, "vault", "", "Vault to fetch servers from")
	flag.StringVar(&op.Tag, "tag", "ssh-gen", "Tag to lookup specific servers to add to the config")
	flag.StringVar(&outputPath, "out", "stdout", "Path of output file or stdout/stderr")
	flag.BoolVar(&printVersion, "version", false, "Print current version")
}

func main() {
	flag.Parse()

	if printVersion {
		fmt.Println(getVersionString())
		os.Exit(0)
	}

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
			val := strings.TrimSpace(f.Value)
			if err := validateConfigDecleration(f.Label, val); err != nil {
				log.Println(err)
				continue
			}
			fmt.Fprintf(handle, "\t%s %s\n", f.Label, val)
		}
	}
}

func getVersionString() string {
	return fmt.Sprintf("%s+%s", VERSION, gitHash)
}

func outputGenerateHeader(handle *os.File) {
	fmt.Fprintf(handle, "# Generated from sshconfig-1password on %v using sshconfig-1password version %s\n", time.Now().Format(time.RFC3339), getVersionString())
}

func getOuptutHandle() *os.File {
	var handle *os.File
	var err error
	switch outputPath {
	case "stdout":
		handle = os.Stdout
	case "stderr":
		handle = os.Stderr
	default:
		handle, err = os.OpenFile(outputPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Println("Unable to open output file", err)
			os.Exit(1)
		}
	}

	return handle
}
