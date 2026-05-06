package model

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
)

type RegistrationMode string
type IdentifierType string
type UserVisibility string

const (
	RegistrationModePublic   RegistrationMode = "public"
	RegistrationModePrivate  RegistrationMode = "private"
	RegistrationModeDisabled RegistrationMode = "disabled"

	IdentifierTypeUserID   IdentifierType = "user_id"
	IdentifierTypeUsername IdentifierType = "username"
	IdentifierTypeEmail    IdentifierType = "email"

	UserVisibilityPublic  UserVisibility = "public"
	UserVisibilityPrivate UserVisibility = "private"
)

var (
	usernamePattern    = regexp.MustCompile(`^[a-z][a-z0-9_-]{1,30}[a-z0-9]$`)
	userIDPattern      = regexp.MustCompile(`^\d+$`)
	emailLocalPattern  = regexp.MustCompile(`^[a-z0-9.+_-]+$`)
	emailDomainPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9.-]*[a-z0-9]\.[a-z]{2,}$`)

	criticalUsernameSubstrings = []string{"admin", "root", "system", "mod", "official", "verified"}
	usernameBlocklist          = buildStringSet([]string{
		"admin", "administrator", "root", "system", "sysadmin", "superuser",
		"master", "owner", "operator", "manager", "moderator", "mod",
		"staff", "support", "helpdesk", "help", "service", "daemon",
		"server", "host", "node", "cluster", "api", "www", "web", "mail",
		"email", "smtp", "ftp", "ssh", "dns", "proxy", "gateway", "router",
		"firewall", "localhost", "local", "internal", "external", "public",
		"private", "network", "database", "db", "cache", "redis", "mysql",
		"postgres", "mongodb", "elastic", "nginx", "apache", "docker",
		"app", "application", "bot", "robot", "crawler", "spider", "scraper",
		"webhook", "callback", "cron", "scheduler", "worker", "queue", "job",
		"task", "process", "service", "microservice", "lambda", "function",
		"auth", "authentication", "login", "logout", "signin", "signout",
		"signup", "register", "password", "passwd", "token", "oauth", "sso",
		"saml", "ldap", "kerberos", "security", "secure", "ssl", "tls",
		"certificate", "cert", "key", "secret", "credential", "session",
		"guest", "anonymous", "anon", "user", "users", "member", "members",
		"subscriber", "editor", "author", "contributor", "reviewer", "auditor",
		"analyst", "developer", "dev", "devops", "engineer", "architect",
		"designer", "tester", "qa", "billing", "finance", "legal", "hr",
		"sales", "marketing", "ceo", "cto", "cfo", "coo", "founder", "cofounder",
		"account", "accounts", "profile", "profiles", "settings", "config",
		"configuration", "dashboard", "panel", "console", "portal", "home",
		"index", "main", "default", "null", "nil", "undefined", "void",
		"true", "false", "test", "testing", "debug", "demo", "example",
		"sample", "temp", "temporary", "tmp", "backup", "archive", "log",
		"logs", "audit", "report", "reports", "analytics", "stats", "status",
		"api", "rest", "graphql", "grpc", "websocket", "ws", "wss", "http",
		"https", "endpoint", "endpoints", "route", "routes", "path", "url",
		"uri", "callback", "hook", "hooks", "event", "events", "stream",
		"blog", "news", "article", "articles", "post", "posts", "page", "pages",
		"feed", "rss", "atom", "sitemap", "robots", "favicon", "static",
		"assets", "images", "image", "img", "media", "upload", "uploads",
		"download", "downloads", "file", "files", "document", "documents",
		"contact", "message", "messages", "chat", "notification", "notifications",
		"alert", "alerts", "inbox", "outbox", "sent", "draft", "drafts",
		"spam", "abuse", "report", "flag", "block", "mute", "ban",
		"shop", "store", "cart", "checkout", "order", "orders", "invoice",
		"invoices", "payment", "payments", "subscription", "subscriptions",
		"plan", "plans", "pricing", "refund", "coupon", "discount",
		"follow", "follower", "followers", "following", "friend", "friends",
		"like", "likes", "share", "shares", "comment", "comments", "reply",
		"mention", "mentions", "tag", "tags", "group", "groups", "team", "teams",
		"community", "communities", "forum", "forums", "channel", "channels",
		"official", "verified", "trusted", "partner", "affiliate", "sponsor",
		"brand", "trademark", "copyright", "legal", "terms", "privacy",
		"policy", "policies", "tos", "eula", "gdpr", "dmca", "abuse",
		"fuck", "shit", "ass", "bitch", "bastard", "damn", "cunt", "dick",
		"penis", "vagina", "sex", "porn", "xxx", "nude", "naked", "nsfw",
		"kill", "murder", "death", "die", "suicide", "hate", "nazi", "hitler",
		"racist", "racism", "terrorist", "terrorism", "isis", "alqaeda",
		"0", "1", "123", "1234", "12345", "000", "111", "666", "911", "420", "69",
		"info", "noreply", "no-reply", "donotreply", "mailer", "postmaster",
		"webmaster", "hostmaster", "abuse", "spam", "junk", "trash",
		"caspbx", "casapps",
	})
)

type User struct {
	ID                int64
	Username          string
	DisplayName       string
	AccountEmail      string
	NotificationEmail string
	PasswordHash      string
	Visibility        UserVisibility
	OrgVisibility     bool
	AvatarURLs        map[string]string
	Enabled           bool
	Verified          bool
	CreatedAt         time.Time
	UpdatedAt         time.Time
	LastLoginAt       time.Time
}

func DefaultRegistrationMode() RegistrationMode {
	return RegistrationModePrivate
}

func ParseRegistrationMode(input string) (RegistrationMode, error) {
	switch strings.TrimSpace(strings.ToLower(input)) {
	case "public":
		return RegistrationModePublic, nil
	case "private":
		return RegistrationModePrivate, nil
	case "disabled":
		return RegistrationModeDisabled, nil
	default:
		return "", fmt.Errorf("invalid registration mode %q", input)
	}
}

func DetectIdentifierType(input string) IdentifierType {
	switch {
	case userIDPattern.MatchString(strings.TrimSpace(input)):
		return IdentifierTypeUserID
	case strings.Contains(input, "@"):
		return IdentifierTypeEmail
	default:
		return IdentifierTypeUsername
	}
}

func ValidateUsername(input string) error {
	username := strings.TrimSpace(input)
	if username == "" {
		return errors.New("username is required")
	}
	if username != strings.ToLower(username) {
		return errors.New("username can only contain lowercase letters, numbers, underscore, and hyphen")
	}
	if len(username) < 3 {
		return errors.New("username must be at least 3 characters")
	}
	if len(username) > 32 {
		return errors.New("username cannot exceed 32 characters")
	}
	if !usernamePattern.MatchString(username) {
		return errors.New("username can only contain lowercase letters, numbers, underscore, and hyphen")
	}
	if strings.Contains(username, "__") || strings.Contains(username, "--") || strings.Contains(username, "_-") || strings.Contains(username, "-_") {
		return errors.New("username cannot contain consecutive separators")
	}
	if IsReservedName(username) {
		return fmt.Errorf("username contains blocked word: %s", username)
	}
	if _, blocked := usernameBlocklist[username]; blocked {
		return fmt.Errorf("username contains blocked word: %s", username)
	}
	for _, criticalSubstring := range criticalUsernameSubstrings {
		if strings.Contains(username, criticalSubstring) {
			return fmt.Errorf("username contains blocked word: %s", criticalSubstring)
		}
	}
	return nil
}

func NormalizeEmail(input string) string {
	return strings.TrimSpace(strings.ToLower(input))
}

func ValidateEmail(input string) error {
	email := NormalizeEmail(input)
	if len(email) > 254 {
		return errors.New("email too long (max 254 characters)")
	}

	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return errors.New("invalid email format")
	}

	localPart := parts[0]
	domainPart := parts[1]
	if len(localPart) == 0 || len(localPart) > 64 {
		return errors.New("invalid local part length")
	}
	if strings.HasPrefix(localPart, ".") || strings.HasSuffix(localPart, ".") {
		return errors.New("local part cannot start or end with dot")
	}
	if strings.Contains(localPart, "..") {
		return errors.New("local part cannot have consecutive dots")
	}
	if len(domainPart) == 0 || len(domainPart) > 255 {
		return errors.New("invalid domain length")
	}
	if !strings.Contains(domainPart, ".") {
		return errors.New("domain must have valid TLD")
	}
	if !emailLocalPattern.MatchString(localPart) {
		return errors.New("invalid characters in local part")
	}
	if !emailDomainPattern.MatchString(domainPart) {
		return errors.New("invalid domain format")
	}
	return nil
}

func MaskEmail(input string) string {
	email := NormalizeEmail(input)
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return ""
	}

	localPart := parts[0]
	domainPart := parts[1]
	domainParts := strings.Split(domainPart, ".")
	if len(domainParts) < 2 {
		return ""
	}

	maskedLocalPart := string(localPart[0]) + "***"
	if len(localPart) > 1 {
		maskedLocalPart = string(localPart[0]) + "***" + string(localPart[len(localPart)-1])
	}

	maskedDomainPart := string(domainParts[0][0]) + "***"
	return maskedLocalPart + "@" + maskedDomainPart + "." + domainParts[len(domainParts)-1]
}

func (user User) PublicProfileVisibleTo(requestingUserID int64) bool {
	return user.Visibility == UserVisibilityPublic || user.ID == requestingUserID
}

func (user User) MaskedAccountEmail() string {
	return MaskEmail(user.AccountEmail)
}

func buildStringSet(values []string) map[string]struct{} {
	result := make(map[string]struct{}, len(values))
	for _, value := range values {
		result[value] = struct{}{}
	}
	return result
}
