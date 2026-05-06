package model

import "strings"

var ReservedNames = []string{
	"api", "admin", "static", "assets", "healthz", "metrics",
	"login", "logout", "register", "signup", "signin", "auth",
	"oauth", "callback", "webhook", "webhooks",
	"users", "orgs", "organizations", "teams", "groups",
	"settings", "profile", "account", "dashboard",
	"search", "explore", "discover", "trending",
	"help", "support", "docs", "documentation",
	"about", "contact", "terms", "privacy", "legal",
	"graphql", "rest", "rpc", "ws", "websocket",
	"cdn", "media", "uploads", "files", "images",
	".well-known", "robots.txt", "sitemap.xml", "favicon.ico",
	"caspbx", "casapps",
}

func NormalizeSharedName(input string) string {
	return strings.TrimSpace(strings.ToLower(input))
}

func IsReservedName(input string) bool {
	normalizedName := NormalizeSharedName(input)
	for _, reservedName := range ReservedNames {
		if normalizedName == reservedName {
			return true
		}
	}
	return false
}
