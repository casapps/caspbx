package swagger

import (
	"encoding/json"
	"net/http"
)

type Spec struct {
	OpenAPI string `json:"openapi"`
	Title   string `json:"title"`
	Version string `json:"version"`
	BaseURL string `json:"base_url"`
}

func NewSpec(projectName string, version string, baseURL string) Spec {
	return Spec{
		OpenAPI: "3.1.0",
		Title:   projectName,
		Version: version,
		BaseURL: baseURL,
	}
}

func NewHandler(spec Spec) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet, http.MethodHead:
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			_ = json.NewEncoder(w).Encode(spec)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
}
