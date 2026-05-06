package model

import "time"

type AdminRole string

const (
	AdminRoleSuperAdmin AdminRole = "superadmin"
	AdminRoleAdmin      AdminRole = "admin"
	AdminRoleReadOnly   AdminRole = "readonly"
)

type Admin struct {
	ID                int64
	Username          string
	AccountEmail      string
	NotificationEmail string
	PasswordHash      string
	Role              AdminRole
	Enabled           bool
	Source            string
	ExternalID        string
	Groups            []string
	CreatedAt         time.Time
	UpdatedAt         time.Time
	LastLoginAt       time.Time
	LastSyncAt        time.Time
}
