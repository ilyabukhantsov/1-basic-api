package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"1-basic-api/jwt"
)

func newHandler(t *testing.T) (http.Handler, *jwt.Manager) {
	t.Helper()
	m, err := jwt.NewManager([]byte("test-secret"), time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u, ok := UsernameFromContext(r.Context())
		if !ok {
			t.Error("username missing from context")
		}
		role, ok := RoleFromContext(r.Context())
		if !ok {
			t.Error("role missing from context")
		}
		w.Header().Set("X-User", u+":"+role)
		w.WriteHeader(http.StatusOK)
	})
	return Auth(m)(next), m
}

func TestAuthAllowsValidToken(t *testing.T) {
	h, m := newHandler(t)
	token, _ := m.GenerateToken("alice", "admin")

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body)
	}
	if got := rec.Header().Get("X-User"); got != "alice:admin" {
		t.Fatalf("unexpected user header %q", got)
	}
}

func TestAuthRejects(t *testing.T) {
	h, m := newHandler(t)
	other, _ := jwt.NewManager([]byte("other"), time.Minute)
	foreign, _ := other.GenerateToken("mallory", "admin")
	valid, _ := m.GenerateToken("alice", "admin")

	cases := map[string]string{
		"missing header": "",
		"no bearer":      valid,
		"wrong scheme":   "Basic " + valid,
		"empty token":    "Bearer ",
		"garbage token":  "Bearer abc.def.ghi",
		"foreign secret": "Bearer " + foreign,
	}
	for name, header := range cases {
		t.Run(name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if header != "" {
				req.Header.Set("Authorization", header)
			}
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)
			if rec.Code != http.StatusUnauthorized {
				t.Fatalf("expected 401, got %d", rec.Code)
			}
			if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
				t.Fatalf("expected JSON error, got %q", ct)
			}
		})
	}
}

func TestContextHelpersWithoutValues(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	if _, ok := UsernameFromContext(req.Context()); ok {
		t.Fatal("expected no username")
	}
	if _, ok := RoleFromContext(req.Context()); ok {
		t.Fatal("expected no role")
	}
}
