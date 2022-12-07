package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

type OpListItem struct {
	Id string `json:"id"`
}

type OpItemDetails struct {
	Title         string      `json:"title"`
	Sections      []OpSection `json:"sections"`
	Fields        []OpField   `json:"fields"`
	FieldLabelMap map[string]OpField
	SectionMap    map[string][]OpField
}

func (item OpItemDetails) getHostname() string {
	v := item.FieldLabelMap["URL"]
	return v.Value
}

func (item OpItemDetails) getHost() string {
	str := strings.Split(item.Title, " ")[0]
	return strings.ToLower(str)
}

func (item OpItemDetails) getUser() string {
	v := item.FieldLabelMap["username"]
	return v.Value
}

type OpSection struct {
	Id    string `json:"id"`
	Label string `json:"label"`
}

type OpField struct {
	Id      string     `json:"id"`
	Type    string     `json:"type"`
	Label   string     `json:"label"`
	Value   string     `json:"value"`
	Section *OpSection `json:"section"`
}

type Op struct {
	Vault string
	Tag   string
}

func (op *Op) exec(args ...string) ([]byte, error) {
	args = append(args, "--format=json")
	if op.Vault != "" {
		args = append(args, fmt.Sprintf("--vault=%s", op.Vault))
	}
	cmd := exec.Command("op", args...)
	result, err := cmd.CombinedOutput()
	if err != nil {
		err = fmt.Errorf("%w; %s", err, string(result))
		return nil, err
	}
	return result, err
}

func (op *Op) whoAmI() error {
	cmd := exec.Command("op", "whoami")
	result, err := cmd.CombinedOutput()
	e := &exec.ExitError{}
	if errors.As(err, &e) {
		if e.ExitCode() != 0 {
			return err
		}

		return fmt.Errorf("%w; unknown error: %s", err, result)
	}

	return err
}

func (op *Op) listItems() ([]OpListItem, error) {
	result, err := op.exec("item", "list", fmt.Sprintf("--tags=%s", op.Tag), "--categories=SERVER")
	if err != nil {
		return nil, err
	}
	list := []OpListItem{}
	json.Unmarshal(result, &list)
	return list, err
}

func (op *Op) getItem(id string) (OpItemDetails, error) {
	result, err := op.exec("item", "get", id)
	if err != nil {
		return OpItemDetails{}, err
	}
	item := OpItemDetails{}
	item.FieldLabelMap = make(map[string]OpField)
	item.SectionMap = make(map[string][]OpField)
	json.Unmarshal(result, &item)
	for _, f := range item.Fields {
		item.FieldLabelMap[f.Label] = f
		if f.Section != nil {
			arr, ok := item.SectionMap[f.Section.Label]
			if !ok {
				arr = []OpField{}
			}

			arr = append(arr, f)
			item.SectionMap[f.Section.Label] = arr
		}
	}

	return item, err
}

var op Op
var outputPath string

func init() {
	flag.StringVar(&op.Vault, "vault", "", "Vault to fetch servers from")
	flag.StringVar(&op.Tag, "tag", "ssh-gen", "Tag to lookup specific servers to add to the config")
	flag.StringVar(&outputPath, "out", "", "Path of output file (defaults to stdout)")
	flag.Parse()
}

func main() {
	if err := op.whoAmI(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	items, err := op.listItems()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	var handle *os.File
	if outputPath != "" {
		handle, err = os.OpenFile(outputPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Println("Unable to open output file", err)
			os.Exit(1)
		}
	} else {
		handle = os.Stdout
	}

	fmt.Fprintf(handle, "# Generated from sshconfig-1password on %v\n", time.Now().Format(time.RFC3339))
	for _, item := range items {
		details, err := op.getItem(item.Id)
		if err != nil {
			fmt.Println(err)
			continue
		}

		fmt.Fprintf(handle, "Host %s\n", details.getHost())

		fmt.Fprintf(handle, "\tHostname %s\n", details.getHostname())
		fmt.Fprintf(handle, "\tUser %s\n", details.getUser())

		if section, ok := details.SectionMap["SSH Config"]; ok {
			for _, f := range section {
			    fmt.Fprintf(handle, "\t%s %s\n", f.Label, strings.TrimSpace(f.Value))
			}
		}

		handle.WriteString("\n")
	}
}
