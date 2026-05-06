package model

import (
	"errors"
	"testing"
	"time"
)

func TestDefaultOrganizationPreferences(t *testing.T) {
	preferences := DefaultOrganizationPreferences()
	if preferences.DefaultMemberRole != OrganizationRoleMember {
		t.Fatalf("expected default member role, got %s", preferences.DefaultMemberRole)
	}
	if preferences.RequireTwoFactor {
		t.Fatalf("expected require_2fa to default false")
	}
	if !preferences.AllowInvites || !preferences.ShowMembers || !preferences.ShowActivity || !preferences.NotifyNewMember || !preferences.NotifyMemberLeave {
		t.Fatalf("expected org preferences boolean defaults to be enabled")
	}
}

func TestValidateOrgSlug(t *testing.T) {
	if validationError := ValidateOrgSlug("acme-corp", nil); validationError != nil {
		t.Fatalf("expected org slug to validate, got %v", validationError)
	}
	if validationError := ValidateOrgSlug("Acme", nil); validationError == nil {
		t.Fatalf("expected uppercase org slug to fail")
	}
	if validationError := ValidateOrgSlug("a", nil); validationError == nil {
		t.Fatalf("expected short org slug to fail")
	}
	if validationError := ValidateOrgSlug("acme--corp", nil); validationError == nil {
		t.Fatalf("expected consecutive hyphens to fail")
	}
	if validationError := ValidateOrgSlug("acme-", nil); validationError == nil {
		t.Fatalf("expected trailing hyphen org slug to fail")
	}
	if validationError := ValidateOrgSlug("abcdefghijklmnopqrstuvwxyzabcdefghijklmn", nil); validationError == nil {
		t.Fatalf("expected overlong org slug to fail")
	}
	if validationError := ValidateOrgSlug("admin", nil); validationError == nil {
		t.Fatalf("expected reserved org slug to fail")
	}
	if validationError := ValidateOrgSlug("taken-slug", func(string) error {
		return errors.New("already taken")
	}); validationError == nil {
		t.Fatalf("expected name collision to fail")
	}
	if validationError := ValidateOrgSlug("available-slug", func(string) error {
		return nil
	}); validationError != nil {
		t.Fatalf("expected available slug to validate, got %v", validationError)
	}
}

func TestBuildOrganizationMemberProfile(t *testing.T) {
	joinedAt := time.Unix(1, 0)

	publicProfile := BuildOrganizationMemberProfile(User{
		Username:    "public-user",
		DisplayName: "Public User",
		Visibility:  UserVisibilityPublic,
		AvatarURLs:  map[string]string{"64": "avatar"},
	}, OrganizationRoleMember, joinedAt)
	if publicProfile.ProfileVisibility != "public" || publicProfile.DisplayName != "Public User" {
		t.Fatalf("expected public org member profile, got %+v", publicProfile)
	}

	orgOnlyProfile := BuildOrganizationMemberProfile(User{
		Username:      "private-user",
		DisplayName:   "Private User",
		Visibility:    UserVisibilityPrivate,
		OrgVisibility: true,
		AvatarURLs:    map[string]string{"64": "avatar"},
	}, OrganizationRoleAdmin, joinedAt)
	if orgOnlyProfile.ProfileVisibility != "org_only" || orgOnlyProfile.DisplayName != "Private User" {
		t.Fatalf("expected org-only member profile, got %+v", orgOnlyProfile)
	}

	hiddenProfile := BuildOrganizationMemberProfile(User{
		Username:   "hidden-user",
		Visibility: UserVisibilityPrivate,
	}, OrganizationRoleOwner, joinedAt)
	if hiddenProfile.ProfileVisibility != "hidden" || hiddenProfile.DisplayName != "" {
		t.Fatalf("expected hidden profile, got %+v", hiddenProfile)
	}
}

func TestOrganizationMemberCanManageOrganization(t *testing.T) {
	if !(OrganizationMember{Role: OrganizationRoleOwner}).CanManageOrganization() {
		t.Fatalf("expected owner to manage organization")
	}
	if !(OrganizationMember{Role: OrganizationRoleAdmin}).CanManageOrganization() {
		t.Fatalf("expected admin to manage organization")
	}
	if (OrganizationMember{Role: OrganizationRoleMember}).CanManageOrganization() {
		t.Fatalf("expected member not to manage organization")
	}
}
