package controller

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// Exercise the public routes, including installations under a custom base path.
// Old clients must receive an empty array, never a contact link or logo content.
func TestLegacySponsorRoutesAreDisabled(t *testing.T) {
	for _, base := range []string{"", "/custom-base"} {
		t.Run(base, func(t *testing.T) {
			router := gin.New()
			NewIndexController(router.Group(base))
			w := httptest.NewRecorder()
			router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, base+"/sponsors", nil))
			if w.Code != http.StatusOK || w.Header().Get("Cache-Control") != "no-store" {
				t.Fatalf("list status=%d headers=%v", w.Code, w.Header())
			}
			var response struct {
				Success bool                       `json:"success"`
				Obj     map[string]json.RawMessage `json:"obj"`
			}
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				t.Fatal(err)
			}
			if !response.Success || len(response.Obj) != 1 || string(response.Obj["sponsors"]) != "[]" {
				t.Fatalf("unexpected sponsor response: %s", w.Body.String())
			}

			w = httptest.NewRecorder()
			router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, base+"/sponsors/logo/acme.png", nil))
			if w.Code != http.StatusNotFound || w.Body.Len() != 0 || w.Header().Get("Cache-Control") != "no-store" {
				t.Fatalf("logo status=%d headers=%v body=%q", w.Code, w.Header(), w.Body.String())
			}
		})
	}
}
