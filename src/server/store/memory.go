package store

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/casapps/caspbx/src/server/model"
)

type MemoryStore struct {
	mu              sync.RWMutex
	nextAdminID     int64
	nextUserID      int64
	nextOrgID       int64
	nextOrgMemberID int64
	nextTokenID     int64
	adminsByID      map[int64]model.Admin
	adminIDsByName  map[string]int64
	usersByID       map[int64]model.User
	userIDsByName   map[string]int64
	userIDsByEmail  map[string]int64
	orgsByID        map[int64]model.Organization
	orgIDsBySlug    map[string]int64
	orgPrefsByOrgID map[int64]model.OrganizationPreferences
	orgMembersByID  map[int64]model.OrganizationMember
	adminSessions   map[string]model.Session
	userSessions    map[string]model.Session
	adminTokens     map[string]model.Token
	userTokens      map[string]model.Token
	orgTokens       map[string]model.Token
	usernames       map[string]struct{}
	orgSlugs        map[string]struct{}
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		nextAdminID:    1,
		nextUserID:     1,
		nextOrgID:      1,
		nextOrgMemberID: 1,
		nextTokenID:    1,
		adminsByID:     map[int64]model.Admin{},
		adminIDsByName: map[string]int64{},
		usersByID:      map[int64]model.User{},
		userIDsByName:  map[string]int64{},
		userIDsByEmail: map[string]int64{},
		orgsByID:       map[int64]model.Organization{},
		orgIDsBySlug:   map[string]int64{},
		orgPrefsByOrgID: map[int64]model.OrganizationPreferences{},
		orgMembersByID: map[int64]model.OrganizationMember{},
		adminSessions:  map[string]model.Session{},
		userSessions:   map[string]model.Session{},
		adminTokens:    map[string]model.Token{},
		userTokens:     map[string]model.Token{},
		orgTokens:      map[string]model.Token{},
		usernames:      map[string]struct{}{},
		orgSlugs:       map[string]struct{}{},
	}
}

func (memoryStore *MemoryStore) UserExistsByName(_ context.Context, username string) (bool, error) {
	memoryStore.mu.RLock()
	defer memoryStore.mu.RUnlock()
	_, exists := memoryStore.usernames[strings.TrimSpace(username)]
	return exists, nil
}

func (memoryStore *MemoryStore) OrgExistsBySlug(_ context.Context, slug string) (bool, error) {
	memoryStore.mu.RLock()
	defer memoryStore.mu.RUnlock()
	_, exists := memoryStore.orgSlugs[strings.TrimSpace(slug)]
	return exists, nil
}

func (memoryStore *MemoryStore) SaveAdmin(_ context.Context, admin model.Admin) (model.Admin, error) {
	memoryStore.mu.Lock()
	defer memoryStore.mu.Unlock()

	admin.Username = strings.TrimSpace(admin.Username)
	if admin.ID == 0 {
		admin.ID = memoryStore.nextAdminID
		memoryStore.nextAdminID++
	}
	memoryStore.adminsByID[admin.ID] = admin
	memoryStore.adminIDsByName[admin.Username] = admin.ID
	return admin, nil
}

func (memoryStore *MemoryStore) FindAdminByUsername(_ context.Context, username string) (model.Admin, error) {
	memoryStore.mu.RLock()
	defer memoryStore.mu.RUnlock()

	adminID, found := memoryStore.adminIDsByName[strings.TrimSpace(username)]
	if !found {
		return model.Admin{}, ErrNotFound
	}
	return memoryStore.adminsByID[adminID], nil
}

func (memoryStore *MemoryStore) FindAdminByID(_ context.Context, id int64) (model.Admin, error) {
	memoryStore.mu.RLock()
	defer memoryStore.mu.RUnlock()

	admin, found := memoryStore.adminsByID[id]
	if !found {
		return model.Admin{}, ErrNotFound
	}
	return admin, nil
}

func (memoryStore *MemoryStore) SaveUser(_ context.Context, user model.User) (model.User, error) {
	memoryStore.mu.Lock()
	defer memoryStore.mu.Unlock()

	user.Username = strings.TrimSpace(user.Username)
	user.AccountEmail = model.NormalizeEmail(user.AccountEmail)
	if user.ID == 0 {
		user.ID = memoryStore.nextUserID
		memoryStore.nextUserID++
	}
	memoryStore.usersByID[user.ID] = user
	memoryStore.userIDsByName[user.Username] = user.ID
	memoryStore.usernames[user.Username] = struct{}{}
	if user.AccountEmail != "" {
		memoryStore.userIDsByEmail[user.AccountEmail] = user.ID
	}
	return user, nil
}

func (memoryStore *MemoryStore) FindUserByUsername(_ context.Context, username string) (model.User, error) {
	memoryStore.mu.RLock()
	defer memoryStore.mu.RUnlock()

	userID, found := memoryStore.userIDsByName[strings.TrimSpace(username)]
	if !found {
		return model.User{}, ErrNotFound
	}
	return memoryStore.usersByID[userID], nil
}

func (memoryStore *MemoryStore) FindUserByEmail(_ context.Context, email string) (model.User, error) {
	memoryStore.mu.RLock()
	defer memoryStore.mu.RUnlock()

	userID, found := memoryStore.userIDsByEmail[model.NormalizeEmail(email)]
	if !found {
		return model.User{}, ErrNotFound
	}
	return memoryStore.usersByID[userID], nil
}

func (memoryStore *MemoryStore) FindUserByID(_ context.Context, id int64) (model.User, error) {
	memoryStore.mu.RLock()
	defer memoryStore.mu.RUnlock()

	user, found := memoryStore.usersByID[id]
	if !found {
		return model.User{}, ErrNotFound
	}
	return user, nil
}

func (memoryStore *MemoryStore) SaveSession(_ context.Context, session model.Session) (model.Session, error) {
	memoryStore.mu.Lock()
	defer memoryStore.mu.Unlock()

	switch session.Kind {
	case model.SessionKindAdmin:
		memoryStore.adminSessions[session.TokenHash] = session
	case model.SessionKindUser:
		memoryStore.userSessions[session.TokenHash] = session
	default:
		return model.Session{}, ErrNotFound
	}
	return session, nil
}

func (memoryStore *MemoryStore) FindSessionByTokenHash(_ context.Context, kind model.SessionKind, tokenHash string) (model.Session, error) {
	memoryStore.mu.RLock()
	defer memoryStore.mu.RUnlock()

	switch kind {
	case model.SessionKindAdmin:
		session, found := memoryStore.adminSessions[tokenHash]
		if !found {
			return model.Session{}, ErrNotFound
		}
		return session, nil
	case model.SessionKindUser:
		session, found := memoryStore.userSessions[tokenHash]
		if !found {
			return model.Session{}, ErrNotFound
		}
		return session, nil
	default:
		return model.Session{}, ErrNotFound
	}
}

func (memoryStore *MemoryStore) DeleteSessionByTokenHash(_ context.Context, kind model.SessionKind, tokenHash string) error {
	memoryStore.mu.Lock()
	defer memoryStore.mu.Unlock()

	switch kind {
	case model.SessionKindAdmin:
		if _, found := memoryStore.adminSessions[tokenHash]; !found {
			return ErrNotFound
		}
		delete(memoryStore.adminSessions, tokenHash)
	case model.SessionKindUser:
		if _, found := memoryStore.userSessions[tokenHash]; !found {
			return ErrNotFound
		}
		delete(memoryStore.userSessions, tokenHash)
	default:
		return ErrNotFound
	}

	return nil
}

func (memoryStore *MemoryStore) SaveOrganization(_ context.Context, organization model.Organization) (model.Organization, error) {
	memoryStore.mu.Lock()
	defer memoryStore.mu.Unlock()

	organization.Slug = strings.TrimSpace(organization.Slug)
	now := organization.UpdatedAt
	if now.IsZero() {
		now = time.Now().UTC()
	}
	if organization.ID == 0 {
		organization.ID = memoryStore.nextOrgID
		memoryStore.nextOrgID++
		if organization.CreatedAt.IsZero() {
			organization.CreatedAt = now
		}
	}
	organization.UpdatedAt = now
	memoryStore.orgsByID[organization.ID] = organization
	memoryStore.orgIDsBySlug[organization.Slug] = organization.ID
	memoryStore.orgSlugs[organization.Slug] = struct{}{}
	return organization, nil
}

func (memoryStore *MemoryStore) FindOrganizationBySlug(_ context.Context, slug string) (model.Organization, error) {
	memoryStore.mu.RLock()
	defer memoryStore.mu.RUnlock()

	orgID, found := memoryStore.orgIDsBySlug[strings.TrimSpace(slug)]
	if !found {
		return model.Organization{}, ErrNotFound
	}
	return memoryStore.orgsByID[orgID], nil
}

func (memoryStore *MemoryStore) FindOrganizationByID(_ context.Context, id int64) (model.Organization, error) {
	memoryStore.mu.RLock()
	defer memoryStore.mu.RUnlock()

	organization, found := memoryStore.orgsByID[id]
	if !found {
		return model.Organization{}, ErrNotFound
	}
	return organization, nil
}

func (memoryStore *MemoryStore) SaveOrganizationPreferences(_ context.Context, preferences model.OrganizationPreferences) (model.OrganizationPreferences, error) {
	memoryStore.mu.Lock()
	defer memoryStore.mu.Unlock()

	now := preferences.UpdatedAt
	if now.IsZero() {
		now = time.Now().UTC()
	}
	if preferences.CreatedAt.IsZero() {
		preferences.CreatedAt = now
	}
	preferences.UpdatedAt = now
	memoryStore.orgPrefsByOrgID[preferences.OrgID] = preferences
	return preferences, nil
}

func (memoryStore *MemoryStore) FindOrganizationPreferencesByOrgID(_ context.Context, orgID int64) (model.OrganizationPreferences, error) {
	memoryStore.mu.RLock()
	defer memoryStore.mu.RUnlock()

	preferences, found := memoryStore.orgPrefsByOrgID[orgID]
	if !found {
		return model.OrganizationPreferences{}, ErrNotFound
	}
	return preferences, nil
}

func (memoryStore *MemoryStore) SaveOrganizationMember(_ context.Context, member model.OrganizationMember) (model.OrganizationMember, error) {
	memoryStore.mu.Lock()
	defer memoryStore.mu.Unlock()

	if member.ID == 0 {
		member.ID = memoryStore.nextOrgMemberID
		memoryStore.nextOrgMemberID++
	}
	if member.CreatedAt.IsZero() {
		member.CreatedAt = time.Now().UTC()
	}
	memoryStore.orgMembersByID[member.ID] = member
	return member, nil
}

func (memoryStore *MemoryStore) FindOrganizationMember(_ context.Context, id int64) (model.OrganizationMember, error) {
	memoryStore.mu.RLock()
	defer memoryStore.mu.RUnlock()

	member, found := memoryStore.orgMembersByID[id]
	if !found {
		return model.OrganizationMember{}, ErrNotFound
	}
	return member, nil
}

func (memoryStore *MemoryStore) FindOrganizationMemberByUserID(_ context.Context, orgID int64, userID int64) (model.OrganizationMember, error) {
	memoryStore.mu.RLock()
	defer memoryStore.mu.RUnlock()

	for _, member := range memoryStore.orgMembersByID {
		if member.OrgID == orgID && member.UserID == userID {
			return member, nil
		}
	}
	return model.OrganizationMember{}, ErrNotFound
}

func (memoryStore *MemoryStore) ListOrganizationMembers(_ context.Context, orgID int64) ([]model.OrganizationMember, error) {
	memoryStore.mu.RLock()
	defer memoryStore.mu.RUnlock()

	members := make([]model.OrganizationMember, 0)
	for _, member := range memoryStore.orgMembersByID {
		if member.OrgID == orgID {
			members = append(members, member)
		}
	}
	return members, nil
}

func (memoryStore *MemoryStore) SaveToken(_ context.Context, token model.Token) (model.Token, error) {
	memoryStore.mu.Lock()
	defer memoryStore.mu.Unlock()

	if token.ID == 0 {
		token.ID = memoryStore.nextTokenID
		memoryStore.nextTokenID++
	}

	switch token.OwnerType {
	case model.TokenOwnerAdmin:
		memoryStore.adminTokens[token.TokenHash] = token
	case model.TokenOwnerUser:
		memoryStore.userTokens[token.TokenHash] = token
	case model.TokenOwnerOrg:
		memoryStore.orgTokens[token.TokenHash] = token
	default:
		return model.Token{}, ErrNotFound
	}

	return token, nil
}

func (memoryStore *MemoryStore) FindTokenByHash(_ context.Context, ownerType model.TokenOwnerType, tokenHash string) (model.Token, error) {
	memoryStore.mu.RLock()
	defer memoryStore.mu.RUnlock()

	switch ownerType {
	case model.TokenOwnerAdmin:
		token, found := memoryStore.adminTokens[tokenHash]
		if !found {
			return model.Token{}, ErrNotFound
		}
		return token, nil
	case model.TokenOwnerUser:
		token, found := memoryStore.userTokens[tokenHash]
		if !found {
			return model.Token{}, ErrNotFound
		}
		return token, nil
	case model.TokenOwnerOrg:
		token, found := memoryStore.orgTokens[tokenHash]
		if !found {
			return model.Token{}, ErrNotFound
		}
		return token, nil
	default:
		return model.Token{}, ErrNotFound
	}
}

func (memoryStore *MemoryStore) DeleteTokenByHash(_ context.Context, ownerType model.TokenOwnerType, tokenHash string) error {
	memoryStore.mu.Lock()
	defer memoryStore.mu.Unlock()

	switch ownerType {
	case model.TokenOwnerAdmin:
		if _, found := memoryStore.adminTokens[tokenHash]; !found {
			return ErrNotFound
		}
		delete(memoryStore.adminTokens, tokenHash)
	case model.TokenOwnerUser:
		if _, found := memoryStore.userTokens[tokenHash]; !found {
			return ErrNotFound
		}
		delete(memoryStore.userTokens, tokenHash)
	case model.TokenOwnerOrg:
		if _, found := memoryStore.orgTokens[tokenHash]; !found {
			return ErrNotFound
		}
		delete(memoryStore.orgTokens, tokenHash)
	default:
		return ErrNotFound
	}

	return nil
}
