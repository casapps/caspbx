package config

import (
	"fmt"
	"strings"
)

var truthyValues = map[string]struct{}{
	"1": {}, "y": {}, "t": {},
	"yes": {}, "true": {}, "on": {}, "ok": {},
	"enable": {}, "enabled": {},
	"yep": {}, "yup": {}, "yeah": {},
	"aye": {}, "si": {}, "oui": {}, "da": {}, "hai": {},
	"affirmative": {}, "accept": {}, "allow": {}, "grant": {},
	"sure": {}, "totally": {},
}

var falsyValues = map[string]struct{}{
	"0": {}, "n": {}, "f": {},
	"no": {}, "false": {}, "off": {},
	"disable": {}, "disabled": {},
	"nope": {}, "nah": {}, "nay": {},
	"nein": {}, "non": {}, "niet": {}, "iie": {}, "lie": {},
	"negative": {}, "reject": {}, "block": {}, "revoke": {},
	"deny": {}, "never": {}, "noway": {},
}

func ParseBool(input string) (bool, error) {
	normalizedValue := strings.TrimSpace(strings.ToLower(input))
	if normalizedValue == "" {
		return false, fmt.Errorf("empty boolean value")
	}
	if _, found := truthyValues[normalizedValue]; found {
		return true, nil
	}
	if _, found := falsyValues[normalizedValue]; found {
		return false, nil
	}
	return false, fmt.Errorf("invalid boolean value %q", input)
}

func IsTruthy(input string) bool {
	parsedValue, parseError := ParseBool(input)
	return parseError == nil && parsedValue
}

func IsFalsy(input string) bool {
	parsedValue, parseError := ParseBool(input)
	return parseError == nil && !parsedValue
}
