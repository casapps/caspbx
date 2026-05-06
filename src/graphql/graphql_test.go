package graphql

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDefaultSchemaAndHandler(t *testing.T) {
	schema := DefaultSchema("caspbx")
	if schema.Project != "caspbx" || schema.Status != "scaffold" {
		t.Fatalf("unexpected schema %+v", schema)
	}

	handlerValue := NewHandler(schema)

	getRequest := httptest.NewRequest(http.MethodGet, "/graphql", nil)
	getResponse := httptest.NewRecorder()
	handlerValue.ServeHTTP(getResponse, getRequest)
	if !strings.Contains(getResponse.Body.String(), "\"query\":\"type Query { health: String! }\"") {
		t.Fatalf("unexpected graphql response %q", getResponse.Body.String())
	}

	postRequest := httptest.NewRequest(http.MethodPost, "/graphql", nil)
	postResponse := httptest.NewRecorder()
	handlerValue.ServeHTTP(postResponse, postRequest)
	if postResponse.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected graphql post to fail, got %d", postResponse.Code)
	}
}
