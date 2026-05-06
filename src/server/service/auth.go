package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/bcrypt"

	"github.com/casapps/caspbx/src/server/model"
	"github.com/casapps/caspbx/src/server/store"
)

type Role string
type Permission string

const (
	RoleServerAdmin       Role = "server_admin"
	RoleTenantAdmin       Role = "tenant_admin"
	RoleOrganizationAdmin Role = "organization_admin"
	RoleSupervisor        Role = "supervisor"
	RoleAgent             Role = "agent"
	RoleEndUser           Role = "end_user"

	PermissionManagePlatform     Permission = "manage_platform"
	PermissionManageTenant       Permission = "manage_tenant"
	PermissionManageOrganization Permission = "manage_organization"
	PermissionUseSupervisorTools Permission = "use_supervisor_tools"
	PermissionManageSelf         Permission = "manage_self"
	PermissionManageUserDomains  Permission = "manage_user_domains"
	PermissionManageOrgDomains   Permission = "manage_org_domains"
)

const (
	ArgonTime    uint32 = 3
	ArgonMemory  uint32 = 64 * 1024
	ArgonThreads uint8  = 4
	ArgonKeyLen  uint32 = 32
	ArgonSaltLen int    = 16
	SessionBytes int    = 32
	TokenBytes   int    = 32
)

var (
	ErrInvalidCredentials     = errors.New("invalid credentials")
	ErrInvalidPassword        = errors.New("invalid password")
	ErrSessionNotFound        = errors.New("session not found")
	ErrSessionExpired         = errors.New("session expired")
	ErrTokenNotFound          = errors.New("token not found")
	ErrTokenExpired           = errors.New("token expired")
	ErrInvalidTokenFormat     = errors.New("invalid token format")
	ErrUnknownTokenType       = errors.New("unknown token type")
	ErrRegistrationRestricted = errors.New("registration requires invitation")
	passwordSaltReader        io.Reader = rand.Reader
)

type RoleBinding struct {
	Role           Role
	TenantID       int64
	OrganizationID int64
	UserID         int64
}

type Scope struct {
	TenantID       int64
	OrganizationID int64
	UserID         int64
}

type SessionConfig struct {
	AdminTTL         time.Duration
	UserTTL          time.Duration
	ExtendOnActivity bool
}

type LoginSession struct {
	Session    model.Session
	Token      string
	TokenHash  string
	RehashHint bool
}

type IssuedToken struct {
	Token      model.Token
	Value      string
	RehashHint bool
}

type AuthService struct {
	store        store.AuthStore
	randomReader io.Reader
	now          func() time.Time
	session      SessionConfig
}

func Can(bindings []RoleBinding, permission Permission, scope Scope) bool {
	for _, binding := range bindings {
		if binding.Role == RoleServerAdmin {
			return true
		}

		switch permission {
		case PermissionManagePlatform:
			continue

		case PermissionManageTenant:
			if binding.Role == RoleTenantAdmin && binding.TenantID == scope.TenantID {
				return true
			}

		case PermissionManageOrganization:
			if binding.Role == RoleTenantAdmin && binding.TenantID == scope.TenantID {
				return true
			}
			if binding.Role == RoleOrganizationAdmin && binding.OrganizationID == scope.OrganizationID {
				return true
			}

		case PermissionUseSupervisorTools:
			if binding.Role == RoleTenantAdmin && binding.TenantID == scope.TenantID {
				return true
			}
			if binding.Role == RoleOrganizationAdmin && binding.OrganizationID == scope.OrganizationID {
				return true
			}
			if binding.Role == RoleSupervisor && binding.TenantID == scope.TenantID {
				return true
			}

		case PermissionManageSelf:
			if binding.UserID != 0 && binding.UserID == scope.UserID {
				return true
			}

		case PermissionManageUserDomains:
			if binding.Role == RoleTenantAdmin && binding.TenantID == scope.TenantID {
				return true
			}
			if binding.UserID != 0 && binding.UserID == scope.UserID {
				return true
			}

		case PermissionManageOrgDomains:
			if binding.Role == RoleTenantAdmin && binding.TenantID == scope.TenantID {
				return true
			}
			if binding.Role == RoleOrganizationAdmin && binding.OrganizationID == scope.OrganizationID {
				return true
			}
		}
	}

	return false
}

func DefaultSessionConfig() SessionConfig {
	return SessionConfig{
		AdminTTL:         30 * 24 * time.Hour,
		UserTTL:          7 * 24 * time.Hour,
		ExtendOnActivity: true,
	}
}

func NewAuthService(authStore store.AuthStore, sessionConfig SessionConfig) AuthService {
	if sessionConfig.AdminTTL <= 0 {
		sessionConfig.AdminTTL = DefaultSessionConfig().AdminTTL
	}
	if sessionConfig.UserTTL <= 0 {
		sessionConfig.UserTTL = DefaultSessionConfig().UserTTL
	}

	return AuthService{
		store:        authStore,
		randomReader: rand.Reader,
		now:          time.Now,
		session:      sessionConfig,
	}
}

func (service AuthService) WithClock(now func() time.Time) AuthService {
	service.now = now
	return service
}

func (service AuthService) WithRandomReader(reader io.Reader) AuthService {
	service.randomReader = reader
	return service
}

func HashPassword(password string) (string, error) {
	if password == "" || strings.TrimSpace(password) != password {
		return "", ErrInvalidPassword
	}

	salt := make([]byte, ArgonSaltLen)
	if _, readError := io.ReadFull(passwordSaltReader, salt); readError != nil {
		return "", readError
	}
	return encodeArgon2Hash(salt, argon2.IDKey([]byte(password), salt, ArgonTime, ArgonMemory, ArgonThreads, ArgonKeyLen)), nil
}

func VerifyPassword(passwordHash string, password string) (bool, bool, error) {
	if password == "" || strings.TrimSpace(password) != password {
		return false, false, nil
	}

	if strings.HasPrefix(passwordHash, "$2") {
		verifyError := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password))
		if verifyError == nil {
			return true, true, nil
		}
		if errors.Is(verifyError, bcrypt.ErrMismatchedHashAndPassword) {
			return false, false, nil
		}
		return false, false, verifyError
	}

	salt, expectedHash, parseError := decodeArgon2Hash(passwordHash)
	if parseError != nil {
		return false, false, parseError
	}

	actualHash := argon2.IDKey([]byte(password), salt, ArgonTime, ArgonMemory, ArgonThreads, uint32(len(expectedHash)))
	return subtle.ConstantTimeCompare(actualHash, expectedHash) == 1, false, nil
}

func HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

func (service AuthService) AuthenticateAdmin(ctx context.Context, username string, password string, ipAddress string, userAgent string) (model.Admin, LoginSession, error) {
	admin, rehashHint, authError := service.authenticateAdminCredentials(ctx, username, password)
	if authError != nil {
		return model.Admin{}, LoginSession{}, authError
	}
	session, sessionError := service.createSession(ctx, model.SessionKindAdmin, admin.ID, service.session.AdminTTL, ipAddress, userAgent)
	if sessionError != nil {
		return model.Admin{}, LoginSession{}, sessionError
	}
	session.RehashHint = rehashHint
	return admin, session, nil
}

func (service AuthService) AuthenticateUser(ctx context.Context, identifier string, password string, ipAddress string, userAgent string) (model.User, LoginSession, error) {
	user, rehashHint, authError := service.authenticateUserCredentials(ctx, identifier, password)
	if authError != nil {
		return model.User{}, LoginSession{}, authError
	}
	session, sessionError := service.createSession(ctx, model.SessionKindUser, user.ID, service.session.UserTTL, ipAddress, userAgent)
	if sessionError != nil {
		return model.User{}, LoginSession{}, sessionError
	}
	session.RehashHint = rehashHint
	return user, session, nil
}

func (service AuthService) AuthenticateAdminAPI(ctx context.Context, username string, password string) (model.Admin, IssuedToken, error) {
	admin, rehashHint, authError := service.authenticateAdminCredentials(ctx, username, password)
	if authError != nil {
		return model.Admin{}, IssuedToken{}, authError
	}
	token, tokenError := service.createToken(ctx, model.TokenOwnerAdmin, admin.ID, "default", model.TokenScopeGlobal, service.session.AdminTTL)
	if tokenError != nil {
		return model.Admin{}, IssuedToken{}, tokenError
	}
	token.RehashHint = rehashHint
	return admin, token, nil
}

func (service AuthService) AuthenticateUserAPI(ctx context.Context, identifier string, password string) (model.User, IssuedToken, error) {
	user, rehashHint, authError := service.authenticateUserCredentials(ctx, identifier, password)
	if authError != nil {
		return model.User{}, IssuedToken{}, authError
	}
	token, tokenError := service.createToken(ctx, model.TokenOwnerUser, user.ID, "default", model.TokenScopeGlobal, service.session.UserTTL)
	if tokenError != nil {
		return model.User{}, IssuedToken{}, tokenError
	}
	token.RehashHint = rehashHint
	return user, token, nil
}

func (service AuthService) ResolveAdminSession(ctx context.Context, token string) (model.Session, error) {
	return service.resolveSession(ctx, model.SessionKindAdmin, token, service.session.AdminTTL)
}

func (service AuthService) ResolveUserSession(ctx context.Context, token string) (model.Session, error) {
	return service.resolveSession(ctx, model.SessionKindUser, token, service.session.UserTTL)
}

func (service AuthService) FindAdminByID(ctx context.Context, id int64) (model.Admin, error) {
	admin, lookupError := service.store.FindAdminByID(ctx, id)
	if lookupError != nil {
		return model.Admin{}, lookupError
	}
	return admin, nil
}

func (service AuthService) FindUserByID(ctx context.Context, id int64) (model.User, error) {
	user, lookupError := service.store.FindUserByID(ctx, id)
	if lookupError != nil {
		return model.User{}, lookupError
	}
	return user, nil
}

func (service AuthService) ResolveAdminToken(ctx context.Context, token string) (model.Token, error) {
	return service.resolveToken(ctx, model.TokenOwnerAdmin, token)
}

func (service AuthService) ResolveUserToken(ctx context.Context, token string) (model.Token, error) {
	return service.resolveToken(ctx, model.TokenOwnerUser, token)
}

func (service AuthService) ResolveOrgToken(ctx context.Context, token string) (model.Token, error) {
	return service.resolveToken(ctx, model.TokenOwnerOrg, token)
}

func (service AuthService) LogoutAdmin(ctx context.Context, token string) error {
	return service.logout(ctx, model.SessionKindAdmin, token)
}

func (service AuthService) LogoutUser(ctx context.Context, token string) error {
	return service.logout(ctx, model.SessionKindUser, token)
}

func (service AuthService) LogoutAdminToken(ctx context.Context, token string) error {
	return service.logoutToken(ctx, model.TokenOwnerAdmin, token)
}

func (service AuthService) LogoutUserToken(ctx context.Context, token string) error {
	return service.logoutToken(ctx, model.TokenOwnerUser, token)
}

func (service AuthService) RefreshAdminToken(ctx context.Context, token string) (IssuedToken, error) {
	return service.refreshToken(ctx, model.TokenOwnerAdmin, token, service.session.AdminTTL)
}

func (service AuthService) RefreshUserToken(ctx context.Context, token string) (IssuedToken, error) {
	return service.refreshToken(ctx, model.TokenOwnerUser, token, service.session.UserTTL)
}

func (service AuthService) createSession(ctx context.Context, kind model.SessionKind, subjectID int64, ttl time.Duration, ipAddress string, userAgent string) (LoginSession, error) {
	token, tokenError := generateOpaqueToken(service.randomReader, SessionBytes)
	if tokenError != nil {
		return LoginSession{}, tokenError
	}

	now := service.now()
	session := model.Session{
		ID:         generateUUIDv4(service.randomReader),
		Kind:       kind,
		SubjectID:  subjectID,
		TokenHash:  HashToken(token),
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
		CreatedAt:  now,
		ExpiresAt:  now.Add(ttl),
		LastSeenAt: now,
	}

	savedSession, saveError := service.store.SaveSession(ctx, session)
	if saveError != nil {
		return LoginSession{}, saveError
	}

	return LoginSession{
		Session:   savedSession,
		Token:     token,
		TokenHash: savedSession.TokenHash,
	}, nil
}

func (service AuthService) resolveSession(ctx context.Context, kind model.SessionKind, token string, ttl time.Duration) (model.Session, error) {
	if strings.TrimSpace(token) == "" {
		return model.Session{}, ErrSessionNotFound
	}

	session, lookupError := service.store.FindSessionByTokenHash(ctx, kind, HashToken(token))
	if lookupError != nil {
		return model.Session{}, ErrSessionNotFound
	}
	if session.IsExpired(service.now()) {
		_ = service.store.DeleteSessionByTokenHash(ctx, kind, session.TokenHash)
		return model.Session{}, ErrSessionExpired
	}
	if service.session.ExtendOnActivity {
		now := service.now()
		session.LastSeenAt = now
		session.ExpiresAt = now.Add(ttl)
		savedSession, saveError := service.store.SaveSession(ctx, session)
		if saveError != nil {
			return model.Session{}, saveError
		}
		session = savedSession
	}
	return session, nil
}

func (service AuthService) logout(ctx context.Context, kind model.SessionKind, token string) error {
	if strings.TrimSpace(token) == "" {
		return ErrSessionNotFound
	}
	deleteError := service.store.DeleteSessionByTokenHash(ctx, kind, HashToken(token))
	if deleteError != nil {
		return ErrSessionNotFound
	}
	return nil
}

func (service AuthService) authenticateAdminCredentials(ctx context.Context, username string, password string) (model.Admin, bool, error) {
	admin, lookupError := service.store.FindAdminByUsername(ctx, strings.TrimSpace(username))
	if lookupError != nil {
		return model.Admin{}, false, ErrInvalidCredentials
	}
	if !admin.Enabled {
		return model.Admin{}, false, ErrInvalidCredentials
	}

	verified, rehashHint, verifyError := VerifyPassword(admin.PasswordHash, password)
	if verifyError != nil || !verified {
		return model.Admin{}, false, ErrInvalidCredentials
	}

	admin.LastLoginAt = service.now()
	if _, saveError := service.store.SaveAdmin(ctx, admin); saveError != nil {
		return model.Admin{}, false, saveError
	}

	return admin, rehashHint, nil
}

func (service AuthService) authenticateUserCredentials(ctx context.Context, identifier string, password string) (model.User, bool, error) {
	var (
		user        model.User
		lookupError error
	)

	switch model.DetectIdentifierType(identifier) {
	case model.IdentifierTypeEmail:
		user, lookupError = service.store.FindUserByEmail(ctx, identifier)
	case model.IdentifierTypeUsername:
		user, lookupError = service.store.FindUserByUsername(ctx, identifier)
	default:
		return model.User{}, false, ErrInvalidCredentials
	}

	if lookupError != nil {
		return model.User{}, false, ErrInvalidCredentials
	}
	if !user.Enabled {
		return model.User{}, false, ErrInvalidCredentials
	}

	verified, rehashHint, verifyError := VerifyPassword(user.PasswordHash, password)
	if verifyError != nil || !verified {
		return model.User{}, false, ErrInvalidCredentials
	}

	user.LastLoginAt = service.now()
	if _, saveError := service.store.SaveUser(ctx, user); saveError != nil {
		return model.User{}, false, saveError
	}

	return user, rehashHint, nil
}

func (service AuthService) createToken(ctx context.Context, ownerType model.TokenOwnerType, ownerID int64, name string, scope model.TokenScope, ttl time.Duration) (IssuedToken, error) {
	prefix, prefixError := tokenPrefix(ownerType)
	if prefixError != nil {
		return IssuedToken{}, prefixError
	}

	value, tokenError := generatePrefixedOpaqueToken(service.randomReader, prefix, TokenBytes)
	if tokenError != nil {
		return IssuedToken{}, tokenError
	}

	now := service.now()
	token := model.Token{
		OwnerType:   ownerType,
		OwnerID:     ownerID,
		Name:        defaultTokenName(name),
		TokenHash:   HashToken(value),
		TokenPrefix: tokenDisplayPrefix(value),
		Scope:       defaultTokenScope(scope),
		CreatedAt:   now,
		LastUsedAt:  now,
	}
	if ttl > 0 {
		token.ExpiresAt = now.Add(ttl)
	}

	savedToken, saveError := service.store.SaveToken(ctx, token)
	if saveError != nil {
		return IssuedToken{}, saveError
	}

	return IssuedToken{
		Token: savedToken,
		Value: value,
	}, nil
}

func (service AuthService) resolveToken(ctx context.Context, ownerType model.TokenOwnerType, token string) (model.Token, error) {
	if validateError := validateBearerTokenValue(token, ownerType); validateError != nil {
		return model.Token{}, validateError
	}

	storedToken, lookupError := service.store.FindTokenByHash(ctx, ownerType, HashToken(token))
	if lookupError != nil {
		return model.Token{}, ErrTokenNotFound
	}
	if storedToken.IsExpired(service.now()) {
		_ = service.store.DeleteTokenByHash(ctx, ownerType, storedToken.TokenHash)
		return model.Token{}, ErrTokenExpired
	}

	storedToken.LastUsedAt = service.now()
	savedToken, saveError := service.store.SaveToken(ctx, storedToken)
	if saveError != nil {
		return model.Token{}, saveError
	}
	return savedToken, nil
}

func (service AuthService) refreshToken(ctx context.Context, ownerType model.TokenOwnerType, token string, ttl time.Duration) (IssuedToken, error) {
	storedToken, resolveError := service.resolveToken(ctx, ownerType, token)
	if resolveError != nil {
		return IssuedToken{}, resolveError
	}

	var (
		refreshedToken IssuedToken
		createError    error
	)
	for attempts := 0; attempts < 4; attempts++ {
		refreshedToken, createError = service.createToken(ctx, ownerType, storedToken.OwnerID, storedToken.Name, storedToken.Scope, ttl)
		if createError != nil {
			return IssuedToken{}, createError
		}
		if refreshedToken.Token.TokenHash != storedToken.TokenHash {
			break
		}
	}
	if refreshedToken.Token.TokenHash == storedToken.TokenHash {
		return IssuedToken{}, ErrTokenNotFound
	}
	if deleteError := service.store.DeleteTokenByHash(ctx, ownerType, storedToken.TokenHash); deleteError != nil {
		_ = service.store.DeleteTokenByHash(ctx, ownerType, refreshedToken.Token.TokenHash)
		return IssuedToken{}, deleteError
	}
	return refreshedToken, nil
}

func (service AuthService) logoutToken(ctx context.Context, ownerType model.TokenOwnerType, token string) error {
	if validateError := validateBearerTokenValue(token, ownerType); validateError != nil {
		return ErrTokenNotFound
	}
	if deleteError := service.store.DeleteTokenByHash(ctx, ownerType, HashToken(token)); deleteError != nil {
		return ErrTokenNotFound
	}
	return nil
}

func defaultTokenName(name string) string {
	if strings.TrimSpace(name) == "" {
		return "default"
	}
	return strings.TrimSpace(name)
}

func defaultTokenScope(scope model.TokenScope) model.TokenScope {
	switch scope {
	case model.TokenScopeRead, model.TokenScopeReadWrite:
		return scope
	default:
		return model.TokenScopeGlobal
	}
}

func tokenDisplayPrefix(token string) string {
	if len(token) <= 8 {
		return token
	}
	return token[:8]
}

func tokenPrefix(ownerType model.TokenOwnerType) (string, error) {
	switch ownerType {
	case model.TokenOwnerAdmin:
		return "adm_", nil
	case model.TokenOwnerUser:
		return "usr_", nil
	case model.TokenOwnerOrg:
		return "org_", nil
	default:
		return "", ErrUnknownTokenType
	}
}

func validateBearerTokenValue(token string, ownerType model.TokenOwnerType) error {
	prefix, prefixError := tokenPrefix(ownerType)
	if prefixError != nil {
		return prefixError
	}

	trimmed := strings.TrimSpace(token)
	if !strings.HasPrefix(trimmed, prefix) || len(trimmed) != len(prefix)+TokenBytes {
		return ErrInvalidTokenFormat
	}
	for _, character := range trimmed[len(prefix):] {
		if (character < 'a' || character > 'z') && (character < '0' || character > '9') {
			return ErrInvalidTokenFormat
		}
	}
	return nil
}

func encodeArgon2Hash(salt []byte, hash []byte) string {
	encoding := base64.RawStdEncoding
	return fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		ArgonMemory,
		ArgonTime,
		ArgonThreads,
		encoding.EncodeToString(salt),
		encoding.EncodeToString(hash),
	)
}

func decodeArgon2Hash(passwordHash string) ([]byte, []byte, error) {
	parts := strings.Split(passwordHash, "$")
	if len(parts) != 6 || parts[1] != "argon2id" || parts[2] != fmt.Sprintf("v=%d", argon2.Version) {
		return nil, nil, ErrInvalidPassword
	}
	if parts[3] != fmt.Sprintf("m=%d,t=%d,p=%d", ArgonMemory, ArgonTime, ArgonThreads) {
		return nil, nil, ErrInvalidPassword
	}

	encoding := base64.RawStdEncoding
	salt, saltError := encoding.DecodeString(parts[4])
	if saltError != nil {
		return nil, nil, saltError
	}
	hash, hashError := encoding.DecodeString(parts[5])
	if hashError != nil {
		return nil, nil, hashError
	}
	return salt, hash, nil
}

func generateOpaqueToken(reader io.Reader, byteCount int) (string, error) {
	value := make([]byte, byteCount)
	if _, readError := io.ReadFull(reader, value); readError != nil {
		return "", readError
	}
	return base64.RawURLEncoding.EncodeToString(value), nil
}

func generatePrefixedOpaqueToken(reader io.Reader, prefix string, length int) (string, error) {
	const alphabet = "abcdefghijklmnopqrstuvwxyz0123456789"
	if length <= 0 {
		return prefix, nil
	}

	result := make([]byte, length)
	randomByte := make([]byte, 1)
	limit := byte(256 - (256 % len(alphabet)))
	for index := range result {
		for {
			if _, readError := io.ReadFull(reader, randomByte); readError != nil {
				return "", readError
			}
			if randomByte[0] >= limit {
				continue
			}
			result[index] = alphabet[int(randomByte[0])%len(alphabet)]
			break
		}
	}

	return prefix + string(result), nil
}

func generateUUIDv4(reader io.Reader) string {
	value := make([]byte, 16)
	if _, readError := io.ReadFull(reader, value); readError != nil {
		panic(readError)
	}
	value[6] = (value[6] & 0x0f) | 0x40
	value[8] = (value[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", value[0:4], value[4:6], value[6:8], value[8:10], value[10:16])
}
