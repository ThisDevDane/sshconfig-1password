package main

import "testing"

func TestValidateUser(t *testing.T) {
	err := ValidateUser("User", "pi")
	if err != nil {
		t.Error("Says valid user is invalid", err)
	}

	err = ValidateUser("User", "pi yiouad")
	if err == nil {
		t.Error("Says invalid user is valid", err)
	}
}

func TestValidateHostname(t *testing.T) {
	err := ValidateHostname("Hostname", "188.69.91.76")
	if err != nil {
		t.Error("Does not allow IPv4 address", err)
	}

	err = ValidateHostname("Hostname", "8892:205a:9865:b454:05a9:7060:6e3b:868c")
	if err != nil {
		t.Error("Does not allow IPv6 address", err)
	}

	err = ValidateHostname("Hostname", "google.com")
	if err != nil {
		t.Error("Does not allow URL", err)
	}

	err = ValidateHostname("Hostname", `::aest.com`)
	if err == nil {
		t.Error("Allows invalid URL", err)
	}
}

func TestValidateYesNoAsk(t *testing.T) {
	inputs := []struct {
		value       string
		shouldError bool
	}{
		{"yes", false},
		{"no", false},
		{"ask", false},
		{"Yes", true},
		{"No", true},
		{"Ask", true},
		{"wow", true},
	}
	for _, item := range inputs {

		err := ValidateYesNoAsk("Test", item.value)
		didError := err != nil
		if item.shouldError != didError {
			t.Errorf("\"ValidateYesNoAsk('%s')\" failed, expected to error -> %v, got -> %v", item.value, item.shouldError, didError)
		} else {
			t.Logf("\"ValidateYesNoAsk('%s')\" succeded, expected to error -> %v, got -> %v", item.value, item.shouldError, didError)
		}
	}
}
