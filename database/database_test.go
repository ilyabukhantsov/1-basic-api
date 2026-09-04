package database

import (
	"errors"
	"testing"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func open(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := Connect(Config{DSN: "file:" + t.Name() + "?mode=memory&cache=shared", AdminUsername: "admin", AdminPassword: "pw"})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = Close(db) })
	return db
}

func TestConnectSeedsAdminOnce(t *testing.T) {
	db := open(t)
	u, err := FindUserByUsername(db, "admin")
	if err != nil {
		t.Fatal(err)
	}
	if u.Role != "admin" || bcrypt.CompareHashAndPassword([]byte(u.Password), []byte("pw")) != nil {
		t.Fatalf("unexpected seeded user %+v", u)
	}
	// second seed with a different password must not overwrite the existing user
	if err := seedAdminUser(db, "admin", "other"); err != nil {
		t.Fatal(err)
	}
	u2, _ := FindUserByUsername(db, "admin")
	if u2.Password != u.Password {
		t.Fatal("seed overwrote existing user")
	}
	if _, err := FindUserByUsername(db, "nobody"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestConnectRejectsEmptyAdminPassword(t *testing.T) {
	if _, err := Connect(Config{DSN: ":memory:", AdminUsername: "admin"}); err == nil {
		t.Fatal("expected error")
	}
}

func TestConnectBadDSN(t *testing.T) {
	if _, err := Connect(Config{DSN: "/nonexistent/dir/x.db"}); err == nil {
		t.Fatal("expected error")
	}
}

func TestCategoryAndProductErrors(t *testing.T) {
	db := open(t)

	c, err := CreateCategory(db, "A")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := CreateCategory(db, "A"); !errors.Is(err, ErrConflict) {
		t.Fatalf("expected ErrConflict, got %v", err)
	}
	if _, err := GetCategory(db, 999); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
	if _, err := UpdateCategory(db, 999, "x"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
	if err := DeleteCategory(db, 999); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}

	p, err := CreateProduct(db, ProductInput{Name: "P", Code: "C", Price: 1, CategoryIDs: []uint{c.ID}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := CreateProduct(db, ProductInput{Name: "P2", Code: "C", Price: 1, CategoryIDs: []uint{c.ID}}); !errors.Is(err, ErrConflict) {
		t.Fatalf("expected ErrConflict, got %v", err)
	}
	if _, err := CreateProduct(db, ProductInput{Name: "P3", Code: "C3", Price: 1, CategoryIDs: []uint{999}}); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
	got, err := GetProduct(db, p.ID)
	if err != nil || len(got.CategoryIDs) != 1 {
		t.Fatalf("GetProduct: %v %+v", err, got)
	}
	if _, err := GetProduct(db, 999); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
	if _, err := UpdateProduct(db, 999, ProductInput{}); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
	if err := DeleteProduct(db, 999); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
	// product with no categories is allowed at the data layer
	if _, err := UpdateProduct(db, p.ID, ProductInput{Name: "P", Code: "C", Price: 2}); err != nil {
		t.Fatal(err)
	}
	if _, err := ListProductsByCategory(db, 999); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
