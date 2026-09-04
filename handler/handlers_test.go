package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"1-basic-api/database"
	"1-basic-api/jwt"
	"1-basic-api/models"
)

type testAPI struct {
	t      *testing.T
	srv    http.Handler
	token  string
	tokens *jwt.Manager
}

func newTestAPI(t *testing.T) *testAPI {
	t.Helper()
	db, err := database.Connect(database.Config{
		DSN:           fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_")),
		AdminUsername: "admin",
		AdminPassword: "admin123",
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = database.Close(db) })

	tokens, err := jwt.NewManager([]byte("test-secret"), time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	token, err := tokens.GenerateToken("admin", "admin")
	if err != nil {
		t.Fatal(err)
	}
	return &testAPI{t: t, srv: NewRouter(db, tokens), token: token, tokens: tokens}
}

func (a *testAPI) do(method, path string, body any, auth bool) *httptest.ResponseRecorder {
	a.t.Helper()
	var buf bytes.Buffer
	if body != nil {
		switch b := body.(type) {
		case string:
			buf.WriteString(b)
		default:
			if err := json.NewEncoder(&buf).Encode(b); err != nil {
				a.t.Fatal(err)
			}
		}
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	if auth {
		req.Header.Set("Authorization", "Bearer "+a.token)
	}
	rec := httptest.NewRecorder()
	a.srv.ServeHTTP(rec, req)
	return rec
}

func decode[T any](t *testing.T, rec *httptest.ResponseRecorder) T {
	t.Helper()
	var v T
	if err := json.Unmarshal(rec.Body.Bytes(), &v); err != nil {
		t.Fatalf("decode %q: %v", rec.Body.String(), err)
	}
	return v
}

func expect(t *testing.T, rec *httptest.ResponseRecorder, code int) {
	t.Helper()
	if rec.Code != code {
		t.Fatalf("expected %d, got %d: %s", code, rec.Code, rec.Body.String())
	}
	if rec.Code != http.StatusNoContent {
		if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
			t.Fatalf("expected application/json, got %q", ct)
		}
	}
}

func (a *testAPI) createCategory(name string) models.Category {
	a.t.Helper()
	rec := a.do(http.MethodPost, "/categories", CategoryInput{Name: name}, true)
	expect(a.t, rec, http.StatusCreated)
	return decode[models.Category](a.t, rec)
}

func (a *testAPI) createProduct(name, code string, price float64, cats ...uint) models.Product {
	a.t.Helper()
	rec := a.do(http.MethodPost, "/products", map[string]any{
		"name": name, "code": code, "price": price, "categoryIds": cats,
	}, true)
	expect(a.t, rec, http.StatusCreated)
	return decode[models.Product](a.t, rec)
}

// --- auth ---

func TestLogin(t *testing.T) {
	api := newTestAPI(t)

	rec := api.do(http.MethodPost, "/auth/login", LoginInput{Username: "admin", Password: "admin123"}, false)
	expect(t, rec, http.StatusOK)
	body := decode[map[string]string](t, rec)
	claims, err := api.tokens.VerifyToken(body["token"])
	if err != nil {
		t.Fatalf("returned token is invalid: %v", err)
	}
	if claims.Username != "admin" || claims.Role != "admin" {
		t.Fatalf("unexpected claims %+v", claims)
	}

	cases := []struct {
		name string
		body any
		code int
	}{
		{"wrong password", LoginInput{Username: "admin", Password: "nope"}, http.StatusUnauthorized},
		{"unknown user", LoginInput{Username: "ghost", Password: "admin123"}, http.StatusUnauthorized},
		{"empty fields", LoginInput{}, http.StatusBadRequest},
		{"invalid json", "{not json", http.StatusBadRequest},
		{"unknown field", `{"username":"admin","password":"admin123","x":1}`, http.StatusBadRequest},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			expect(t, api.do(http.MethodPost, "/auth/login", tc.body, false), tc.code)
		})
	}
}

func TestProtectedRoutesRequireAuth(t *testing.T) {
	api := newTestAPI(t)
	routes := []struct{ method, path string }{
		{http.MethodPost, "/categories"},
		{http.MethodPut, "/categories/1"},
		{http.MethodDelete, "/categories/1"},
		{http.MethodPost, "/products"},
		{http.MethodPut, "/products/1"},
		{http.MethodDelete, "/products/1"},
	}
	for _, r := range routes {
		t.Run(r.method+" "+r.path, func(t *testing.T) {
			expect(t, api.do(r.method, r.path, map[string]any{}, false), http.StatusUnauthorized)
		})
	}
}

// --- categories ---

func TestCategoriesCRUD(t *testing.T) {
	api := newTestAPI(t)

	rec := api.do(http.MethodGet, "/categories", nil, false)
	expect(t, rec, http.StatusOK)
	if got := decode[[]models.Category](t, rec); len(got) != 0 {
		t.Fatalf("expected empty list, got %v", got)
	}
	if body := strings.TrimSpace(rec.Body.String()); body != "[]" {
		t.Fatalf("expected JSON array, got %s", body)
	}

	c := api.createCategory("  Electronics ")
	if c.ID == 0 || c.Name != "Electronics" {
		t.Fatalf("unexpected category %+v", c)
	}

	expect(t, api.do(http.MethodPost, "/categories", CategoryInput{Name: "Electronics"}, true), http.StatusConflict)

	rec = api.do(http.MethodPut, fmt.Sprintf("/categories/%d", c.ID), CategoryInput{Name: "Gadgets"}, true)
	expect(t, rec, http.StatusOK)
	if got := decode[models.Category](t, rec); got.Name != "Gadgets" || got.ID != c.ID {
		t.Fatalf("unexpected updated category %+v", got)
	}

	rec = api.do(http.MethodGet, "/categories", nil, false)
	expect(t, rec, http.StatusOK)
	if got := decode[[]models.Category](t, rec); len(got) != 1 || got[0].Name != "Gadgets" {
		t.Fatalf("unexpected list %+v", got)
	}

	expect(t, api.do(http.MethodDelete, fmt.Sprintf("/categories/%d", c.ID), nil, true), http.StatusNoContent)
	expect(t, api.do(http.MethodDelete, fmt.Sprintf("/categories/%d", c.ID), nil, true), http.StatusNotFound)
	expect(t, api.do(http.MethodPut, fmt.Sprintf("/categories/%d", c.ID), CategoryInput{Name: "x"}, true), http.StatusNotFound)
}

func TestCategoryValidation(t *testing.T) {
	api := newTestAPI(t)
	cases := []struct {
		name string
		body any
		code int
	}{
		{"empty name", CategoryInput{Name: "   "}, http.StatusBadRequest},
		{"too long", CategoryInput{Name: strings.Repeat("a", 129)}, http.StatusBadRequest},
		{"bad json", "{", http.StatusBadRequest},
		{"unknown field", `{"category":"x"}`, http.StatusBadRequest},
		{"two objects", `{"name":"a"}{"name":"b"}`, http.StatusBadRequest},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			expect(t, api.do(http.MethodPost, "/categories", tc.body, true), tc.code)
		})
	}
	for _, id := range []string{"abc", "0", "-1", "1.5"} {
		t.Run("bad id "+id, func(t *testing.T) {
			expect(t, api.do(http.MethodPut, "/categories/"+id, CategoryInput{Name: "x"}, true), http.StatusBadRequest)
			expect(t, api.do(http.MethodDelete, "/categories/"+id, nil, true), http.StatusBadRequest)
			expect(t, api.do(http.MethodGet, "/categories/"+id+"/products", nil, false), http.StatusBadRequest)
		})
	}
}

func TestRequestBodyTooLarge(t *testing.T) {
	api := newTestAPI(t)
	body := `{"name":"` + strings.Repeat("a", maxBodyBytes+1) + `"}`
	expect(t, api.do(http.MethodPost, "/categories", body, true), http.StatusRequestEntityTooLarge)
}

// --- products ---

func TestProductsCRUD(t *testing.T) {
	api := newTestAPI(t)
	c1 := api.createCategory("Laptops")
	c2 := api.createCategory("Sale")
	c3 := api.createCategory("Empty")

	p := api.createProduct("MacBook", "MB-1", 1999.99, c1.ID, c2.ID)
	if p.ID == 0 || len(p.CategoryIDs) != 2 || p.Price != 1999.99 {
		t.Fatalf("unexpected product %+v", p)
	}

	expect(t, api.do(http.MethodPost, "/products", map[string]any{
		"name": "Dup", "code": "MB-1", "price": 1, "categoryIds": []uint{c1.ID},
	}, true), http.StatusConflict)

	// listing by category
	rec := api.do(http.MethodGet, fmt.Sprintf("/categories/%d/products", c1.ID), nil, false)
	expect(t, rec, http.StatusOK)
	if got := decode[[]models.Product](t, rec); len(got) != 1 || got[0].Code != "MB-1" || len(got[0].CategoryIDs) != 2 {
		t.Fatalf("unexpected products in c1: %+v", got)
	}
	rec = api.do(http.MethodGet, fmt.Sprintf("/categories/%d/products", c3.ID), nil, false)
	expect(t, rec, http.StatusOK)
	if body := strings.TrimSpace(rec.Body.String()); body != "[]" {
		t.Fatalf("expected empty JSON array, got %s", body)
	}
	expect(t, api.do(http.MethodGet, "/categories/9999/products", nil, false), http.StatusNotFound)

	// update: move to c3 only, change price
	rec = api.do(http.MethodPut, fmt.Sprintf("/products/%d", p.ID), map[string]any{
		"name": "MacBook Pro", "code": "MB-2", "price": 2499, "categoryIds": []uint{c3.ID},
	}, true)
	expect(t, rec, http.StatusOK)
	up := decode[models.Product](t, rec)
	if up.Name != "MacBook Pro" || up.Code != "MB-2" || up.Price != 2499 || len(up.CategoryIDs) != 1 || up.CategoryIDs[0] != c3.ID {
		t.Fatalf("unexpected updated product %+v", up)
	}
	rec = api.do(http.MethodGet, fmt.Sprintf("/categories/%d/products", c1.ID), nil, false)
	expect(t, rec, http.StatusOK)
	if got := decode[[]models.Product](t, rec); len(got) != 0 {
		t.Fatalf("product should have left c1, got %+v", got)
	}

	// update with missing category
	expect(t, api.do(http.MethodPut, fmt.Sprintf("/products/%d", p.ID), map[string]any{
		"name": "x", "code": "x", "price": 1, "categoryIds": []uint{9999},
	}, true), http.StatusNotFound)
	expect(t, api.do(http.MethodPut, "/products/9999", map[string]any{
		"name": "x", "code": "x", "price": 1, "categoryIds": []uint{c1.ID},
	}, true), http.StatusNotFound)

	// deleting a category detaches products but keeps them
	expect(t, api.do(http.MethodDelete, fmt.Sprintf("/categories/%d", c3.ID), nil, true), http.StatusNoContent)
	expect(t, api.do(http.MethodGet, fmt.Sprintf("/categories/%d/products", c3.ID), nil, false), http.StatusNotFound)

	expect(t, api.do(http.MethodDelete, fmt.Sprintf("/products/%d", p.ID), nil, true), http.StatusNoContent)
	expect(t, api.do(http.MethodDelete, fmt.Sprintf("/products/%d", p.ID), nil, true), http.StatusNotFound)
}

func TestProductValidation(t *testing.T) {
	api := newTestAPI(t)
	c := api.createCategory("Cat")
	valid := func(mut func(m map[string]any)) map[string]any {
		m := map[string]any{"name": "P", "code": "C", "price": 1.5, "categoryIds": []uint{c.ID}}
		mut(m)
		return m
	}
	cases := []struct {
		name string
		body any
		code int
	}{
		{"missing name", valid(func(m map[string]any) { m["name"] = " " }), http.StatusBadRequest},
		{"missing code", valid(func(m map[string]any) { delete(m, "code") }), http.StatusBadRequest},
		{"missing price", valid(func(m map[string]any) { delete(m, "price") }), http.StatusBadRequest},
		{"negative price", valid(func(m map[string]any) { m["price"] = -1 }), http.StatusBadRequest},
		{"string price", valid(func(m map[string]any) { m["price"] = "1" }), http.StatusBadRequest},
		{"no categories", valid(func(m map[string]any) { m["categoryIds"] = []uint{} }), http.StatusBadRequest},
		{"zero category id", valid(func(m map[string]any) { m["categoryIds"] = []uint{0} }), http.StatusBadRequest},
		{"duplicate category", valid(func(m map[string]any) { m["categoryIds"] = []uint{c.ID, c.ID} }), http.StatusBadRequest},
		{"unknown category", valid(func(m map[string]any) { m["categoryIds"] = []uint{c.ID, 9999} }), http.StatusNotFound},
		{"unknown field", valid(func(m map[string]any) { m["extra"] = 1 }), http.StatusBadRequest},
		{"bad json", "[", http.StatusBadRequest},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			expect(t, api.do(http.MethodPost, "/products", tc.body, true), tc.code)
		})
	}
	expect(t, api.do(http.MethodDelete, "/products/abc", nil, true), http.StatusBadRequest)
}

func TestHealth(t *testing.T) {
	api := newTestAPI(t)
	expect(t, api.do(http.MethodGet, "/health", nil, false), http.StatusOK)
}
