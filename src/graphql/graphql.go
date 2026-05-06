package graphql

import (
	"encoding/json"
	"net/http"
)

type Schema struct {
	Project string `json:"project"`
	Status  string `json:"status"`
	Query   string `json:"query"`
}

func DefaultSchema(projectName string) Schema {
	return Schema{
		Project: projectName,
		Status:  "scaffold",
		Query:   "type Query { health: String! }",
	}
}

func NewHandler(schema Schema) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet, http.MethodHead:
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			_ = json.NewEncoder(w).Encode(schema)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
}
