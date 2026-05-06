package handler

import (
	"fmt"
	"net/http"
)

type HealthResponse struct {
	Status                string `json:"status"`
	Project               string `json:"project"`
	Version               string `json:"version"`
	CommitID              string `json:"commit_id"`
	APIBasePath           string `json:"api_base_path"`
	AdminPath             string `json:"admin_path"`
	AsteriskAdminPath     string `json:"asterisk_admin_path"`
	OfficialSite          string `json:"official_site,omitempty"`
	RuntimeImplementation string `json:"runtime_implementation"`
}

func NewHealthHandler(response HealthResponse) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !allowsReadOnlyMethod(w, r) {
			return
		}

		if prefersText(r.URL.Path, r.Header.Get("Accept"), r.UserAgent()) {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			fmt.Fprintf(w, "status=%s\nproject=%s\nversion=%s\nadmin_path=%s\napi_base_path=%s\n", response.Status, response.Project, response.Version, response.AdminPath, response.APIBasePath)
			return
		}

		writeJSON(w, http.StatusOK, response)
	})
}

func NewVersionHandler(projectName string, version string, commitID string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !allowsReadOnlyMethod(w, r) {
			return
		}

		if prefersText(r.URL.Path, r.Header.Get("Accept"), r.UserAgent()) {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			fmt.Fprintf(w, "%s %s (%s)\n", projectName, version, commitID)
			return
		}

		writeJSON(w, http.StatusOK, map[string]string{
			"project": projectName,
			"version": version,
			"commit":  commitID,
		})
	})
}
