package model

import (
	"strings"
	"testing"
)

func TestRegistrationMode(t *testing.T) {
	if DefaultRegistrationMode() != RegistrationModePrivate {
		t.Fatalf("expected project default registration mode to be private")
	}

	if mode, parseError := ParseRegistrationMode("public"); parseError != nil || mode != RegistrationModePublic {
		t.Fatalf("expected public registration mode, got %v / %v", mode, parseError)
	}
	if mode, parseError := ParseRegistrationMode("private"); parseError != nil || mode != RegistrationModePrivate {
		t.Fatalf("expected private registration mode, got %v / %v", mode, parseError)
	}
	if mode, parseError := ParseRegistrationMode("disabled"); parseError != nil || mode != RegistrationModeDisabled {
		t.Fatalf("expected disabled registration mode, got %v / %v", mode, parseError)
	}
	if _, parseError := ParseRegistrationMode("unknown"); parseError == nil {
		t.Fatalf("expected invalid registration mode to fail")
	}
}

func TestDetectIdentifierType(t *testing.T) {
	if identifierType := DetectIdentifierType("12345"); identifierType != IdentifierTypeUserID {
		t.Fatalf("expected numeric identifier to be user id, got %s", identifierType)
	}
	if identifierType := DetectIdentifierType("user@example.com"); identifierType != IdentifierTypeEmail {
		t.Fatalf("expected email identifier, got %s", identifierType)
	}
	if identifierType := DetectIdentifierType("johndoe"); identifierType != IdentifierTypeUsername {
		t.Fatalf("expected username identifier, got %s", identifierType)
	}
}

func TestValidateUsername(t *testing.T) {
	if validationError := ValidateUsername("john_doe"); validationError != nil {
		t.Fatalf("expected username to validate, got %v", validationError)
	}
	if validationError := ValidateUsername("John"); validationError == nil {
		t.Fatalf("expected uppercase username to fail")
	}
	if validationError := ValidateUsername("jo"); validationError == nil {
		t.Fatalf("expected short username to fail")
	}
	if validationError := ValidateUsername(""); validationError == nil {
		t.Fatalf("expected empty username to fail")
	}
	if validationError := ValidateUsername("john--doe"); validationError == nil {
		t.Fatalf("expected consecutive separators to fail")
	}
	if validationError := ValidateUsername("john-"); validationError == nil {
		t.Fatalf("expected username ending with separator to fail")
	}
	if validationError := ValidateUsername("john_"); validationError == nil {
		t.Fatalf("expected username ending underscore to fail")
	}
	if validationError := ValidateUsername("john_-doe"); validationError == nil {
		t.Fatalf("expected mixed consecutive separators to fail")
	}
	if validationError := ValidateUsername("1john"); validationError == nil {
		t.Fatalf("expected username starting with number to fail")
	}
	if validationError := ValidateUsername("abcdefghijklmnopqrstuvwxyz1234567"); validationError == nil {
		t.Fatalf("expected overlong username to fail")
	}
	if validationError := ValidateUsername("admin"); validationError == nil {
		t.Fatalf("expected blocklisted username to fail")
	}
	if validationError := ValidateUsername("users"); validationError == nil {
		t.Fatalf("expected reserved username to fail")
	}
	if validationError := ValidateUsername("helpdesk"); validationError == nil {
		t.Fatalf("expected non-reserved blocklisted username to fail")
	}
	if validationError := ValidateUsername("officially-john"); validationError == nil {
		t.Fatalf("expected critical substring username to fail")
	}
}

func TestValidateEmail(t *testing.T) {
	if validationError := ValidateEmail("user.name+tag@example.com"); validationError != nil {
		t.Fatalf("expected valid email, got %v", validationError)
	}
	if validationError := ValidateEmail(""); validationError == nil {
		t.Fatalf("expected empty email to fail")
	}
	if validationError := ValidateEmail("user"); validationError == nil {
		t.Fatalf("expected malformed email to fail")
	}
	if validationError := ValidateEmail("@example.com"); validationError == nil {
		t.Fatalf("expected empty local part to fail")
	}
	if validationError := ValidateEmail("user@"); validationError == nil {
		t.Fatalf("expected empty domain part to fail")
	}
	if validationError := ValidateEmail(".user@example.com"); validationError == nil {
		t.Fatalf("expected invalid local part to fail")
	}
	if validationError := ValidateEmail("user..name@example.com"); validationError == nil {
		t.Fatalf("expected consecutive dots to fail")
	}
	if validationError := ValidateEmail("user@localhost"); validationError == nil {
		t.Fatalf("expected invalid domain to fail")
	}
	if validationError := ValidateEmail("user.@example.com"); validationError == nil {
		t.Fatalf("expected local part ending dot to fail")
	}
	if validationError := ValidateEmail("user!name@example.com"); validationError == nil {
		t.Fatalf("expected invalid local part characters to fail")
	}
	if validationError := ValidateEmail("user@-example.com"); validationError == nil {
		t.Fatalf("expected invalid domain format to fail")
	}
	if validationError := ValidateEmail(strings.Repeat("a", 65) + "@example.com"); validationError == nil {
		t.Fatalf("expected oversized local part to fail")
	}
	if validationError := ValidateEmail(strings.Repeat("a", 250) + "@b.com"); validationError == nil {
		t.Fatalf("expected oversized total email length to fail")
	}
}

func TestMaskEmailAndVisibility(t *testing.T) {
	user := User{
		ID:           42,
		AccountEmail: "john@example.com",
		Visibility:   UserVisibilityPrivate,
	}

	if user.MaskedAccountEmail() != "j***n@e***.com" {
		t.Fatalf("unexpected masked email %q", user.MaskedAccountEmail())
	}
	if NormalizeEmail(" User@Example.com ") != "user@example.com" {
		t.Fatalf("expected normalized email output")
	}
	if MaskEmail("invalid") != "" {
		t.Fatalf("expected invalid email mask to be blank")
	}
	if MaskEmail("a@localhost") != "" {
		t.Fatalf("expected invalid domain mask to be blank")
	}
	if MaskEmail("a@example.com") != "a***@e***.com" {
		t.Fatalf("unexpected single-char local email mask %q", MaskEmail("a@example.com"))
	}
	if user.PublicProfileVisibleTo(7) {
		t.Fatalf("expected private profile to be hidden from other users")
	}
	if !user.PublicProfileVisibleTo(42) {
		t.Fatalf("expected private profile to be visible to self")
	}

	publicUser := User{Visibility: UserVisibilityPublic}
	if !publicUser.PublicProfileVisibleTo(7) {
		t.Fatalf("expected public profile to be visible")
	}
}
