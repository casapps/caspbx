package swagger

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewSpecAndHandler(t *testing.T) {
	spec := NewSpec("caspbx", "dev", "/api/v1")
	if spec.OpenAPI != "3.1.0" || spec.Title != "caspbx" {
		t.Fatalf("unexpected spec %+v", spec)
	}

	handlerValue := NewHandler(spec)

	getRequest := httptest.NewRequest(http.MethodGet, "/swagger.json", nil)
	getResponse := httptest.NewRecorder()
	handlerValue.ServeHTTP(getResponse, getRequest)
	if !strings.Contains(getResponse.Body.String(), "\"base_url\":\"/api/v1\"") {
		t.Fatalf("unexpected swagger response %q", getResponse.Body.String())
	}

	postRequest := httptest.NewRequest(http.MethodPost, "/swagger.json", nil)
	postResponse := httptest.NewRecorder()
	handlerValue.ServeHTTP(postResponse, postRequest)
	if postResponse.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected swagger post to fail, got %d", postResponse.Code)
	}
}
