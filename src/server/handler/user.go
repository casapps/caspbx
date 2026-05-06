package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/casapps/caspbx/src/server/service"
)

type UserHandler struct {
	routePrefix string
	authService service.AuthService
	userCookie  SessionCookieConfig
}

func NewUserHandler(routePrefix string, authService service.AuthService, userCookie SessionCookieConfig) http.Handler {
	return UserHandler{
		routePrefix: routePrefix,
		authService: authService,
		userCookie:  userCookie,
	}
}

func (handler UserHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
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

	user, userError := handler.authService.FindUserByID(r.Context(), session.SubjectID)
	if userError != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "authentication required"})
		return
	}

	switch routeTail(handler.routePrefix, r.URL.Path) {
	case "", "profile":
		if prefersJSON(r.Header.Get("Accept")) {
			writeJSON(w, http.StatusOK, map[string]any{
				"username":      user.Username,
				"display_name":  user.DisplayName,
				"account_email": user.AccountEmail,
				"visibility":    user.Visibility,
				"session_id":    session.ID,
				"expires_at":    session.ExpiresAt.UTC().Format(time.RFC3339),
			})
			return
		}

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		fmt.Fprintf(w, "User: %s\nEmail: %s\nSession expires: %s\n", user.Username, user.AccountEmail, session.ExpiresAt.UTC().Format(time.RFC3339))
	default:
		http.NotFound(w, r)
	}
}
