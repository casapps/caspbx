package server

const DefaultAPIVersion = "v1"

var validAdminRootPaths = map[string]struct{}{
	"":              {},
	"profile":       {},
	"preferences":   {},
	"notifications": {},
	"server":        {},
}
