package main

import (
	"fmt"
	"net/netip"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

var validKeys = map[string]func(string, string) error{
	"Host":                             bypassValidation,
	"Match":                            bypassValidation,
	"AddKeysToAgent":                   bypassValidation,
	"AddressFamily":                    bypassValidation,
	"BatchMode":                        validateYesNo,
	"BindAddress":                      bypassValidation,
	"BindInterface":                    bypassValidation,
	"CanonicalDomains":                 bypassValidation,
	"CanonicalizeFallbackLocal":        validateYesNo,
	"CanonicalizeHostname":             bypassValidation,
	"CanonicalizeMaxDots":              bypassValidation,
	"CanonicalizePermittedCNAMEs":      bypassValidation,
	"CASignatureAlgorithms":            bypassValidation,
	"CertificateFile":                  bypassValidation,
	"CheckHostIP":                      validateYesNo,
	"Ciphers":                          bypassValidation,
	"ClearAllForwardings":              validateYesNo,
	"Compression":                      validateYesNo,
	"ConnectionAttempts":               bypassValidation,
	"ConnectTimeout":                   bypassValidation,
	"ControlMaster":                    bypassValidation,
	"ControlPath":                      bypassValidation,
	"ControlPersist":                   bypassValidation,
	"DynamicForward":                   bypassValidation,
	"EnableSSHKeysign":                 validateYesNo,
	"EscapeChar":                       bypassValidation,
	"ExitOnForwardFailure":             validateYesNo,
	"FingerprintHash":                  bypassValidation,
	"ForkAfterAuthentication":          validateYesNo,
	"ForwardAgent":                     bypassValidation,
	"ForwardX11":                       validateYesNo,
	"ForwardX11Timeout":                bypassValidation,
	"ForwardX11Trusted":                validateYesNo,
	"GatewayPorts":                     validateYesNo,
	"GlobalKnownHostsFile":             bypassValidation,
	"GSSAPIAuthentication":             validateYesNo,
	"GSSAPIDelegateCredentials":        validateYesNo,
	"HashKnownHosts":                   validateYesNo,
	"HostbasedAcceptedAlgorithms":      bypassValidation,
	"HostbasedAuthentication":          validateYesNo,
	"HostKeyAlgorithms":                bypassValidation,
	"HostKeyAlias":                     bypassValidation,
	"Hostname":                         ValidateHostname,
	"IdentitiesOnly":                   validateYesNo,
	"IdentityAgent":                    bypassValidation,
	"IdentityFile":                     bypassValidation,
	"IgnoreUnknown":                    bypassValidation,
	"Include":                          bypassValidation,
	"IPQoS":                            bypassValidation,
	"KbdInteractiveAuthentication":     validateYesNo,
	"KbdInteractiveDevices":            bypassValidation,
	"KexAlgorithms":                    bypassValidation,
	"KnownHostsCommand":                bypassValidation,
	"LocalCommand":                     bypassValidation,
	"LocalForward":                     bypassValidation,
	"LogLevel":                         bypassValidation,
	"LogVerbose":                       bypassValidation,
	"MACs":                             bypassValidation,
	"NoHostAuthenticationForLocalhost": validateYesNo,
	"NumberOfPasswordPrompts":          bypassValidation,
	"PasswordAuthentication":           validateYesNo,
	"PermitLocalCommand":               validateYesNo,
	"PermitRemoteOpen":                 bypassValidation,
	"PKCS11Provider":                   bypassValidation,
	"Port":                             validateIsNumber,
	"PreferredAuthentications":         bypassValidation,
	"ProxyCommand":                     bypassValidation,
	"ProxyJump":                        bypassValidation,
	"ProxyUseFdpass":                   bypassValidation,
	"PubkeyAcceptedAlgorithms":         bypassValidation,
	"PubkeyAuthentication":             validateYesNo,
	"RekeyLimit":                       bypassValidation,
	"RemoteCommand":                    bypassValidation,
	"RemoteForward":                    bypassValidation,
	"RequestTTY":                       bypassValidation,
	"RevokedHostKeys":                  bypassValidation,
	"SecurityKeyProvider":              bypassValidation,
	"SendEnv":                          bypassValidation,
	"ServerAliveCountMax":              bypassValidation,
	"ServerAliveInterval":              bypassValidation,
	"SessionType":                      bypassValidation,
	"SetEnv":                           bypassValidation,
	"StdinNull":                        validateYesNo,
	"StreamLocalBindMask":              bypassValidation,
	"StreamLocalBindUnlink":            validateYesNo,
	"StrictHostKeyChecking":            bypassValidation,
	"SyslogFacility":                   bypassValidation,
	"TCPKeepAlive":                     validateYesNo,
	"Tunnel":                           bypassValidation,
	"TunnelDevice":                     bypassValidation,
	"UpdateHostKeys":                   ValidateYesNoAsk,
	"User":                             ValidateUser,
	"UserKnownHostsFile":               bypassValidation,
	"VerifyHostKeyDNS":                 ValidateYesNoAsk,
	"VisualHostKey":                    validateYesNo,
	"XAuthLocation":                    bypassValidation,
}

func validateConfigDecleration(key string, value string) error {
	validateFunc, ok := validKeys[key]
	if !ok {
		return fmt.Errorf("%s is not a valid ssh config decleration", key)
	}

	cleanedValue := strings.TrimSpace(value)
	if cleanedValue == "" {
		return fmt.Errorf("value for '%s' is an empty string or pure whitspace '%s'", key, value)
	}

	return validateFunc(key, cleanedValue)
}

func bypassValidation(key string, value string) error {
	return nil
}

var whitespaceRegex = regexp.MustCompile(`\s`)

func ValidateUser(key string, value string) error {
	if whitespaceRegex.MatchString(value) {
		return fmt.Errorf("decleration '%s' value: '%s' contains a whitespace character which is not permitted", key, value)
	}

	return nil
}

func ValidateHostname(key string, value string) error {
	_, netErr := netip.ParseAddr(value)
	_, urlErr := url.Parse(value)

	if netErr != nil && urlErr != nil {
		return fmt.Errorf("decleration value for '%s' is not valid: '%s' (%w)", key, value, errCoalesce(netErr, urlErr))
	}

	return nil
}

func validateIsNumber(key string, value string) error {
	_, err := strconv.Atoi(value)
	if err != nil {
		return fmt.Errorf("decleration value for '%s' is not a valid number: '%s'", key, value)
	}
	return nil
}

func validateYesNo(key string, value string) error {
	if value == "Yes" || value == "No" {
		return fmt.Errorf("declaration value for '%s' has wrong casing, yes/no values must be all lowercase, current: '%s'", key, value)
	}

	if value != "yes" && value != "no" {
		return fmt.Errorf("declaration value for '%s' is invalid, must be yes or no: '%s'", key, value)
	}

	return nil
}

func ValidateYesNoAsk(key string, value string) error {
	if value == "Ask" {
		return fmt.Errorf("declaration value for '%s' has wrong casing, yes/no/ask value must be all lowercase, current: '%s'", key, value)
	}

	err := validateYesNo(key, value)

	if value != "ask" && err != nil {
		return fmt.Errorf("declaration value for '%s' is invalid, must be yes, no or ask: '%s'", key, value)
	}

	return nil
}

func errCoalesce(errors ...error) error {
	for _, e := range errors {
		if e != nil {
			return e
		}
	}

	return nil
}
