package domain

import (
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

type SavedAddress struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Label     string
	Line      string
	UpdatedAt time.Time
}

type SavedAddressJSON struct {
	ID        uuid.UUID `json:"id"`
	Label     string    `json:"label"`
	Line      string    `json:"line"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type AddressBody struct {
	Label string `json:"label"`
	Line  string `json:"line"`
}

func (a SavedAddress) JSON() SavedAddressJSON {
	return SavedAddressJSON{
		ID:        a.ID,
		Label:     a.Label,
		Line:      a.Line,
		UpdatedAt: a.UpdatedAt.UTC(),
	}
}

func ParseAddress(body AddressBody) (string, string, map[string]string) {
	label := strings.TrimSpace(body.Label)
	line := strings.TrimSpace(body.Line)
	errs := map[string]string{}
	if label == "" {
		label = "Address"
	}
	if utf8.RuneCountInString(label) > 40 {
		errs["label"] = "Use a shorter name"
	}
	if utf8.RuneCountInString(line) < 8 {
		errs["line"] = "Enter a full street address"
	}
	if utf8.RuneCountInString(line) > 240 {
		errs["line"] = "Use a shorter address"
	}
	if len(errs) > 0 {
		return "", "", errs
	}
	return label, line, nil
}
