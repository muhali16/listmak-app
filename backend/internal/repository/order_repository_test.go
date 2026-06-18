package repository

import (
	"testing"

	"github.com/muhali16/listmak-service/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// newTestDB spins up an in-memory SQLite database with the Order schema migrated.
func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared&_pragma=foreign_keys(1)"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&models.Order{}); err != nil {
		t.Fatalf("failed to migrate Order schema: %v", err)
	}
	// Ensure a clean table for every test that shares the cached in-memory DB.
	if err := db.Exec("DELETE FROM orders").Error; err != nil {
		t.Fatalf("failed to clear orders: %v", err)
	}
	return db
}

func seedOrder(t *testing.T, db *gorm.DB, listmakID uint, name string) models.Order {
	t.Helper()
	o := models.Order{
		ListmakID:   listmakID,
		Name:        name,
		OrderDetail: "item",
		Price:       10000,
		Qty:         1,
	}
	if err := db.Create(&o).Error; err != nil {
		t.Fatalf("failed to seed order %q: %v", name, err)
	}
	return o
}

// TestUpdateOrdersPaidByName_NameMatching is the core requirement: name matching
// must be exact, case-insensitive, and whitespace-trimmed (LOWER(TRIM(...))),
// never a LIKE/substring match.
func TestUpdateOrdersPaidByName_NameMatching(t *testing.T) {
	tests := []struct {
		desc       string
		storedName string
		queryName  string
		wantMatch  bool
	}{
		{"identical", "Budi", "Budi", true},
		{"different capitalization", "Budi", "budi", true},
		{"all caps query", "Budi", "BUDI", true},
		{"stored has surrounding spaces", "  Budi  ", "Budi", true},
		{"query has surrounding spaces", "Budi", "   Budi   ", true},
		{"both differ in case and spaces", "  budi ", "BUDI", true},
		{"different name", "Budi", "Andi", false},
		{"substring must not match", "Budiman", "Budi", false},
		{"inner whitespace is significant", "Budi Santoso", "BudiSantoso", false},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			db := newTestDB(t)
			repo := NewOrderRepository(db)

			seeded := seedOrder(t, db, 1, tc.storedName)

			affected, err := repo.UpdateOrdersPaidByName(1, tc.queryName, true)
			if err != nil {
				t.Fatalf("UpdateOrdersPaidByName returned error: %v", err)
			}

			wantAffected := int64(0)
			if tc.wantMatch {
				wantAffected = 1
			}
			if affected != wantAffected {
				t.Fatalf("affected = %d, want %d", affected, wantAffected)
			}

			var got models.Order
			if err := db.First(&got, seeded.ID).Error; err != nil {
				t.Fatalf("reload order: %v", err)
			}
			if got.IsPaid != tc.wantMatch {
				t.Fatalf("IsPaid = %v, want %v", got.IsPaid, tc.wantMatch)
			}
		})
	}
}

// TestUpdateOrdersPaidByName_AllItemsOfNameUpdated verifies that every row
// belonging to a person (one person can have many items) is flipped, while
// other people's rows are left untouched.
func TestUpdateOrdersPaidByName_AllItemsOfNameUpdated(t *testing.T) {
	db := newTestDB(t)
	repo := NewOrderRepository(db)

	a1 := seedOrder(t, db, 1, "Budi")
	a2 := seedOrder(t, db, 1, " budi ") // same person, messy whitespace/case
	other := seedOrder(t, db, 1, "Andi")

	affected, err := repo.UpdateOrdersPaidByName(1, "BUDI", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if affected != 2 {
		t.Fatalf("affected = %d, want 2", affected)
	}

	for _, id := range []uint{a1.ID, a2.ID} {
		var o models.Order
		db.First(&o, id)
		if !o.IsPaid {
			t.Errorf("order %d should be paid", id)
		}
		if o.PaidAt == nil {
			t.Errorf("order %d PaidAt should be set when marked paid", id)
		}
	}

	var untouched models.Order
	db.First(&untouched, other.ID)
	if untouched.IsPaid {
		t.Errorf("Andi's order should not be affected")
	}
}

// TestUpdateOrdersPaidByName_PaidAtBehavior mirrors UpdateOrderPaidStatus:
// paid_at is set when marking paid and cleared when un-marking.
func TestUpdateOrdersPaidByName_PaidAtBehavior(t *testing.T) {
	db := newTestDB(t)
	repo := NewOrderRepository(db)

	o := seedOrder(t, db, 1, "Budi")

	if _, err := repo.UpdateOrdersPaidByName(1, "Budi", true); err != nil {
		t.Fatalf("mark paid: %v", err)
	}
	var paid models.Order
	db.First(&paid, o.ID)
	if !paid.IsPaid || paid.PaidAt == nil {
		t.Fatalf("expected paid with PaidAt set, got IsPaid=%v PaidAt=%v", paid.IsPaid, paid.PaidAt)
	}

	if _, err := repo.UpdateOrdersPaidByName(1, "Budi", false); err != nil {
		t.Fatalf("mark unpaid: %v", err)
	}
	var unpaid models.Order
	db.First(&unpaid, o.ID)
	if unpaid.IsPaid {
		t.Fatalf("expected IsPaid=false")
	}
	if unpaid.PaidAt != nil {
		t.Fatalf("expected PaidAt cleared, got %v", unpaid.PaidAt)
	}
}

// TestUpdateOrdersPaidByName_ScopedToListmak ensures the same name in a
// different listmak is not touched.
func TestUpdateOrdersPaidByName_ScopedToListmak(t *testing.T) {
	db := newTestDB(t)
	repo := NewOrderRepository(db)

	inScope := seedOrder(t, db, 1, "Budi")
	otherListmak := seedOrder(t, db, 2, "Budi")

	affected, err := repo.UpdateOrdersPaidByName(1, "Budi", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if affected != 1 {
		t.Fatalf("affected = %d, want 1", affected)
	}

	var other models.Order
	db.First(&other, otherListmak.ID)
	if other.IsPaid {
		t.Errorf("order in listmak 2 should not be affected")
	}

	var in models.Order
	db.First(&in, inScope.ID)
	if !in.IsPaid {
		t.Errorf("order in listmak 1 should be paid")
	}
}

// TestUpdateOrdersPaidByName_NoMatchReturnsZero ensures no match reports zero
// rows affected (the service layer maps this to a 404 rather than a silent OK).
func TestUpdateOrdersPaidByName_NoMatchReturnsZero(t *testing.T) {
	db := newTestDB(t)
	repo := NewOrderRepository(db)

	seedOrder(t, db, 1, "Budi")

	affected, err := repo.UpdateOrdersPaidByName(1, "Nonexistent", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if affected != 0 {
		t.Fatalf("affected = %d, want 0", affected)
	}
}
