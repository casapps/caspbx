package server

import (
	"fmt"
	"path"
	"regexp"
	"strings"
)

var (
	apiVersionPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*$`)
	adminPathPattern  = regexp.MustCompile(`^[a-z0-9](?:[a-z0-9-]{0,30}[a-z0-9])$`)
	adminPathReserved = map[string]struct{}{
		"api": {}, "static": {}, "assets": {}, "health": {}, "healthz": {},
		"version": {}, "metrics": {}, ".well-known": {},
	}
)

type RouteCatalog struct {
	APIVersion            string
	APIBasePath           string
	AdminSegment          string
	AdminBasePath         string
	AdminAPIBasePath      string
	AsteriskAdminBasePath string
	AsteriskAdminAPIPath  string
	ServerBasePath        string
	ServerAPIBasePath     string
	AuthBasePath          string
	AuthAPIBasePath       string
	UsersBasePath         string
	UsersAPIBasePath      string
	OrgsBasePath          string
	OrgsAPIBasePath       string
}

func NewRouteCatalog(apiVersion string, adminPath string) (RouteCatalog, error) {
	normalizedAPIVersion := strings.TrimSpace(strings.ToLower(apiVersion))
	if !apiVersionPattern.MatchString(normalizedAPIVersion) {
		return RouteCatalog{}, fmt.Errorf("%w: %q", ErrInvalidAPIVersion, apiVersion)
	}

	if validationError := ValidateAdminPath(adminPath); validationError != nil {
		return RouteCatalog{}, validationError
	}

	adminBasePath := joinRouteParts(adminPath)
	apiBasePath := joinRouteParts("api", normalizedAPIVersion)

	return RouteCatalog{
		APIVersion:            normalizedAPIVersion,
		APIBasePath:           apiBasePath,
		AdminSegment:          adminPath,
		AdminBasePath:         adminBasePath,
		AdminAPIBasePath:      joinRouteParts("api", normalizedAPIVersion, adminPath),
		AsteriskAdminBasePath: joinRouteParts(adminPath, "server", "asterisk"),
		AsteriskAdminAPIPath:  joinRouteParts("api", normalizedAPIVersion, adminPath, "server", "asterisk"),
		ServerBasePath:        "/server",
		ServerAPIBasePath:     joinRouteParts("api", normalizedAPIVersion, "server"),
		AuthBasePath:          "/auth",
		AuthAPIBasePath:       joinRouteParts("api", normalizedAPIVersion, "auth"),
		UsersBasePath:         "/users",
		UsersAPIBasePath:      joinRouteParts("api", normalizedAPIVersion, "users"),
		OrgsBasePath:          "/orgs",
		OrgsAPIBasePath:       joinRouteParts("api", normalizedAPIVersion, "orgs"),
	}, nil
}

func ValidateAdminPath(adminPath string) error {
	normalizedPath := strings.TrimSpace(strings.ToLower(adminPath))
	if !adminPathPattern.MatchString(normalizedPath) {
		return fmt.Errorf("%w: %q", ErrInvalidAdminPath, adminPath)
	}
	if _, reserved := adminPathReserved[normalizedPath]; reserved {
		return fmt.Errorf("%w: %q is reserved", ErrInvalidAdminPath, adminPath)
	}
	return nil
}

func ValidateAdminRootRoute(adminRelativePath string) error {
	normalizedPath := strings.Trim(strings.TrimSpace(adminRelativePath), "/")
	if normalizedPath == "" {
		return nil
	}

	firstSegment := strings.Split(normalizedPath, "/")[0]
	if _, found := validAdminRootPaths[firstSegment]; !found {
		return fmt.Errorf("%w: /%s/*", ErrInvalidAdminRoute, firstSegment)
	}
	return nil
}

func (catalog RouteCatalog) AuthPath(parts ...string) string {
	return joinRouteParts(append([]string{"auth"}, parts...)...)
}

func (catalog RouteCatalog) AuthAPIPath(parts ...string) string {
	return joinRouteParts(append([]string{"api", catalog.APIVersion, "auth"}, parts...)...)
}

func (catalog RouteCatalog) UserPath(parts ...string) string {
	return joinRouteParts(append([]string{"users"}, parts...)...)
}

func (catalog RouteCatalog) UserAPIPath(parts ...string) string {
	return joinRouteParts(append([]string{"api", catalog.APIVersion, "users"}, parts...)...)
}

func (catalog RouteCatalog) OrgPath(slug string, parts ...string) string {
	allParts := []string{"orgs"}
	if strings.TrimSpace(slug) != "" {
		allParts = append(allParts, slug)
	}
	allParts = append(allParts, parts...)
	return joinRouteParts(allParts...)
}

func (catalog RouteCatalog) OrgAPIPath(slug string, parts ...string) string {
	allParts := []string{"api", catalog.APIVersion, "orgs"}
	if strings.TrimSpace(slug) != "" {
		allParts = append(allParts, slug)
	}
	allParts = append(allParts, parts...)
	return joinRouteParts(allParts...)
}

func (catalog RouteCatalog) AdminRoute(parts ...string) string {
	return joinRouteParts(append([]string{catalog.AdminSegment}, parts...)...)
}

func (catalog RouteCatalog) AdminAPIRoute(parts ...string) string {
	return joinRouteParts(append([]string{"api", catalog.APIVersion, catalog.AdminSegment}, parts...)...)
}

func (catalog RouteCatalog) AsteriskAdminRoute(parts ...string) string {
	return joinRouteParts(append([]string{catalog.AdminSegment, "server", "asterisk"}, parts...)...)
}

func (catalog RouteCatalog) AsteriskAdminAPIRoute(parts ...string) string {
	return joinRouteParts(append([]string{"api", catalog.APIVersion, catalog.AdminSegment, "server", "asterisk"}, parts...)...)
}

func joinRouteParts(parts ...string) string {
	cleanParts := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmedPart := strings.TrimSpace(part)
		if trimmedPart == "" {
			continue
		}
		cleanParts = append(cleanParts, strings.Trim(trimmedPart, "/"))
	}

	if len(cleanParts) == 0 {
		return "/"
	}

	joinedPath := path.Join(cleanParts...)
	if joinedPath == "." {
		return "/"
	}

	return "/" + joinedPath
}
