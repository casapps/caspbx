package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/casapps/caspbx/src/server/model"
	"github.com/casapps/caspbx/src/server/service"
)

type SessionCookieConfig struct {
	Name      string
	Path      string
	HTTPOnly  bool
	Secure    string
	SameSite  http.SameSite
	MaxAge    time.Duration
}

type AuthHandler struct {
	routePrefix      string
	authService      service.AuthService
	userCookie       SessionCookieConfig
	registrationMode model.RegistrationMode
}

type authLoginRequest struct {
	Identifier string `json:"identifier"`
	Username   string `json:"username"`
	Password   string `json:"password"`
}

func NewAuthHandler(routePrefix string, authService service.AuthService, userCookie SessionCookieConfig, registrationMode model.RegistrationMode) http.Handler {
	return AuthHandler{
		routePrefix:      routePrefix,
		authService:      authService,
		userCookie:       userCookie,
		registrationMode: registrationMode,
	}
}

func (handler AuthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch routeTail(handler.routePrefix, r.URL.Path) {
	case "", "/":
		handler.writeOverview(w, r)
	case "login":
		handler.handleLogin(w, r)
	case "register":
		handler.handleRegister(w, r)
	case "logout":
		handler.handleLogout(w, r)
	case "refresh":
		handler.handleRefresh(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (handler AuthHandler) writeOverview(w http.ResponseWriter, r *http.Request) {
	if !allowsReadOnlyMethod(w, r) {
		return
	}

	response := map[string]string{
		"scope":             "auth",
		"login_path":        joinPath(handler.routePrefix, "login"),
		"register_path":     joinPath(handler.routePrefix, "register"),
		"logout_path":       joinPath(handler.routePrefix, "logout"),
		"registration_mode": string(handler.registrationMode),
	}
	if prefersJSON(r.Header.Get("Accept")) {
		writeJSON(w, http.StatusOK, response)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprintf(w, "Auth routes are active.\nLogin: %s\nRegister: %s\nRegistration mode: %s\n", response["login_path"], response["register_path"], response["registration_mode"])
}

func (handler AuthHandler) handleLogin(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet, http.MethodHead:
		response := map[string]string{
			"route":    joinPath(handler.routePrefix, "login"),
			"method":   "POST",
			"identity": "username_or_email",
		}
		if prefersJSON(r.Header.Get("Accept")) {
			writeJSON(w, http.StatusOK, response)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		fmt.Fprintf(w, "POST %s with identifier and password.\n", response["route"])
	case http.MethodPost:
		requestBody, parseError := readAuthLoginRequest(r)
		if parseError != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid login request"})
			return
		}

		user, issuedSession, authError := handler.authService.AuthenticateUser(r.Context(), firstNonEmpty(requestBody.Identifier, requestBody.Username), requestBody.Password, requestIP(r), r.UserAgent())
		if authError != nil {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
			return
		}

		http.SetCookie(w, handler.userCookie.Build(r, issuedSession.Token, issuedSession.Session.ExpiresAt))
		writeJSON(w, http.StatusOK, map[string]any{
			"status":      "authenticated",
			"username":    user.Username,
			"session_id":  issuedSession.Session.ID,
			"expires_at":  issuedSession.Session.ExpiresAt.UTC().Format(time.RFC3339),
			"rehash_hint": issuedSession.RehashHint,
		})
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (handler AuthHandler) handleRegister(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet, http.MethodHead:
		writeJSON(w, http.StatusOK, map[string]string{
			"route":             joinPath(handler.routePrefix, "register"),
			"registration_mode": string(handler.registrationMode),
		})
	case http.MethodPost:
		if handler.registrationMode != model.RegistrationModePublic {
			writeJSON(w, http.StatusForbidden, map[string]string{"error": service.ErrRegistrationRestricted.Error()})
			return
		}
		writeJSON(w, http.StatusNotImplemented, map[string]string{"error": "registration flow not implemented yet"})
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (handler AuthHandler) handleLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	sessionToken, tokenError := readSessionCookie(r, handler.userCookie.Name)
	if tokenError != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "authentication required"})
		return
	}
	if logoutError := handler.authService.LogoutUser(r.Context(), sessionToken); logoutError != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "authentication required"})
		return
	}

	http.SetCookie(w, handler.userCookie.Clear(r))
	writeJSON(w, http.StatusOK, map[string]string{"status": "logged_out"})
}

func (handler AuthHandler) handleRefresh(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	sessionToken, tokenError := readSessionCookie(r, handler.userCookie.Name)
	if tokenError != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "authentication required"})
		return
	}

	session, resolveError := handler.authService.ResolveUserSession(r.Context(), sessionToken)
	if resolveError != nil {
		http.SetCookie(w, handler.userCookie.Clear(r))
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "authentication required"})
		return
	}

	http.SetCookie(w, handler.userCookie.Build(r, sessionToken, session.ExpiresAt))
	writeJSON(w, http.StatusOK, map[string]string{
		"status":     "refreshed",
		"session_id": session.ID,
		"expires_at": session.ExpiresAt.UTC().Format(time.RFC3339),
	})
}

func readAuthLoginRequest(r *http.Request) (authLoginRequest, error) {
	if strings.Contains(r.Header.Get("Content-Type"), "application/json") {
		var requestBody authLoginRequest
		if decodeError := json.NewDecoder(r.Body).Decode(&requestBody); decodeError != nil {
			return authLoginRequest{}, decodeError
		}
		return requestBody, nil
	}

	if parseError := r.ParseForm(); parseError != nil {
		return authLoginRequest{}, parseError
	}
	return authLoginRequest{
		Identifier: r.FormValue("identifier"),
		Username:   r.FormValue("username"),
		Password:   r.FormValue("password"),
	}, nil
}

func readSessionCookie(r *http.Request, cookieName string) (string, error) {
	cookie, cookieError := r.Cookie(cookieName)
	if cookieError != nil || strings.TrimSpace(cookie.Value) == "" {
		return "", http.ErrNoCookie
	}
	return cookie.Value, nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func requestIP(r *http.Request) string {
	forwardedFor := strings.TrimSpace(r.Header.Get("X-Forwarded-For"))
	if forwardedFor != "" {
		return strings.Split(forwardedFor, ",")[0]
	}
	return r.RemoteAddr
}

func routeTail(routePrefix string, requestPath string) string {
	trimmed := strings.TrimPrefix(requestPath, routePrefix)
	trimmed = strings.Trim(strings.TrimSpace(trimmed), "/")
	return trimmed
}

func joinPath(parts ...string) string {
	segments := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.Trim(strings.TrimSpace(part), "/")
		if trimmed != "" {
			segments = append(segments, trimmed)
		}
	}
	if len(segments) == 0 {
		return "/"
	}
	return "/" + strings.Join(segments, "/")
}

func (configValue SessionCookieConfig) Build(r *http.Request, token string, expiresAt time.Time) *http.Cookie {
	return &http.Cookie{
		Name:     configValue.Name,
		Value:    token,
		Path:     configValue.Path,
		Expires:  expiresAt,
		HttpOnly: configValue.HTTPOnly,
		Secure:   configValue.isSecure(r),
		SameSite: configValue.SameSite,
	}
}

func (configValue SessionCookieConfig) Clear(r *http.Request) *http.Cookie {
	return &http.Cookie{
		Name:     configValue.Name,
		Value:    "",
		Path:     configValue.Path,
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
		HttpOnly: configValue.HTTPOnly,
		Secure:   configValue.isSecure(r),
		SameSite: configValue.SameSite,
	}
}

func (configValue SessionCookieConfig) isSecure(r *http.Request) bool {
	switch configValue.Secure {
	case "always":
		return true
	case "never":
		return false
	default:
		return r.TLS != nil || strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https")
	}
}
