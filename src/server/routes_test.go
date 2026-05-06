package server

import "testing"

func TestNewRouteCatalog(t *testing.T) {
	routeCatalog, routeError := NewRouteCatalog(DefaultAPIVersion, "admin")
	if routeError != nil {
		t.Fatalf("expected valid route catalog, got %v", routeError)
	}

	if routeCatalog.APIBasePath != "/api/v1" {
		t.Fatalf("unexpected api base path %q", routeCatalog.APIBasePath)
	}
	if routeCatalog.AdminBasePath != "/admin" {
		t.Fatalf("unexpected admin base path %q", routeCatalog.AdminBasePath)
	}
	if routeCatalog.AsteriskAdminBasePath != "/admin/server/asterisk" {
		t.Fatalf("unexpected asterisk admin path %q", routeCatalog.AsteriskAdminBasePath)
	}
	if routeCatalog.AuthPath("login") != "/auth/login" {
		t.Fatalf("unexpected auth path %q", routeCatalog.AuthPath("login"))
	}
	if routeCatalog.AuthAPIPath("login") != "/api/v1/auth/login" {
		t.Fatalf("unexpected auth api path %q", routeCatalog.AuthAPIPath("login"))
	}
	if routeCatalog.UserPath("settings") != "/users/settings" {
		t.Fatalf("unexpected user path %q", routeCatalog.UserPath("settings"))
	}
	if routeCatalog.UserAPIPath("domains") != "/api/v1/users/domains" {
		t.Fatalf("unexpected user api path %q", routeCatalog.UserAPIPath("domains"))
	}
	if routeCatalog.OrgPath("acme", "members") != "/orgs/acme/members" {
		t.Fatalf("unexpected org path %q", routeCatalog.OrgPath("acme", "members"))
	}
	if routeCatalog.OrgAPIPath("acme", "members") != "/api/v1/orgs/acme/members" {
		t.Fatalf("unexpected org api path %q", routeCatalog.OrgAPIPath("acme", "members"))
	}
	if routeCatalog.AdminRoute("server", "settings") != "/admin/server/settings" {
		t.Fatalf("unexpected admin route %q", routeCatalog.AdminRoute("server", "settings"))
	}
	if routeCatalog.AdminAPIRoute("server", "users") != "/api/v1/admin/server/users" {
		t.Fatalf("unexpected admin api route %q", routeCatalog.AdminAPIRoute("server", "users"))
	}
	if routeCatalog.AsteriskAdminRoute("fax") != "/admin/server/asterisk/fax" {
		t.Fatalf("unexpected asterisk admin route %q", routeCatalog.AsteriskAdminRoute("fax"))
	}
	if routeCatalog.AsteriskAdminAPIRoute("fax") != "/api/v1/admin/server/asterisk/fax" {
		t.Fatalf("unexpected asterisk admin api route %q", routeCatalog.AsteriskAdminAPIRoute("fax"))
	}
}

func TestValidateAdminPath(t *testing.T) {
	if validationError := ValidateAdminPath("admin"); validationError != nil {
		t.Fatalf("expected valid admin path, got %v", validationError)
	}
	if validationError := ValidateAdminPath("a"); validationError == nil {
		t.Fatalf("expected short admin path to fail")
	}
	if validationError := ValidateAdminPath("api"); validationError == nil {
		t.Fatalf("expected reserved admin path to fail")
	}
	if _, routeError := NewRouteCatalog("bad/version", "admin"); routeError == nil {
		t.Fatalf("expected invalid api version to fail")
	}
}

func TestValidateAdminRootRoute(t *testing.T) {
	if validationError := ValidateAdminRootRoute(""); validationError != nil {
		t.Fatalf("expected empty admin route to validate, got %v", validationError)
	}
	if validationError := ValidateAdminRootRoute("server/settings"); validationError != nil {
		t.Fatalf("expected valid admin route, got %v", validationError)
	}
	if validationError := ValidateAdminRootRoute("profile"); validationError != nil {
		t.Fatalf("expected valid admin profile route, got %v", validationError)
	}
	if validationError := ValidateAdminRootRoute("users"); validationError == nil {
		t.Fatalf("expected invalid root-level admin route to fail")
	}
}

func TestNormalizeURLPath(t *testing.T) {
	canonicalPath, redirectRequired := NormalizeURLPath("/")
	if canonicalPath != "/" || redirectRequired {
		t.Fatalf("expected root path to remain unchanged, got %q / %t", canonicalPath, redirectRequired)
	}

	canonicalPath, redirectRequired = NormalizeURLPath("/users/")
	if canonicalPath != "/users" || !redirectRequired {
		t.Fatalf("expected users path redirect, got %q / %t", canonicalPath, redirectRequired)
	}

	canonicalPath, redirectRequired = NormalizeURLPath("/static/app.css")
	if canonicalPath != "/static/app.css" || redirectRequired {
		t.Fatalf("expected file path to remain unchanged, got %q / %t", canonicalPath, redirectRequired)
	}

	canonicalPath, redirectRequired = NormalizeURLPath("")
	if canonicalPath != "/" || redirectRequired {
		t.Fatalf("expected empty path to normalize to root, got %q / %t", canonicalPath, redirectRequired)
	}

	if !isFileLikePath("favicon.ico") {
		t.Fatalf("expected filename without slash to be file-like")
	}
	if isFileLikePath("users") {
		t.Fatalf("expected route segment without dot not to be file-like")
	}
	if !isFileLikePath("/assets/app.css") {
		t.Fatalf("expected dotted path suffix to be file-like")
	}
	if isFileLikePath("/users/profile") {
		t.Fatalf("expected undotted path suffix not to be file-like")
	}
	if joinedPath := joinRouteParts("", " /admin/ ", "server", "", "asterisk/"); joinedPath != "/admin/server/asterisk" {
		t.Fatalf("unexpected joined route path %q", joinedPath)
	}
	if joinedPath := joinRouteParts("."); joinedPath != "/" {
		t.Fatalf("expected dot route to normalize to root, got %q", joinedPath)
	}
	if joinedPath := joinRouteParts("/"); joinedPath != "/" {
		t.Fatalf("expected slash route to normalize to root, got %q", joinedPath)
	}
	if joinedPath := joinRouteParts(); joinedPath != "/" {
		t.Fatalf("expected no route parts to normalize to root, got %q", joinedPath)
	}
	if joinedPath := joinRouteParts("", " ", "/"); joinedPath != "/" {
		t.Fatalf("expected empty route parts to normalize to root, got %q", joinedPath)
	}
}

func TestDetectClientType(t *testing.T) {
	if clientType := DetectClientType("text/html", ""); clientType != ClientTypeHTML {
		t.Fatalf("expected html client type, got %s", clientType)
	}
	if clientType := DetectClientType("text/plain", ""); clientType != ClientTypeText {
		t.Fatalf("expected text client type, got %s", clientType)
	}
	if clientType := DetectClientType("application/json", ""); clientType != ClientTypeJSON {
		t.Fatalf("expected json client type, got %s", clientType)
	}
	if clientType := DetectClientType("", "Mozilla/5.0"); clientType != ClientTypeHTML {
		t.Fatalf("expected browser user-agent to map to html, got %s", clientType)
	}
	if clientType := DetectClientType("", "curl/8.0"); clientType != ClientTypeText {
		t.Fatalf("expected cli user-agent to map to text, got %s", clientType)
	}
	if clientType := DetectClientType("", ""); clientType != ClientTypeText {
		t.Fatalf("expected empty user-agent to map to text, got %s", clientType)
	}
	if clientType := DetectClientType("", "unknown"); clientType != ClientTypeHTML {
		t.Fatalf("expected unknown user-agent to fall back to html, got %s", clientType)
	}
}

func TestDetectAPIResponseFormat(t *testing.T) {
	if clientType := DetectAPIResponseFormat("/api/v1/users.txt", "", ""); clientType != ClientTypeText {
		t.Fatalf("expected txt api route to map to text, got %s", clientType)
	}
	if clientType := DetectAPIResponseFormat("/api/v1/users", "text/plain", ""); clientType != ClientTypeText {
		t.Fatalf("expected plain accept header to map to text, got %s", clientType)
	}
	if clientType := DetectAPIResponseFormat("/api/v1/users", "", "curl/8.0"); clientType != ClientTypeText {
		t.Fatalf("expected cli api client to map to text, got %s", clientType)
	}
	if clientType := DetectAPIResponseFormat("/api/v1/users", "", ""); clientType != ClientTypeText {
		t.Fatalf("expected empty api user-agent to map to text, got %s", clientType)
	}
	if clientType := DetectAPIResponseFormat("/api/v1/users", "application/json", "Mozilla/5.0"); clientType != ClientTypeJSON {
		t.Fatalf("expected explicit json accept header, got %s", clientType)
	}
	if clientType := DetectAPIResponseFormat("/api/v1/users", "", "Mozilla/5.0"); clientType != ClientTypeJSON {
		t.Fatalf("expected browser api client to default to json, got %s", clientType)
	}
}

func TestBootstrapSummary(t *testing.T) {
	bootstrap, bootstrapError := NewBootstrap(DefaultAPIVersion, "admin")
	if bootstrapError != nil {
		t.Fatalf("expected bootstrap to build, got %v", bootstrapError)
	}

	expectedSummary := "API base path: /api/v1\nAdmin path: /admin\nAsterisk admin path: /admin/server/asterisk"
	if bootstrap.Summary() != expectedSummary {
		t.Fatalf("unexpected bootstrap summary %q", bootstrap.Summary())
	}
	if _, bootstrapError := NewBootstrap(DefaultAPIVersion, "api"); bootstrapError == nil {
		t.Fatalf("expected bootstrap to fail on invalid admin path")
	}
}
