package config

import (
	"fmt"
	"strings"
)

type AppMode int

const (
	AppModeProduction AppMode = iota
	AppModeDevelopment
)

func (appMode AppMode) String() string {
	switch appMode {
	case AppModeDevelopment:
		return "development"
	default:
		return "production"
	}
}

func ParseAppMode(input string) (AppMode, error) {
	switch strings.ToLower(strings.TrimSpace(input)) {
	case "", "prod", "production":
		return AppModeProduction, nil
	case "dev", "development":
		return AppModeDevelopment, nil
	default:
		return AppModeProduction, fmt.Errorf("invalid app mode %q", input)
	}
}

func ResolveAppMode(cliModeValue string, cliModeSet bool, envModeValue string) AppMode {
	if cliModeSet {
		parsedMode, parseError := ParseAppMode(cliModeValue)
		if parseError == nil {
			return parsedMode
		}
		return AppModeProduction
	}

	if envModeValue != "" {
		parsedMode, parseError := ParseAppMode(envModeValue)
		if parseError == nil {
			return parsedMode
		}
	}

	return AppModeProduction
}

func ResolveDebugEnabled(cliDebugValue bool, cliDebugSet bool, envDebugValue string) bool {
	if cliDebugSet {
		return cliDebugValue
	}
	return IsTruthy(envDebugValue)
}

func FormatAppModeLabel(appMode AppMode, debugEnabled bool) string {
	if debugEnabled {
		return appMode.String() + " [debugging]"
	}
	return appMode.String()
}
