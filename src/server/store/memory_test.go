package store

import (
	"context"
	"testing"
	"time"

	"github.com/casapps/caspbx/src/server/model"
)

func TestMemoryStoreAdminUserAndSessionPersistence(t *testing.T) {
	memoryStore := NewMemoryStore()

	if exists, lookupError := memoryStore.UserExistsByName(context.Background(), "alice"); lookupError != nil || exists {
		t.Fatalf("expected missing username lookup, got %v / %t", lookupError, exists)
	}
	memoryStore.orgSlugs["acme"] = struct{}{}
	if exists, lookupError := memoryStore.OrgExistsBySlug(context.Background(), "acme"); lookupError != nil || !exists {
		t.Fatalf("expected org slug lookup, got %v / %t", lookupError, exists)
	}

	admin, saveAdminError := memoryStore.SaveAdmin(context.Background(), model.Admin{Username: "root-admin", Enabled: true})
	if saveAdminError != nil || admin.ID == 0 {
		t.Fatalf("expected saved admin, got %v / %+v", saveAdminError, admin)
	}
	if _, lookupError := memoryStore.FindAdminByUsername(context.Background(), "missing"); lookupError != ErrNotFound {
		t.Fatalf("expected missing admin lookup error, got %v", lookupError)
	}
	if foundAdmin, lookupError := memoryStore.FindAdminByUsername(context.Background(), "root-admin"); lookupError != nil || foundAdmin.ID != admin.ID {
		t.Fatalf("expected found admin by username, got %v / %+v", lookupError, foundAdmin)
	}
	if foundAdmin, lookupError := memoryStore.FindAdminByID(context.Background(), admin.ID); lookupError != nil || foundAdmin.Username != "root-admin" {
		t.Fatalf("expected found admin, got %v / %+v", lookupError, foundAdmin)
	}
	if _, lookupError := memoryStore.FindAdminByID(context.Background(), 999); lookupError != ErrNotFound {
		t.Fatalf("expected missing admin id lookup error, got %v", lookupError)
	}

	user, saveUserError := memoryStore.SaveUser(context.Background(), model.User{Username: "alice", AccountEmail: "Alice@Example.com", Enabled: true})
	if saveUserError != nil || user.ID == 0 {
		t.Fatalf("expected saved user, got %v / %+v", saveUserError, user)
	}
	if exists, lookupError := memoryStore.UserExistsByName(context.Background(), "alice"); lookupError != nil || !exists {
		t.Fatalf("expected username to exist, got %v / %t", lookupError, exists)
	}
	if foundUser, lookupError := memoryStore.FindUserByUsername(context.Background(), "alice"); lookupError != nil || foundUser.AccountEmail != "alice@example.com" {
		t.Fatalf("expected found user by username, got %v / %+v", lookupError, foundUser)
	}
	if foundUser, lookupError := memoryStore.FindUserByEmail(context.Background(), "alice@example.com"); lookupError != nil || foundUser.Username != "alice" {
		t.Fatalf("expected found user by email, got %v / %+v", lookupError, foundUser)
	}
	if foundUser, lookupError := memoryStore.FindUserByID(context.Background(), user.ID); lookupError != nil || foundUser.Username != "alice" {
		t.Fatalf("expected found user by id, got %v / %+v", lookupError, foundUser)
	}
	if _, lookupError := memoryStore.FindUserByUsername(context.Background(), "missing"); lookupError != ErrNotFound {
		t.Fatalf("expected missing user lookup error, got %v", lookupError)
	}
	if _, lookupError := memoryStore.FindUserByEmail(context.Background(), "missing@example.com"); lookupError != ErrNotFound {
		t.Fatalf("expected missing email lookup error, got %v", lookupError)
	}
	if _, lookupError := memoryStore.FindUserByID(context.Background(), 999); lookupError != ErrNotFound {
		t.Fatalf("expected missing user id lookup error, got %v", lookupError)
	}

	adminSession := model.Session{Kind: model.SessionKindAdmin, TokenHash: "adm-token", SubjectID: admin.ID, ExpiresAt: time.Unix(10, 0)}
	if savedSession, saveSessionError := memoryStore.SaveSession(context.Background(), adminSession); saveSessionError != nil || savedSession.TokenHash != "adm-token" {
		t.Fatalf("expected saved admin session, got %v / %+v", saveSessionError, savedSession)
	}
	if foundSession, lookupError := memoryStore.FindSessionByTokenHash(context.Background(), model.SessionKindAdmin, "adm-token"); lookupError != nil || foundSession.SubjectID != admin.ID {
		t.Fatalf("expected found admin session, got %v / %+v", lookupError, foundSession)
	}
	if deleteError := memoryStore.DeleteSessionByTokenHash(context.Background(), model.SessionKindAdmin, "adm-token"); deleteError != nil {
		t.Fatalf("expected deleted admin session, got %v", deleteError)
	}
	if deleteError := memoryStore.DeleteSessionByTokenHash(context.Background(), model.SessionKindAdmin, "adm-token"); deleteError != ErrNotFound {
		t.Fatalf("expected missing admin session delete error, got %v", deleteError)
	}
	if _, lookupError := memoryStore.FindSessionByTokenHash(context.Background(), model.SessionKindAdmin, "adm-token"); lookupError != ErrNotFound {
		t.Fatalf("expected missing admin session lookup error, got %v", lookupError)
	}

	userSession := model.Session{Kind: model.SessionKindUser, TokenHash: "usr-token", SubjectID: user.ID, ExpiresAt: time.Unix(10, 0)}
	if _, saveSessionError := memoryStore.SaveSession(context.Background(), userSession); saveSessionError != nil {
		t.Fatalf("expected saved user session, got %v", saveSessionError)
	}
	if foundSession, lookupError := memoryStore.FindSessionByTokenHash(context.Background(), model.SessionKindUser, "usr-token"); lookupError != nil || foundSession.SubjectID != user.ID {
		t.Fatalf("expected found user session, got %v / %+v", lookupError, foundSession)
	}
	if deleteError := memoryStore.DeleteSessionByTokenHash(context.Background(), model.SessionKindUser, "usr-token"); deleteError != nil {
		t.Fatalf("expected deleted user session, got %v", deleteError)
	}
	if deleteError := memoryStore.DeleteSessionByTokenHash(context.Background(), model.SessionKindUser, "usr-token"); deleteError != ErrNotFound {
		t.Fatalf("expected missing user session delete error, got %v", deleteError)
	}
	if _, lookupError := memoryStore.FindSessionByTokenHash(context.Background(), model.SessionKindUser, "usr-token"); lookupError != ErrNotFound {
		t.Fatalf("expected missing user session lookup error, got %v", lookupError)
	}
	if _, saveSessionError := memoryStore.SaveSession(context.Background(), model.Session{Kind: "unknown"}); saveSessionError != ErrNotFound {
		t.Fatalf("expected invalid kind save error, got %v", saveSessionError)
	}
	if _, lookupError := memoryStore.FindSessionByTokenHash(context.Background(), "unknown", "missing"); lookupError != ErrNotFound {
		t.Fatalf("expected invalid kind lookup error, got %v", lookupError)
	}
	if deleteError := memoryStore.DeleteSessionByTokenHash(context.Background(), "unknown", "missing"); deleteError != ErrNotFound {
		t.Fatalf("expected invalid kind delete error, got %v", deleteError)
	}

	adminToken := model.Token{OwnerType: model.TokenOwnerAdmin, OwnerID: admin.ID, Name: "default", TokenHash: "adm-hash", TokenPrefix: "adm_ab12", Scope: model.TokenScopeGlobal, ExpiresAt: time.Unix(20, 0)}
	if savedToken, saveTokenError := memoryStore.SaveToken(context.Background(), adminToken); saveTokenError != nil || savedToken.ID == 0 {
		t.Fatalf("expected saved admin token, got %v / %+v", saveTokenError, savedToken)
	}
	if foundToken, lookupError := memoryStore.FindTokenByHash(context.Background(), model.TokenOwnerAdmin, "adm-hash"); lookupError != nil || foundToken.OwnerID != admin.ID {
		t.Fatalf("expected found admin token, got %v / %+v", lookupError, foundToken)
	}
	if deleteError := memoryStore.DeleteTokenByHash(context.Background(), model.TokenOwnerAdmin, "adm-hash"); deleteError != nil {
		t.Fatalf("expected deleted admin token, got %v", deleteError)
	}
	if deleteError := memoryStore.DeleteTokenByHash(context.Background(), model.TokenOwnerAdmin, "adm-hash"); deleteError != ErrNotFound {
		t.Fatalf("expected missing admin token delete error, got %v", deleteError)
	}
	if _, lookupError := memoryStore.FindTokenByHash(context.Background(), model.TokenOwnerAdmin, "adm-hash"); lookupError != ErrNotFound {
		t.Fatalf("expected missing admin token lookup error, got %v", lookupError)
	}

	userToken := model.Token{OwnerType: model.TokenOwnerUser, OwnerID: user.ID, Name: "default", TokenHash: "usr-hash", TokenPrefix: "usr_ab12", Scope: model.TokenScopeGlobal}
	if _, saveTokenError := memoryStore.SaveToken(context.Background(), userToken); saveTokenError != nil {
		t.Fatalf("expected saved user token, got %v", saveTokenError)
	}
	if foundToken, lookupError := memoryStore.FindTokenByHash(context.Background(), model.TokenOwnerUser, "usr-hash"); lookupError != nil || foundToken.OwnerID != user.ID {
		t.Fatalf("expected found user token, got %v / %+v", lookupError, foundToken)
	}
	if deleteError := memoryStore.DeleteTokenByHash(context.Background(), model.TokenOwnerUser, "usr-hash"); deleteError != nil {
		t.Fatalf("expected deleted user token, got %v", deleteError)
	}
	if deleteError := memoryStore.DeleteTokenByHash(context.Background(), model.TokenOwnerUser, "usr-hash"); deleteError != ErrNotFound {
		t.Fatalf("expected missing user token delete error, got %v", deleteError)
	}
	if _, lookupError := memoryStore.FindTokenByHash(context.Background(), model.TokenOwnerUser, "usr-hash"); lookupError != ErrNotFound {
		t.Fatalf("expected missing user token lookup error, got %v", lookupError)
	}

	orgToken := model.Token{OwnerType: model.TokenOwnerOrg, OwnerID: 7, Name: "org", TokenHash: "org-hash", TokenPrefix: "org_ab12", Scope: model.TokenScopeRead}
	if _, saveTokenError := memoryStore.SaveToken(context.Background(), orgToken); saveTokenError != nil {
		t.Fatalf("expected saved org token, got %v", saveTokenError)
	}
	if foundToken, lookupError := memoryStore.FindTokenByHash(context.Background(), model.TokenOwnerOrg, "org-hash"); lookupError != nil || foundToken.OwnerID != 7 {
		t.Fatalf("expected found org token, got %v / %+v", lookupError, foundToken)
	}
	if deleteError := memoryStore.DeleteTokenByHash(context.Background(), model.TokenOwnerOrg, "org-hash"); deleteError != nil {
		t.Fatalf("expected deleted org token, got %v", deleteError)
	}
	if deleteError := memoryStore.DeleteTokenByHash(context.Background(), model.TokenOwnerOrg, "org-hash"); deleteError != ErrNotFound {
		t.Fatalf("expected missing org token delete error, got %v", deleteError)
	}
	if _, lookupError := memoryStore.FindTokenByHash(context.Background(), model.TokenOwnerOrg, "org-hash"); lookupError != ErrNotFound {
		t.Fatalf("expected missing org token lookup error, got %v", lookupError)
	}
	if _, saveTokenError := memoryStore.SaveToken(context.Background(), model.Token{OwnerType: "unknown"}); saveTokenError != ErrNotFound {
		t.Fatalf("expected invalid token owner save error, got %v", saveTokenError)
	}
	if _, lookupError := memoryStore.FindTokenByHash(context.Background(), "unknown", "missing"); lookupError != ErrNotFound {
		t.Fatalf("expected invalid token owner lookup error, got %v", lookupError)
	}
	if deleteError := memoryStore.DeleteTokenByHash(context.Background(), "unknown", "missing"); deleteError != ErrNotFound {
		t.Fatalf("expected invalid token owner delete error, got %v", deleteError)
	}

	organization, saveOrgError := memoryStore.SaveOrganization(context.Background(), model.Organization{
		Slug:       "acme",
		Name:       "Acme",
		Visibility: model.OrganizationVisibilityPublic,
	})
	if saveOrgError != nil || organization.ID == 0 {
		t.Fatalf("expected saved organization, got %v / %+v", saveOrgError, organization)
	}
	if foundOrg, lookupError := memoryStore.FindOrganizationBySlug(context.Background(), "acme"); lookupError != nil || foundOrg.ID != organization.ID {
		t.Fatalf("expected found organization by slug, got %v / %+v", lookupError, foundOrg)
	}
	if foundOrg, lookupError := memoryStore.FindOrganizationByID(context.Background(), organization.ID); lookupError != nil || foundOrg.Slug != "acme" {
		t.Fatalf("expected found organization by id, got %v / %+v", lookupError, foundOrg)
	}
	if _, lookupError := memoryStore.FindOrganizationBySlug(context.Background(), "missing"); lookupError != ErrNotFound {
		t.Fatalf("expected missing organization slug lookup error, got %v", lookupError)
	}
	if _, lookupError := memoryStore.FindOrganizationByID(context.Background(), 999); lookupError != ErrNotFound {
		t.Fatalf("expected missing organization id lookup error, got %v", lookupError)
	}

	preferences := model.DefaultOrganizationPreferences()
	preferences.OrgID = organization.ID
	if savedPreferences, savePreferencesError := memoryStore.SaveOrganizationPreferences(context.Background(), preferences); savePreferencesError != nil || savedPreferences.OrgID != organization.ID {
		t.Fatalf("expected saved organization preferences, got %v / %+v", savePreferencesError, savedPreferences)
	}
	if foundPreferences, lookupError := memoryStore.FindOrganizationPreferencesByOrgID(context.Background(), organization.ID); lookupError != nil || !foundPreferences.AllowInvites {
		t.Fatalf("expected found organization preferences, got %v / %+v", lookupError, foundPreferences)
	}
	if _, lookupError := memoryStore.FindOrganizationPreferencesByOrgID(context.Background(), 999); lookupError != ErrNotFound {
		t.Fatalf("expected missing organization preferences lookup error, got %v", lookupError)
	}

	member, saveMemberError := memoryStore.SaveOrganizationMember(context.Background(), model.OrganizationMember{
		OrgID:  organization.ID,
		UserID: user.ID,
		Role:   model.OrganizationRoleOwner,
	})
	if saveMemberError != nil || member.ID == 0 {
		t.Fatalf("expected saved organization member, got %v / %+v", saveMemberError, member)
	}
	if foundMember, lookupError := memoryStore.FindOrganizationMember(context.Background(), member.ID); lookupError != nil || foundMember.UserID != user.ID {
		t.Fatalf("expected found organization member, got %v / %+v", lookupError, foundMember)
	}
	if foundMember, lookupError := memoryStore.FindOrganizationMemberByUserID(context.Background(), organization.ID, user.ID); lookupError != nil || foundMember.ID != member.ID {
		t.Fatalf("expected found organization member by user id, got %v / %+v", lookupError, foundMember)
	}
	if listedMembers, listError := memoryStore.ListOrganizationMembers(context.Background(), organization.ID); listError != nil || len(listedMembers) != 1 {
		t.Fatalf("expected listed organization members, got %v / %+v", listError, listedMembers)
	}
	if _, lookupError := memoryStore.FindOrganizationMember(context.Background(), 999); lookupError != ErrNotFound {
		t.Fatalf("expected missing organization member lookup error, got %v", lookupError)
	}
	if _, lookupError := memoryStore.FindOrganizationMemberByUserID(context.Background(), organization.ID, 999); lookupError != ErrNotFound {
		t.Fatalf("expected missing organization member-by-user lookup error, got %v", lookupError)
	}
}
