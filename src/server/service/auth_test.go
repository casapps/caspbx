package service

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/casapps/caspbx/src/server/model"
	"github.com/casapps/caspbx/src/server/store"
)

type failingAuthStore struct {
	findAdminError   error
	findUserError    error
	saveAdminError   error
	saveUserError    error
	saveSessionError error
	findSessionError error
	deleteError      error
	saveTokenError   error
	findTokenError   error
	deleteTokenError error
	admin            model.Admin
	user             model.User
	session          model.Session
	token            model.Token
}

func (authStore failingAuthStore) SaveAdmin(context.Context, model.Admin) (model.Admin, error) {
	return model.Admin{}, authStore.saveAdminError
}

func (authStore failingAuthStore) FindAdminByUsername(context.Context, string) (model.Admin, error) {
	if authStore.findAdminError != nil {
		return model.Admin{}, authStore.findAdminError
	}
	return authStore.admin, nil
}

func (authStore failingAuthStore) FindAdminByID(context.Context, int64) (model.Admin, error) {
	if authStore.findAdminError != nil {
		return model.Admin{}, authStore.findAdminError
	}
	return authStore.admin, nil
}

func (authStore failingAuthStore) SaveUser(context.Context, model.User) (model.User, error) {
	return model.User{}, authStore.saveUserError
}

func (authStore failingAuthStore) FindUserByUsername(context.Context, string) (model.User, error) {
	if authStore.findUserError != nil {
		return model.User{}, authStore.findUserError
	}
	return authStore.user, nil
}

func (authStore failingAuthStore) FindUserByEmail(context.Context, string) (model.User, error) {
	if authStore.findUserError != nil {
		return model.User{}, authStore.findUserError
	}
	return authStore.user, nil
}

func (authStore failingAuthStore) FindUserByID(context.Context, int64) (model.User, error) {
	if authStore.findUserError != nil {
		return model.User{}, authStore.findUserError
	}
	return authStore.user, nil
}

func (authStore failingAuthStore) SaveSession(context.Context, model.Session) (model.Session, error) {
	if authStore.saveSessionError != nil {
		return model.Session{}, authStore.saveSessionError
	}
	return authStore.session, nil
}

func (authStore failingAuthStore) FindSessionByTokenHash(context.Context, model.SessionKind, string) (model.Session, error) {
	if authStore.findSessionError != nil {
		return model.Session{}, authStore.findSessionError
	}
	return authStore.session, nil
}

func (authStore failingAuthStore) DeleteSessionByTokenHash(context.Context, model.SessionKind, string) error {
	return authStore.deleteError
}

func (authStore failingAuthStore) SaveToken(_ context.Context, token model.Token) (model.Token, error) {
	if authStore.saveTokenError != nil {
		return model.Token{}, authStore.saveTokenError
	}
	if authStore.token.TokenHash == "" && authStore.token.ID == 0 {
		return token, nil
	}
	return authStore.token, nil
}

func (authStore failingAuthStore) FindTokenByHash(context.Context, model.TokenOwnerType, string) (model.Token, error) {
	if authStore.findTokenError != nil {
		return model.Token{}, authStore.findTokenError
	}
	return authStore.token, nil
}

func (authStore failingAuthStore) DeleteTokenByHash(context.Context, model.TokenOwnerType, string) error {
	return authStore.deleteTokenError
}

type updatingSessionStore struct {
	store.AuthStore
	found model.Session
	saved model.Session
}

func (authStore updatingSessionStore) FindSessionByTokenHash(context.Context, model.SessionKind, string) (model.Session, error) {
	return authStore.found, nil
}

func (authStore updatingSessionStore) SaveSession(context.Context, model.Session) (model.Session, error) {
	return authStore.saved, nil
}

type updatingTokenStore struct {
	store.AuthStore
	found model.Token
	saved model.Token
}

func (authStore updatingTokenStore) FindTokenByHash(context.Context, model.TokenOwnerType, string) (model.Token, error) {
	return authStore.found, nil
}

func (authStore updatingTokenStore) SaveToken(context.Context, model.Token) (model.Token, error) {
	return authStore.saved, nil
}

type refreshDeleteFailStore struct {
	failingAuthStore
}

func (authStore refreshDeleteFailStore) SaveToken(_ context.Context, token model.Token) (model.Token, error) {
	return token, nil
}

type refreshCreateFailStore struct {
	token model.Token
}

func (authStore refreshCreateFailStore) SaveAdmin(context.Context, model.Admin) (model.Admin, error) {
	return model.Admin{}, nil
}

func (authStore refreshCreateFailStore) FindAdminByUsername(context.Context, string) (model.Admin, error) {
	return model.Admin{}, nil
}

func (authStore refreshCreateFailStore) FindAdminByID(context.Context, int64) (model.Admin, error) {
	return model.Admin{}, nil
}

func (authStore refreshCreateFailStore) SaveUser(context.Context, model.User) (model.User, error) {
	return model.User{}, nil
}

func (authStore refreshCreateFailStore) FindUserByUsername(context.Context, string) (model.User, error) {
	return model.User{}, nil
}

func (authStore refreshCreateFailStore) FindUserByEmail(context.Context, string) (model.User, error) {
	return model.User{}, nil
}

func (authStore refreshCreateFailStore) FindUserByID(context.Context, int64) (model.User, error) {
	return model.User{}, nil
}

func (authStore refreshCreateFailStore) SaveSession(context.Context, model.Session) (model.Session, error) {
	return model.Session{}, nil
}

func (authStore refreshCreateFailStore) FindSessionByTokenHash(context.Context, model.SessionKind, string) (model.Session, error) {
	return model.Session{}, nil
}

func (authStore refreshCreateFailStore) DeleteSessionByTokenHash(context.Context, model.SessionKind, string) error {
	return nil
}

func (authStore refreshCreateFailStore) SaveToken(_ context.Context, token model.Token) (model.Token, error) {
	if token.TokenHash == authStore.token.TokenHash {
		return token, nil
	}
	return model.Token{}, errors.New("create failed")
}

func (authStore refreshCreateFailStore) FindTokenByHash(context.Context, model.TokenOwnerType, string) (model.Token, error) {
	return authStore.token, nil
}

func (authStore refreshCreateFailStore) DeleteTokenByHash(context.Context, model.TokenOwnerType, string) error {
	return nil
}

func TestCan(t *testing.T) {
	serverAdminBindings := []RoleBinding{{Role: RoleServerAdmin}}
	if !Can(serverAdminBindings, PermissionManagePlatform, Scope{}) {
		t.Fatalf("expected server admin to manage platform")
	}
	if Can(nil, PermissionManagePlatform, Scope{}) {
		t.Fatalf("expected empty bindings to fail")
	}

	tenantAdminBindings := []RoleBinding{{Role: RoleTenantAdmin, TenantID: 7, UserID: 70}}
	if !Can(tenantAdminBindings, PermissionManageTenant, Scope{TenantID: 7}) {
		t.Fatalf("expected tenant admin to manage same tenant")
	}
	if Can(tenantAdminBindings, PermissionManageTenant, Scope{TenantID: 8}) {
		t.Fatalf("expected tenant admin to fail on other tenant")
	}
	if Can(tenantAdminBindings, PermissionManagePlatform, Scope{}) {
		t.Fatalf("expected tenant admin not to manage platform")
	}
	if !Can(tenantAdminBindings, PermissionManageOrganization, Scope{TenantID: 7, OrganizationID: 11}) {
		t.Fatalf("expected tenant admin to manage organization in same tenant")
	}
	if !Can(tenantAdminBindings, PermissionManageUserDomains, Scope{TenantID: 7, UserID: 70}) {
		t.Fatalf("expected tenant admin to manage user domains in same tenant")
	}
	if !Can(tenantAdminBindings, PermissionUseSupervisorTools, Scope{TenantID: 7}) {
		t.Fatalf("expected tenant admin to use supervisor tools in same tenant")
	}
	if !Can(tenantAdminBindings, PermissionManageOrgDomains, Scope{TenantID: 7, OrganizationID: 11}) {
		t.Fatalf("expected tenant admin to manage org domains in same tenant")
	}

	orgAdminBindings := []RoleBinding{{Role: RoleOrganizationAdmin, OrganizationID: 11, UserID: 101}}
	if !Can(orgAdminBindings, PermissionManageOrganization, Scope{OrganizationID: 11}) {
		t.Fatalf("expected org admin to manage same org")
	}
	if Can(orgAdminBindings, PermissionManageOrganization, Scope{OrganizationID: 12}) {
		t.Fatalf("expected org admin to fail on other org")
	}
	if !Can(orgAdminBindings, PermissionManageOrgDomains, Scope{OrganizationID: 11}) {
		t.Fatalf("expected org admin to manage same org domains")
	}
	if !Can(orgAdminBindings, PermissionUseSupervisorTools, Scope{OrganizationID: 11}) {
		t.Fatalf("expected org admin to use supervisor tools in same org")
	}

	supervisorBindings := []RoleBinding{{Role: RoleSupervisor, TenantID: 7, OrganizationID: 11}}
	if !Can(supervisorBindings, PermissionUseSupervisorTools, Scope{TenantID: 7, OrganizationID: 11}) {
		t.Fatalf("expected supervisor to use supervisor tools")
	}
	if Can(supervisorBindings, PermissionManageOrganization, Scope{OrganizationID: 11}) {
		t.Fatalf("expected supervisor not to manage organization")
	}

	endUserBindings := []RoleBinding{{Role: RoleEndUser, UserID: 42}}
	if !Can(endUserBindings, PermissionManageSelf, Scope{UserID: 42}) {
		t.Fatalf("expected end user to manage self")
	}
	if Can(endUserBindings, PermissionManageSelf, Scope{UserID: 43}) {
		t.Fatalf("expected end user not to manage another user")
	}
	if Can(endUserBindings, PermissionManageUserDomains, Scope{TenantID: 7, UserID: 43}) {
		t.Fatalf("expected end user not to manage another user's domains")
	}
	if !Can(endUserBindings, PermissionManageUserDomains, Scope{TenantID: 7, UserID: 42}) {
		t.Fatalf("expected end user to manage own domains")
	}

	agentBindings := []RoleBinding{{Role: RoleAgent, TenantID: 7, OrganizationID: 11, UserID: 90}}
	if Can(agentBindings, PermissionUseSupervisorTools, Scope{TenantID: 7, OrganizationID: 11}) {
		t.Fatalf("expected agent not to use supervisor tools")
	}
	if Can(agentBindings, Permission("unknown"), Scope{}) {
		t.Fatalf("expected unknown permission to fail")
	}
}

func TestNamespaceRegistry(t *testing.T) {
	registry := NewNamespaceRegistry([]string{"alice"}, []string{"acme"}, []string{"reserved-name"})
	if availabilityError := registry.CheckNameAvailable("alice"); availabilityError == nil {
		t.Fatalf("expected existing user name to be unavailable")
	}
	if availabilityError := registry.CheckNameAvailable("api"); availabilityError == nil {
		t.Fatalf("expected reserved route name to be unavailable")
	}
	if availabilityError := registry.CheckNameAvailable("new-name"); availabilityError != nil {
		t.Fatalf("expected new name to be available, got %v", availabilityError)
	}
	if reserveError := registry.ReserveUser("new-name"); reserveError != nil {
		t.Fatalf("expected reserve user to succeed, got %v", reserveError)
	}
	if reserveError := registry.ReserveUser("api"); reserveError == nil {
		t.Fatalf("expected reserve user to fail for reserved name")
	}
	if reserveError := registry.ReserveOrg("new-name"); reserveError == nil {
		t.Fatalf("expected shared namespace collision after reserving user")
	}

	freshRegistry := NewNamespaceRegistry(nil, nil, nil)
	if reserveError := freshRegistry.ReserveOrg("org-name"); reserveError != nil {
		t.Fatalf("expected reserve org to succeed, got %v", reserveError)
	}
}

func TestTenantResolver(t *testing.T) {
	resolver := NewTenantResolver(
		[]string{"pbx.example.com", "pbx.example.com:443"},
		[]DomainBinding{
			{
				Domain:         "customer.example.com",
				TenantID:       10,
				OrganizationID: 20,
				UserID:         30,
				Status:         model.DomainStatusActive,
			},
			{
				Domain: "inactive.example.com",
				Status: model.DomainStatusSuspended,
			},
		},
	)

	platformContext, resolveError := resolver.Resolve("pbx.example.com")
	if resolveError != nil {
		t.Fatalf("expected platform host to resolve, got %v", resolveError)
	}
	if !platformContext.PlatformHost {
		t.Fatalf("expected platform host context")
	}

	customDomainContext, resolveError := resolver.Resolve("customer.example.com:443")
	if resolveError != nil {
		t.Fatalf("expected active custom domain to resolve, got %v", resolveError)
	}
	if customDomainContext.TenantID != 10 || customDomainContext.OrganizationID != 20 || customDomainContext.UserID != 30 {
		t.Fatalf("unexpected tenant context %+v", customDomainContext)
	}

	if _, resolveError := resolver.Resolve("inactive.example.com"); resolveError != ErrInactiveDomainHost {
		t.Fatalf("expected inactive domain error, got %v", resolveError)
	}
	if _, resolveError := resolver.Resolve("missing.example.com"); resolveError != ErrUnknownTenantHost {
		t.Fatalf("expected unknown tenant host error, got %v", resolveError)
	}
}

func TestPasswordHelpers(t *testing.T) {
	passwordHash, hashError := HashPassword("correct horse battery staple")
	if hashError != nil {
		t.Fatalf("expected password hash, got %v", hashError)
	}
	if !strings.HasPrefix(passwordHash, "$argon2id$") {
		t.Fatalf("unexpected password hash %q", passwordHash)
	}

	verified, rehashHint, verifyError := VerifyPassword(passwordHash, "correct horse battery staple")
	if verifyError != nil || !verified || rehashHint {
		t.Fatalf("expected argon password verify, got %v / %t / %t", verifyError, verified, rehashHint)
	}
	if verified, _, _ = VerifyPassword(passwordHash, "wrong password"); verified {
		t.Fatalf("expected wrong password to fail")
	}
	if verified, _, _ = VerifyPassword(passwordHash, " leading-space"); verified {
		t.Fatalf("expected whitespace password to fail")
	}
	if _, invalidHashError := HashPassword(" trailing-space "); invalidHashError != ErrInvalidPassword {
		t.Fatalf("expected invalid password error, got %v", invalidHashError)
	}

	bcryptHash, bcryptError := bcrypt.GenerateFromPassword([]byte("correct horse battery staple"), 12)
	if bcryptError != nil {
		t.Fatalf("generate bcrypt hash: %v", bcryptError)
	}
	verified, rehashHint, verifyError = VerifyPassword(string(bcryptHash), "correct horse battery staple")
	if verifyError != nil || !verified || !rehashHint {
		t.Fatalf("expected bcrypt verification with rehash hint, got %v / %t / %t", verifyError, verified, rehashHint)
	}
	if verified, _, verifyError = VerifyPassword(string(bcryptHash), "wrong password"); verifyError != nil || verified {
		t.Fatalf("expected bcrypt mismatch to fail cleanly, got %v / %t", verifyError, verified)
	}
	if _, _, verifyError = VerifyPassword("$argon2id$v=19$m=1,t=1,p=1$bad$bad", "correct horse battery staple"); verifyError == nil {
		t.Fatalf("expected invalid argon params to fail")
	}

	if HashToken("abc") != "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad" {
		t.Fatalf("unexpected token hash %q", HashToken("abc"))
	}
}

func TestAuthServiceLifecycle(t *testing.T) {
	memoryStore := store.NewMemoryStore()
	passwordHash, hashError := HashPassword("correct horse battery staple")
	if hashError != nil {
		t.Fatalf("hash password: %v", hashError)
	}
	disabledHash, disabledHashError := HashPassword("disabled password")
	if disabledHashError != nil {
		t.Fatalf("hash disabled password: %v", disabledHashError)
	}

	user, saveUserError := memoryStore.SaveUser(context.Background(), model.User{
		Username:     "alice",
		AccountEmail: "alice@example.com",
		PasswordHash: passwordHash,
		Enabled:      true,
	})
	if saveUserError != nil {
		t.Fatalf("save user: %v", saveUserError)
	}
	if _, saveUserError = memoryStore.SaveUser(context.Background(), model.User{
		Username:     "disabled-user",
		AccountEmail: "disabled@example.com",
		PasswordHash: disabledHash,
		Enabled:      false,
	}); saveUserError != nil {
		t.Fatalf("save disabled user: %v", saveUserError)
	}
	admin, saveAdminError := memoryStore.SaveAdmin(context.Background(), model.Admin{
		Username:     "root-admin",
		AccountEmail: "root@example.com",
		PasswordHash: passwordHash,
		Enabled:      true,
		Role:         model.AdminRoleAdmin,
	})
	if saveAdminError != nil {
		t.Fatalf("save admin: %v", saveAdminError)
	}
	if _, saveAdminError = memoryStore.SaveAdmin(context.Background(), model.Admin{
		Username:     "disabled-admin",
		AccountEmail: "disabled-admin@example.com",
		PasswordHash: disabledHash,
		Enabled:      false,
		Role:         model.AdminRoleAdmin,
	}); saveAdminError != nil {
		t.Fatalf("save disabled admin: %v", saveAdminError)
	}

	baseTime := time.Unix(1_700_000_000, 0)
	authService := NewAuthService(memoryStore, SessionConfig{}).
		WithClock(func() time.Time { return baseTime }).
		WithRandomReader(bytes.NewReader(bytes.Repeat([]byte("abcdefghijklmnopqrstuvwxyz0123456789"), 16)))

	if authService.randomReader == nil {
		t.Fatalf("expected custom random reader")
	}
	if authService.session.AdminTTL != DefaultSessionConfig().AdminTTL || authService.session.UserTTL != DefaultSessionConfig().UserTTL {
		t.Fatalf("expected default session config, got %+v", authService.session)
	}

	authenticatedUser, userSession, authError := authService.AuthenticateUser(context.Background(), "alice@example.com", "correct horse battery staple", "127.0.0.1", "curl/8.0")
	if authError != nil {
		t.Fatalf("authenticate user: %v", authError)
	}
	if authenticatedUser.ID != user.ID || userSession.Token == "" || userSession.TokenHash == "" {
		t.Fatalf("unexpected user auth result %+v %+v", authenticatedUser, userSession)
	}
	if _, _, authError = authService.AuthenticateUser(context.Background(), "disabled@example.com", "disabled password", "127.0.0.1", "curl/8.0"); authError != ErrInvalidCredentials {
		t.Fatalf("expected disabled user auth failure, got %v", authError)
	}
	if _, _, authError = authService.AuthenticateUser(context.Background(), "12345", "correct horse battery staple", "127.0.0.1", "curl/8.0"); authError != ErrInvalidCredentials {
		t.Fatalf("expected numeric identifier auth failure, got %v", authError)
	}
	if _, _, authError = authService.AuthenticateUser(context.Background(), "missing-user", "correct horse battery staple", "127.0.0.1", "curl/8.0"); authError != ErrInvalidCredentials {
		t.Fatalf("expected missing username auth failure, got %v", authError)
	}

	resolvedUserSession, resolveError := authService.ResolveUserSession(context.Background(), userSession.Token)
	if resolveError != nil || resolvedUserSession.SubjectID != user.ID {
		t.Fatalf("expected resolved user session, got %v / %+v", resolveError, resolvedUserSession)
	}
	if resolvedUserSession.ExpiresAt != baseTime.Add(DefaultSessionConfig().UserTTL) {
		t.Fatalf("expected user session expiry to be extended to default ttl, got %v", resolvedUserSession.ExpiresAt)
	}

	foundUser, findUserError := authService.FindUserByID(context.Background(), user.ID)
	if findUserError != nil || foundUser.Username != "alice" {
		t.Fatalf("expected user lookup, got %v / %+v", findUserError, foundUser)
	}
	if logoutError := authService.LogoutUser(context.Background(), userSession.Token); logoutError != nil {
		t.Fatalf("logout user: %v", logoutError)
	}
	if logoutError := authService.LogoutUser(context.Background(), ""); logoutError != ErrSessionNotFound {
		t.Fatalf("expected empty user logout to fail, got %v", logoutError)
	}
	if _, resolveError = authService.ResolveUserSession(context.Background(), userSession.Token); resolveError != ErrSessionNotFound {
		t.Fatalf("expected deleted user session to fail, got %v", resolveError)
	}

	authenticatedAdmin, adminSession, adminAuthError := authService.AuthenticateAdmin(context.Background(), "root-admin", "correct horse battery staple", "127.0.0.1", "curl/8.0")
	if adminAuthError != nil {
		t.Fatalf("authenticate admin: %v", adminAuthError)
	}
	if authenticatedAdmin.ID != admin.ID || adminSession.Session.Kind != model.SessionKindAdmin {
		t.Fatalf("unexpected admin auth result %+v %+v", authenticatedAdmin, adminSession)
	}
	if _, _, adminAuthError = authService.AuthenticateAdmin(context.Background(), "disabled-admin", "disabled password", "127.0.0.1", "curl/8.0"); adminAuthError != ErrInvalidCredentials {
		t.Fatalf("expected disabled admin auth failure, got %v", adminAuthError)
	}
	if _, _, adminAuthError = authService.AuthenticateAdmin(context.Background(), "missing-admin", "correct horse battery staple", "127.0.0.1", "curl/8.0"); adminAuthError != ErrInvalidCredentials {
		t.Fatalf("expected missing admin auth failure, got %v", adminAuthError)
	}

	resolvedAdminSession, resolveAdminError := authService.ResolveAdminSession(context.Background(), adminSession.Token)
	if resolveAdminError != nil || resolvedAdminSession.SubjectID != admin.ID {
		t.Fatalf("expected resolved admin session, got %v / %+v", resolveAdminError, resolvedAdminSession)
	}
	foundAdmin, findAdminError := authService.FindAdminByID(context.Background(), admin.ID)
	if findAdminError != nil || foundAdmin.Username != "root-admin" {
		t.Fatalf("expected admin lookup, got %v / %+v", findAdminError, foundAdmin)
	}
	if logoutError := authService.LogoutAdmin(context.Background(), adminSession.Token); logoutError != nil {
		t.Fatalf("logout admin: %v", logoutError)
	}
	if logoutError := authService.LogoutAdmin(context.Background(), ""); logoutError != ErrSessionNotFound {
		t.Fatalf("expected empty admin logout to fail, got %v", logoutError)
	}
	if _, resolveAdminError = authService.ResolveAdminSession(context.Background(), adminSession.Token); resolveAdminError != ErrSessionNotFound {
		t.Fatalf("expected deleted admin session to fail, got %v", resolveAdminError)
	}

	apiUser, userToken, userTokenError := authService.AuthenticateUserAPI(context.Background(), "alice", "correct horse battery staple")
	if userTokenError != nil || apiUser.ID != user.ID || !strings.HasPrefix(userToken.Value, "usr_") {
		t.Fatalf("expected api user token, got %v / %+v / %+v", userTokenError, apiUser, userToken)
	}
	if _, _, userTokenError = authService.AuthenticateUserAPI(context.Background(), "alice", "wrong"); userTokenError != ErrInvalidCredentials {
		t.Fatalf("expected invalid user api auth failure, got %v", userTokenError)
	}
	resolvedUserToken, resolveUserTokenError := authService.ResolveUserToken(context.Background(), userToken.Value)
	if resolveUserTokenError != nil || resolvedUserToken.OwnerID != user.ID {
		t.Fatalf("expected resolved user token, got %v / %+v", resolveUserTokenError, resolvedUserToken)
	}
	refreshedUserToken, refreshUserTokenError := authService.RefreshUserToken(context.Background(), userToken.Value)
	if refreshUserTokenError != nil || refreshedUserToken.Value == userToken.Value || !strings.HasPrefix(refreshedUserToken.Value, "usr_") {
		t.Fatalf("expected refreshed user token, got %v / %+v", refreshUserTokenError, refreshedUserToken)
	}
	if _, resolveUserTokenError = authService.ResolveUserToken(context.Background(), userToken.Value); resolveUserTokenError != ErrTokenNotFound {
		t.Fatalf("expected old user token to be revoked, got %v", resolveUserTokenError)
	}
	if logoutTokenError := authService.LogoutUserToken(context.Background(), refreshedUserToken.Value); logoutTokenError != nil {
		t.Fatalf("logout user token: %v", logoutTokenError)
	}
	if logoutTokenError := authService.LogoutUserToken(context.Background(), ""); logoutTokenError != ErrTokenNotFound {
		t.Fatalf("expected empty user token logout to fail, got %v", logoutTokenError)
	}

	apiAdmin, adminToken, adminTokenError := authService.AuthenticateAdminAPI(context.Background(), "root-admin", "correct horse battery staple")
	if adminTokenError != nil || apiAdmin.ID != admin.ID || !strings.HasPrefix(adminToken.Value, "adm_") {
		t.Fatalf("expected api admin token, got %v / %+v / %+v", adminTokenError, apiAdmin, adminToken)
	}
	if _, _, adminTokenError = authService.AuthenticateAdminAPI(context.Background(), "missing-admin", "correct horse battery staple"); adminTokenError != ErrInvalidCredentials {
		t.Fatalf("expected missing admin api auth failure, got %v", adminTokenError)
	}
	resolvedAdminToken, resolveAdminTokenError := authService.ResolveAdminToken(context.Background(), adminToken.Value)
	if resolveAdminTokenError != nil || resolvedAdminToken.OwnerID != admin.ID {
		t.Fatalf("expected resolved admin token, got %v / %+v", resolveAdminTokenError, resolvedAdminToken)
	}
	refreshedAdminToken, refreshAdminTokenError := authService.RefreshAdminToken(context.Background(), adminToken.Value)
	if refreshAdminTokenError != nil || !strings.HasPrefix(refreshedAdminToken.Value, "adm_") {
		t.Fatalf("expected refreshed admin token, got %v / %+v", refreshAdminTokenError, refreshedAdminToken)
	}
	if _, resolveAdminTokenError = authService.ResolveAdminToken(context.Background(), adminToken.Value); resolveAdminTokenError != ErrTokenNotFound {
		t.Fatalf("expected old admin token to be revoked, got %v", resolveAdminTokenError)
	}
	if logoutTokenError := authService.LogoutAdminToken(context.Background(), refreshedAdminToken.Value); logoutTokenError != nil {
		t.Fatalf("logout admin token: %v", logoutTokenError)
	}
	if logoutTokenError := authService.LogoutAdminToken(context.Background(), ""); logoutTokenError != ErrTokenNotFound {
		t.Fatalf("expected empty admin token logout to fail, got %v", logoutTokenError)
	}

	orgTokenValue := "org_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	if _, saveTokenError := memoryStore.SaveToken(context.Background(), model.Token{
		OwnerType:   model.TokenOwnerOrg,
		OwnerID:     44,
		Name:        "default",
		TokenHash:   HashToken(orgTokenValue),
		TokenPrefix: "org_aaaa",
		Scope:       model.TokenScopeGlobal,
		ExpiresAt:   baseTime.Add(DefaultSessionConfig().UserTTL),
	}); saveTokenError != nil {
		t.Fatalf("save org token: %v", saveTokenError)
	}
	resolvedOrgToken, resolveOrgTokenError := authService.ResolveOrgToken(context.Background(), orgTokenValue)
	if resolveOrgTokenError != nil || resolvedOrgToken.OwnerID != 44 {
		t.Fatalf("expected resolved org token, got %v / %+v", resolveOrgTokenError, resolvedOrgToken)
	}

	expiringStore := store.NewMemoryStore()
	expiringHash, expiringHashError := HashPassword("correct horse battery staple")
	if expiringHashError != nil {
		t.Fatalf("hash expiring password: %v", expiringHashError)
	}
	expiringUser, expiringSaveError := expiringStore.SaveUser(context.Background(), model.User{
		Username:     "bob",
		AccountEmail: "bob@example.com",
		PasswordHash: expiringHash,
		Enabled:      true,
	})
	if expiringSaveError != nil {
		t.Fatalf("save expiring user: %v", expiringSaveError)
	}

	expireTime := baseTime
	expiringService := NewAuthService(expiringStore, SessionConfig{
		AdminTTL:         time.Hour,
		UserTTL:          time.Hour,
		ExtendOnActivity: false,
	}).WithClock(func() time.Time { return expireTime })
	expiringSessionUser, expiringSession, expiringAuthError := expiringService.AuthenticateUser(context.Background(), "bob", "correct horse battery staple", "127.0.0.1", "curl/8.0")
	if expiringAuthError != nil || expiringSessionUser.ID != expiringUser.ID {
		t.Fatalf("expected expiring auth success, got %v / %+v", expiringAuthError, expiringSessionUser)
	}
	expireTime = baseTime.Add(2 * time.Hour)
	if _, resolveExpiredError := expiringService.ResolveUserSession(context.Background(), expiringSession.Token); resolveExpiredError != ErrSessionExpired {
		t.Fatalf("expected expired session error, got %v", resolveExpiredError)
	}
	expireTime = baseTime
	expiringTokenUser, expiringToken, expiringTokenError := expiringService.AuthenticateUserAPI(context.Background(), "bob", "correct horse battery staple")
	if expiringTokenError != nil || expiringTokenUser.ID != expiringUser.ID {
		t.Fatalf("expected expiring api token auth success, got %v / %+v", expiringTokenError, expiringTokenUser)
	}
	expireTime = baseTime.Add(2 * time.Hour)
	if _, resolveExpiredError := expiringService.ResolveUserToken(context.Background(), expiringToken.Value); resolveExpiredError != ErrTokenExpired {
		t.Fatalf("expected expired token error, got %v", resolveExpiredError)
	}
	if _, resolveExpiredError := expiringService.ResolveAdminSession(context.Background(), ""); resolveExpiredError != ErrSessionNotFound {
		t.Fatalf("expected empty admin resolve failure, got %v", resolveExpiredError)
	}
	if _, resolveExpiredError := expiringService.ResolveAdminToken(context.Background(), ""); resolveExpiredError != ErrInvalidTokenFormat {
		t.Fatalf("expected empty admin token resolve failure, got %v", resolveExpiredError)
	}
}

func TestAuthServiceErrorPathsAndInternals(t *testing.T) {
	previousSaltReader := passwordSaltReader
	passwordSaltReader = io.LimitReader(bytes.NewReader(nil), 0)
	if _, hashError := HashPassword("valid-password"); hashError == nil {
		t.Fatalf("expected password salt reader error")
	}
	passwordSaltReader = previousSaltReader

	if _, _, verifyError := VerifyPassword("$2a$", "correct horse battery staple"); verifyError == nil {
		t.Fatalf("expected invalid bcrypt hash error")
	}
	if _, _, verifyError := VerifyPassword("$argon2id$v=19$m=65536,t=3,p=4$***$bad", "correct horse battery staple"); verifyError == nil {
		t.Fatalf("expected invalid argon salt error")
	}
	if _, _, verifyError := VerifyPassword("$argon2id$v=19$m=65536,t=3,p=4$YWJjZA$***", "correct horse battery staple"); verifyError == nil {
		t.Fatalf("expected invalid argon hash error")
	}

	passwordHash, hashError := HashPassword("correct horse battery staple")
	if hashError != nil {
		t.Fatalf("hash password: %v", hashError)
	}

	saveError := errors.New("save failed")
	sessionError := errors.New("session failed")
	findError := errors.New("find failed")
	validAdmin := model.Admin{ID: 7, Username: "root-admin", PasswordHash: passwordHash, Enabled: true}
	validUser := model.User{ID: 9, Username: "alice", AccountEmail: "alice@example.com", PasswordHash: passwordHash, Enabled: true}

	if _, _, authError := NewAuthService(failingAuthStore{admin: validAdmin, saveAdminError: saveError}, DefaultSessionConfig()).AuthenticateAdmin(context.Background(), "root-admin", "correct horse battery staple", "", ""); authError != saveError {
		t.Fatalf("expected save admin error, got %v", authError)
	}
	if _, _, authError := NewAuthService(failingAuthStore{user: validUser, saveUserError: saveError}, DefaultSessionConfig()).AuthenticateUser(context.Background(), "alice", "correct horse battery staple", "", ""); authError != saveError {
		t.Fatalf("expected save user error, got %v", authError)
	}
	if _, _, authError := NewAuthService(failingAuthStore{admin: model.Admin{Username: "bad", PasswordHash: "invalid", Enabled: true}}, DefaultSessionConfig()).AuthenticateAdmin(context.Background(), "bad", "correct horse battery staple", "", ""); authError != ErrInvalidCredentials {
		t.Fatalf("expected invalid admin credentials, got %v", authError)
	}
	if _, _, authError := NewAuthService(failingAuthStore{user: model.User{Username: "bad", PasswordHash: "invalid", Enabled: true}}, DefaultSessionConfig()).AuthenticateUser(context.Background(), "bad", "correct horse battery staple", "", ""); authError != ErrInvalidCredentials {
		t.Fatalf("expected invalid user credentials, got %v", authError)
	}
	if _, _, authError := NewAuthService(failingAuthStore{admin: validAdmin, saveSessionError: sessionError}, DefaultSessionConfig()).AuthenticateAdmin(context.Background(), "root-admin", "correct horse battery staple", "", ""); authError != sessionError {
		t.Fatalf("expected session creation error for admin, got %v", authError)
	}
	if _, _, authError := NewAuthService(failingAuthStore{user: validUser, saveSessionError: sessionError}, DefaultSessionConfig()).AuthenticateUser(context.Background(), "alice", "correct horse battery staple", "", ""); authError != sessionError {
		t.Fatalf("expected session creation error for user, got %v", authError)
	}
	if _, _, authError := NewAuthService(failingAuthStore{admin: validAdmin, saveTokenError: saveError}, DefaultSessionConfig()).AuthenticateAdminAPI(context.Background(), "root-admin", "correct horse battery staple"); authError != saveError {
		t.Fatalf("expected api admin token create error, got %v", authError)
	}
	if _, _, authError := NewAuthService(failingAuthStore{user: validUser, saveTokenError: saveError}, DefaultSessionConfig()).AuthenticateUserAPI(context.Background(), "alice", "correct horse battery staple"); authError != saveError {
		t.Fatalf("expected api user token create error, got %v", authError)
	}
	if _, findAdminError := NewAuthService(failingAuthStore{findAdminError: findError}, DefaultSessionConfig()).FindAdminByID(context.Background(), 1); findAdminError != findError {
		t.Fatalf("expected find admin error, got %v", findAdminError)
	}
	if _, findUserError := NewAuthService(failingAuthStore{findUserError: findError}, DefaultSessionConfig()).FindUserByID(context.Background(), 1); findUserError != findError {
		t.Fatalf("expected find user error, got %v", findUserError)
	}
	if _, createError := NewAuthService(failingAuthStore{}, DefaultSessionConfig()).WithRandomReader(bytes.NewReader(nil)).createSession(context.Background(), model.SessionKindUser, 1, time.Hour, "", ""); createError == nil {
		t.Fatalf("expected create session token error")
	}
	if _, createError := NewAuthService(failingAuthStore{}, DefaultSessionConfig()).WithRandomReader(bytes.NewReader(nil)).createToken(context.Background(), model.TokenOwnerUser, 1, "", model.TokenScopeGlobal, time.Hour); createError == nil {
		t.Fatalf("expected create api token error")
	}
	if _, createError := NewAuthService(failingAuthStore{}, DefaultSessionConfig()).createToken(context.Background(), "unknown", 1, "", model.TokenScopeGlobal, time.Hour); createError != ErrUnknownTokenType {
		t.Fatalf("expected unknown token type error, got %v", createError)
	}

	resolveFailureService := NewAuthService(failingAuthStore{
		session: model.Session{ID: "s1", Kind: model.SessionKindUser, SubjectID: 1, TokenHash: HashToken("token"), ExpiresAt: time.Unix(2_000_000_000, 0)},
		saveSessionError: saveError,
	}, DefaultSessionConfig()).WithClock(func() time.Time { return time.Unix(1_900_000_000, 0) })
	if _, resolveError := resolveFailureService.ResolveUserSession(context.Background(), "token"); resolveError != saveError {
		t.Fatalf("expected resolve save error, got %v", resolveError)
	}
	resolvedSession, resolveError := NewAuthService(updatingSessionStore{
		found: model.Session{ID: "s1", Kind: model.SessionKindUser, SubjectID: 1, TokenHash: HashToken("token"), ExpiresAt: time.Unix(2_000_000_000, 0)},
		saved: model.Session{ID: "s2", Kind: model.SessionKindUser, SubjectID: 1, TokenHash: HashToken("token"), ExpiresAt: time.Unix(2_100_000_000, 0)},
	}, DefaultSessionConfig()).WithClock(func() time.Time { return time.Unix(1_900_000_000, 0) }).ResolveUserSession(context.Background(), "token")
	if resolveError != nil || resolvedSession.ID != "s2" {
		t.Fatalf("expected resolved session from saved record, got %v / %+v", resolveError, resolvedSession)
	}
	resolveTokenFailureService := NewAuthService(failingAuthStore{
		token: model.Token{ID: 1, OwnerType: model.TokenOwnerUser, OwnerID: 1, TokenHash: HashToken("usr_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"), TokenPrefix: "usr_aaaa", ExpiresAt: time.Unix(2_000_000_000, 0)},
		saveTokenError: saveError,
	}, DefaultSessionConfig()).WithClock(func() time.Time { return time.Unix(1_900_000_000, 0) })
	if _, resolveError := resolveTokenFailureService.ResolveUserToken(context.Background(), "usr_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"); resolveError != saveError {
		t.Fatalf("expected resolve token save error, got %v", resolveError)
	}
	resolvedToken, resolveError := NewAuthService(updatingTokenStore{
		found: model.Token{ID: 1, OwnerType: model.TokenOwnerUser, OwnerID: 1, TokenHash: HashToken("usr_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"), TokenPrefix: "usr_aaaa", ExpiresAt: time.Unix(2_000_000_000, 0)},
		saved: model.Token{ID: 2, OwnerType: model.TokenOwnerUser, OwnerID: 1, TokenHash: HashToken("usr_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"), TokenPrefix: "usr_bbbb", ExpiresAt: time.Unix(2_100_000_000, 0)},
	}, DefaultSessionConfig()).WithClock(func() time.Time { return time.Unix(1_900_000_000, 0) }).ResolveUserToken(context.Background(), "usr_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	if resolveError != nil || resolvedToken.ID != 2 {
		t.Fatalf("expected resolved token from saved record, got %v / %+v", resolveError, resolvedToken)
	}
	if _, refreshError := NewAuthService(failingAuthStore{findTokenError: findError}, DefaultSessionConfig()).RefreshUserToken(context.Background(), "usr_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"); refreshError != ErrTokenNotFound {
		t.Fatalf("expected refresh token lookup failure, got %v", refreshError)
	}
	if _, refreshError := NewAuthService(refreshDeleteFailStore{failingAuthStore{
		token: model.Token{ID: 1, OwnerType: model.TokenOwnerUser, OwnerID: 1, Name: "default", Scope: model.TokenScopeGlobal, TokenHash: HashToken("usr_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"), TokenPrefix: "usr_aaaa", ExpiresAt: time.Unix(2_000_000_000, 0)},
		deleteTokenError: findError,
	}}, DefaultSessionConfig()).WithClock(func() time.Time { return time.Unix(1_900_000_000, 0) }).RefreshUserToken(context.Background(), "usr_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"); refreshError != findError {
		t.Fatalf("expected refresh token delete failure, got %v", refreshError)
	}
	if _, refreshError := NewAuthService(failingAuthStore{
		token: model.Token{ID: 1, OwnerType: model.TokenOwnerUser, OwnerID: 1, Name: "default", Scope: model.TokenScopeGlobal, TokenHash: HashToken("usr_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"), TokenPrefix: "usr_aaaa", ExpiresAt: time.Unix(2_000_000_000, 0)},
	}, DefaultSessionConfig()).WithClock(func() time.Time { return time.Unix(1_900_000_000, 0) }).RefreshUserToken(context.Background(), "usr_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"); refreshError != ErrTokenNotFound {
		t.Fatalf("expected refresh token collision failure, got %v", refreshError)
	}
	if _, refreshError := NewAuthService(refreshCreateFailStore{
		token: model.Token{ID: 1, OwnerType: model.TokenOwnerUser, OwnerID: 1, Name: "default", Scope: model.TokenScopeGlobal, TokenHash: HashToken("usr_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"), TokenPrefix: "usr_aaaa", ExpiresAt: time.Unix(2_000_000_000, 0)},
	}, DefaultSessionConfig()).WithClock(func() time.Time { return time.Unix(1_900_000_000, 0) }).RefreshUserToken(context.Background(), "usr_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"); refreshError == nil || refreshError.Error() != "create failed" {
		t.Fatalf("expected refresh token create failure, got %v", refreshError)
	}
	if logoutTokenError := NewAuthService(failingAuthStore{deleteTokenError: findError}, DefaultSessionConfig()).LogoutUserToken(context.Background(), "usr_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"); logoutTokenError != ErrTokenNotFound {
		t.Fatalf("expected api token delete failure, got %v", logoutTokenError)
	}
	if _, resolveError := NewAuthService(failingAuthStore{findTokenError: findError}, DefaultSessionConfig()).ResolveUserToken(context.Background(), "usr_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"); resolveError != ErrTokenNotFound {
		t.Fatalf("expected api token lookup failure, got %v", resolveError)
	}
	if _, resolveError := NewAuthService(failingAuthStore{}, DefaultSessionConfig()).ResolveUserToken(context.Background(), "invalid"); resolveError != ErrInvalidTokenFormat {
		t.Fatalf("expected invalid token format, got %v", resolveError)
	}
	if logoutError := NewAuthService(failingAuthStore{deleteError: findError}, DefaultSessionConfig()).LogoutUser(context.Background(), "token"); logoutError != ErrSessionNotFound {
		t.Fatalf("expected logout delete failure, got %v", logoutError)
	}
	if _, _, decodeError := decodeArgon2Hash("$argon2id$v=18$m=65536,t=3,p=4$YWJjZA$YWJjZA"); decodeError == nil {
		t.Fatalf("expected invalid version decode error")
	}
	if _, opaqueError := generateOpaqueToken(bytes.NewReader(nil), 4); opaqueError == nil {
		t.Fatalf("expected opaque token generation failure")
	}
	if _, opaqueError := generatePrefixedOpaqueToken(bytes.NewReader(nil), "usr_", 4); opaqueError == nil {
		t.Fatalf("expected prefixed opaque token generation failure")
	}
	if token, tokenError := generatePrefixedOpaqueToken(bytes.NewReader(bytes.Repeat([]byte{1}, 32)), "usr_", 4); tokenError != nil || !strings.HasPrefix(token, "usr_") || len(token) != 8 {
		t.Fatalf("expected generated prefixed token, got %v / %q", tokenError, token)
	}
	if defaultTokenName("") != "default" || defaultTokenName(" ci ") != "ci" {
		t.Fatalf("unexpected token name defaults")
	}
	if defaultTokenScope("") != model.TokenScopeGlobal || defaultTokenScope(model.TokenScopeRead) != model.TokenScopeRead {
		t.Fatalf("unexpected token scope defaults")
	}
	if tokenDisplayPrefix("short") != "short" || tokenDisplayPrefix("usr_abcdefgh") != "usr_abcd" {
		t.Fatalf("unexpected token display prefix values")
	}
	if prefix, prefixError := tokenPrefix(model.TokenOwnerAdmin); prefixError != nil || prefix != "adm_" {
		t.Fatalf("expected admin token prefix, got %v / %q", prefixError, prefix)
	}
	if prefix, prefixError := tokenPrefix(model.TokenOwnerOrg); prefixError != nil || prefix != "org_" {
		t.Fatalf("expected org token prefix, got %v / %q", prefixError, prefix)
	}
	if prefix, prefixError := tokenPrefix("nope"); prefixError != ErrUnknownTokenType || prefix != "" {
		t.Fatalf("expected unknown token type prefix failure, got %v / %q", prefixError, prefix)
	}
	if validateError := validateBearerTokenValue("usr_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", model.TokenOwnerUser); validateError != nil {
		t.Fatalf("expected valid bearer token format, got %v", validateError)
	}
	if validateError := validateBearerTokenValue("adm_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", model.TokenOwnerUser); validateError != ErrInvalidTokenFormat {
		t.Fatalf("expected wrong token prefix failure, got %v", validateError)
	}
	if validateError := validateBearerTokenValue("org_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", "unknown"); validateError != ErrUnknownTokenType {
		t.Fatalf("expected unknown token type validation failure, got %v", validateError)
	}
	if validateError := validateBearerTokenValue("usr_AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA", model.TokenOwnerUser); validateError != ErrInvalidTokenFormat {
		t.Fatalf("expected invalid token chars, got %v", validateError)
	}
	if token, tokenError := generatePrefixedOpaqueToken(bytes.NewReader(nil), "usr_", 0); tokenError != nil || token != "usr_" {
		t.Fatalf("expected empty-length prefixed token, got %v / %q", tokenError, token)
	}
	func() {
		defer func() {
			if recover() == nil {
				t.Fatalf("expected uuid generation panic")
			}
		}()
		_ = generateUUIDv4(bytes.NewReader(nil))
	}()
}
