package places

import (
	"encoding/json"
	"testing"
)

func TestSuggestionLabelsPrefersFullAddress(t *testing.T) {
	var payload suggestResponse
	raw := []byte(`{"result":{"items":[
		{"full_name":"Short","address_name":"Street","full_address_name":"Almaty, Abay street, 10"},
		{"full_name":"Short","address_name":"Street","full_address_name":"Almaty, Abay street, 10"},
		{"full_name":"Astana, Dostyk, 5"},
		{"address_name":"   "}
	]}}`)
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatal(err)
	}
	labels := suggestionLabels(payload)
	if len(labels) != 2 || labels[0] != "Almaty, Abay street, 10" || labels[1] != "Astana, Dostyk, 5" {
		t.Fatalf("labels = %#v", labels)
	}
}

func TestSuggestWithoutKeyReturnsNothing(t *testing.T) {
	items, err := (TwoGIS{}).Suggest(t.Context(), "Abay 10")
	if err != nil || len(items) != 0 {
		t.Fatalf("items=%#v err=%v", items, err)
	}
}
