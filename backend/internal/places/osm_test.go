package places

import (
	"encoding/json"
	"testing"
)

func TestOSMLabelsPreferDisplayName(t *testing.T) {
	var payload []osmPlace
	raw := []byte(`[
		{"display_name":"10, Abay Avenue, Almaty, Kazakhstan"},
		{"display_name":"10, Abay Avenue, Almaty, Kazakhstan"},
		{"display_name":"5, Dostyk Street, Astana, Kazakhstan"},
		{"display_name":"   "}
	]`)
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatal(err)
	}
	labels := osmLabels(payload)
	if len(labels) != 2 || labels[0] != "10, Abay Avenue, Almaty, Kazakhstan" || labels[1] != "5, Dostyk Street, Astana, Kazakhstan" {
		t.Fatalf("labels = %#v", labels)
	}
}

func TestSuggestShortQuerySkipsOSM(t *testing.T) {
	items, err := (OSM{}).Suggest(t.Context(), "ab")
	if err != nil || len(items) != 0 {
		t.Fatalf("items=%#v err=%v", items, err)
	}
}
