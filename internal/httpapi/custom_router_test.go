package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRouterWithParam(t *testing.T) {
	r := NewRouter()
	r.HandleFunc("GET", "/api/test/{id}", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("id") != "123" {
			t.Errorf("expected id 123, got %s", r.URL.Query().Get("id"))
		}
		w.Write([]byte("ok"))
	})
	req := httptest.NewRequest("GET", "/api/test/123", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}
