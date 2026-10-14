package store

import (
	"testing"

	"cargoprener/backend/internal/domain"
)

func TestProductSortIsWhitelisted(t *testing.T) {
	for _, sort := range []string{domain.SortNewest, domain.SortName, domain.SortPriceAsc, domain.SortPriceDesc} {
		if _, ok := productOrderSQL[sort]; !ok {
			t.Fatalf("missing sort %s", sort)
		}
	}
	if _, ok := productOrderSQL["name; DROP TABLE products"]; ok {
		t.Fatal("sort SQL accepted an unknown value")
	}
}

func TestEscapeLike(t *testing.T) {
	got := ilikeContains(`100%_wrap\`)
	if got != `%100\%\_wrap\\%` {
		t.Fatalf("pattern = %q", got)
	}
}
