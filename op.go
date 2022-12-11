package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
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
    host := item.Title
    if v, ok := item.FieldLabelMap["Host"]; ok {
        host = v.Value
    }
	return host
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
			return fmt.Errorf("1Password CLI not authenticated, please first authenticate your CLI instance and try again; %w", err)
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
