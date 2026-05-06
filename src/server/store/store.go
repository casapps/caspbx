package store

import (
	"context"
	"errors"

	"github.com/casapps/caspbx/src/server/model"
)

var ErrNotFound = errors.New("record not found")

type NamespaceLookup interface {
	UserExistsByName(context.Context, string) (bool, error)
	OrgExistsBySlug(context.Context, string) (bool, error)
}

type TenantLookup interface {
	FindTenantByHost(context.Context, string) (int64, error)
}

type DomainLookup interface {
	FindDomainByHost(context.Context, string) (int64, error)
}

type AdminCredentialStore interface {
	SaveAdmin(context.Context, model.Admin) (model.Admin, error)
	FindAdminByUsername(context.Context, string) (model.Admin, error)
	FindAdminByID(context.Context, int64) (model.Admin, error)
}

type UserCredentialStore interface {
	SaveUser(context.Context, model.User) (model.User, error)
	FindUserByUsername(context.Context, string) (model.User, error)
	FindUserByEmail(context.Context, string) (model.User, error)
	FindUserByID(context.Context, int64) (model.User, error)
}

type SessionStore interface {
	SaveSession(context.Context, model.Session) (model.Session, error)
	FindSessionByTokenHash(context.Context, model.SessionKind, string) (model.Session, error)
	DeleteSessionByTokenHash(context.Context, model.SessionKind, string) error
}

type TokenStore interface {
	SaveToken(context.Context, model.Token) (model.Token, error)
	FindTokenByHash(context.Context, model.TokenOwnerType, string) (model.Token, error)
	DeleteTokenByHash(context.Context, model.TokenOwnerType, string) error
}

type OrganizationStore interface {
	SaveOrganization(context.Context, model.Organization) (model.Organization, error)
	FindOrganizationBySlug(context.Context, string) (model.Organization, error)
	FindOrganizationByID(context.Context, int64) (model.Organization, error)
	SaveOrganizationPreferences(context.Context, model.OrganizationPreferences) (model.OrganizationPreferences, error)
	FindOrganizationPreferencesByOrgID(context.Context, int64) (model.OrganizationPreferences, error)
	SaveOrganizationMember(context.Context, model.OrganizationMember) (model.OrganizationMember, error)
	FindOrganizationMember(context.Context, int64) (model.OrganizationMember, error)
	FindOrganizationMemberByUserID(context.Context, int64, int64) (model.OrganizationMember, error)
	ListOrganizationMembers(context.Context, int64) ([]model.OrganizationMember, error)
}

type AuthStore interface {
	AdminCredentialStore
	UserCredentialStore
	SessionStore
	TokenStore
}

type RuntimeStore interface {
	AuthStore
	OrganizationStore
}
