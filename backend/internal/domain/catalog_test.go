package domain

import (
	"net/url"
	"testing"
)

func TestParseProductQueryDefaults(t *testing.T) {
	query, errs := ParseProductQuery(url.Values{})
	if len(errs) != 0 {
		t.Fatalf("errors = %#v", errs)
	}
	if query.Sort != SortNewest || query.Page != 1 || query.PageSize != 12 {
		t.Fatalf("query = %#v", query)
	}
}

func TestParseProductQueryRejectsBadValues(t *testing.T) {
	_, errs := ParseProductQuery(url.Values{
		"sort":       {"cheapest"},
		"page":       {"0"},
		"pageSize":   {"51"},
		"supplierId": {"not-a-uuid"},
	})
	for _, field := range []string{"sort", "page", "pageSize", "supplierId"} {
		if errs[field] == "" {
			t.Fatalf("missing %s in %#v", field, errs)
		}
	}
}

func TestParseIncludeUnverified(t *testing.T) {
	value, message := ParseIncludeUnverified("")
	if value || message != "" {
		t.Fatalf("empty = %v %q", value, message)
	}
	value, message = ParseIncludeUnverified("true")
	if !value || message != "" {
		t.Fatalf("true = %v %q", value, message)
	}
	if _, message = ParseIncludeUnverified("maybe"); message == "" {
		t.Fatal("expected an error")
	}
}
