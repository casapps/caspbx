package model

import (
	"errors"
	"regexp"
	"strings"
	"time"
)

type OrganizationRole string
type OrganizationVisibility string

const (
	OrganizationRoleOwner  OrganizationRole = "owner"
	OrganizationRoleAdmin  OrganizationRole = "admin"
	OrganizationRoleMember OrganizationRole = "member"

	OrganizationVisibilityPublic  OrganizationVisibility = "public"
	OrganizationVisibilityPrivate OrganizationVisibility = "private"
)

var orgSlugPattern = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]*[a-z0-9])?$`)

type Organization struct {
	ID          int64
	Slug        string
	Name        string
	Description string
	Website     string
	Location    string
	Visibility  OrganizationVisibility
	OwnerID     int64
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type OrganizationPreferences struct {
	OrgID              int64
	DefaultMemberRole OrganizationRole
	RequireTwoFactor  bool
	AllowInvites      bool
	ShowMembers       bool
	ShowActivity      bool
	NotifyNewMember   bool
	NotifyMemberLeave bool
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type OrganizationMember struct {
	ID        int64
	OrgID     int64
	UserID    int64
	Role      OrganizationRole
	CreatedAt time.Time
}

type OrganizationMemberProfile struct {
	Username          string
	DisplayName       string
	Avatar            map[string]string
	Role              OrganizationRole
	JoinedAt          time.Time
	ProfileVisibility string
}

func DefaultOrganizationPreferences() OrganizationPreferences {
	return OrganizationPreferences{
		DefaultMemberRole: OrganizationRoleMember,
		RequireTwoFactor:  false,
		AllowInvites:      true,
		ShowMembers:       true,
		ShowActivity:      true,
		NotifyNewMember:   true,
		NotifyMemberLeave: true,
	}
}

func ValidateOrgSlug(input string, availabilityChecker func(string) error) error {
	slug := strings.TrimSpace(input)
	if slug != strings.ToLower(slug) {
		return errors.New("slug must be lowercase alphanumeric with hyphens")
	}
	if len(slug) < 2 || len(slug) > 39 {
		return errors.New("slug must be 2-39 characters")
	}
	if !orgSlugPattern.MatchString(slug) {
		return errors.New("slug must be lowercase alphanumeric with hyphens")
	}
	if strings.Contains(slug, "--") {
		return errors.New("slug cannot contain consecutive hyphens")
	}
	if IsReservedName(slug) {
		return errors.New("slug is reserved")
	}
	if availabilityChecker != nil {
		return availabilityChecker(slug)
	}
	return nil
}

func BuildOrganizationMemberProfile(user User, role OrganizationRole, joinedAt time.Time) OrganizationMemberProfile {
	profile := OrganizationMemberProfile{
		Username: user.Username,
		Role:     role,
		JoinedAt: joinedAt,
	}

	switch {
	case user.Visibility == UserVisibilityPublic:
		profile.DisplayName = user.DisplayName
		profile.Avatar = user.AvatarURLs
		profile.ProfileVisibility = "public"
	case user.OrgVisibility:
		profile.DisplayName = user.DisplayName
		profile.Avatar = user.AvatarURLs
		profile.ProfileVisibility = "org_only"
	default:
		profile.ProfileVisibility = "hidden"
	}

	return profile
}

func (member OrganizationMember) CanManageOrganization() bool {
	return member.Role == OrganizationRoleOwner || member.Role == OrganizationRoleAdmin
}
