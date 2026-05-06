package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type RootResponse struct {
	Project      string `json:"project"`
	OfficialSite string `json:"official_site,omitempty"`
	AdminPath    string `json:"admin_path"`
	APIBasePath  string `json:"api_base_path"`
}

type SurfaceResponse struct {
	Surface string `json:"surface"`
	Path    string `json:"path"`
	Status  string `json:"status"`
}

func NewRootHandler(projectName string, officialSite string, adminPath string, apiBasePath string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !allowsReadOnlyMethod(w, r) {
			return
		}

		response := RootResponse{
			Project:      projectName,
			OfficialSite: officialSite,
			AdminPath:    adminPath,
			APIBasePath:  apiBasePath,
		}
		if prefersJSON(r.Header.Get("Accept")) {
			writeJSON(w, http.StatusOK, response)
			return
		}

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		fmt.Fprintf(w, "%s runtime scaffold is ready.\nAdmin path: %s\nAPI base path: %s\n", projectName, adminPath, apiBasePath)
	})
}

func NewAPIHandler(surface string, routePrefix string) http.Handler {
	return NewPlaceholderHandler(surface, routePrefix)
}

func NewPlaceholderHandler(surface string, routePrefix string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !allowsReadOnlyMethod(w, r) {
			return
		}

		response := SurfaceResponse{
			Surface: surface,
			Path:    routePrefix,
			Status:  "not_implemented",
		}
		if prefersText(r.URL.Path, r.Header.Get("Accept"), r.UserAgent()) {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(http.StatusNotImplemented)
			fmt.Fprintf(w, "%s surface is scaffolded at %s\n", surface, routePrefix)
			return
		}

		writeJSON(w, http.StatusNotImplemented, response)
	})
}

func allowsReadOnlyMethod(w http.ResponseWriter, r *http.Request) bool {
	switch r.Method {
	case http.MethodGet, http.MethodHead:
		return true
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		return false
	}
}

func writeJSON(w http.ResponseWriter, statusCode int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(value)
}

func prefersJSON(acceptHeader string) bool {
	return strings.Contains(acceptHeader, "application/json")
}

func prefersText(requestPath string, acceptHeader string, userAgent string) bool {
	switch {
	case strings.HasSuffix(requestPath, ".txt"):
		return true
	case strings.Contains(acceptHeader, "text/plain"):
		return true
	case strings.Contains(acceptHeader, "application/json"):
		return false
	}

	for _, cliUserAgent := range []string{"curl/", "Wget/", "HTTPie/", "python-requests/", "Go-http-client/", "node-fetch/"} {
		if strings.Contains(userAgent, cliUserAgent) {
			return true
		}
	}

	return userAgent == ""
}
