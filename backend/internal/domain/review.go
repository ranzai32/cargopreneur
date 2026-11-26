package domain

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

type Review struct {
	ID        uuid.UUID
	Author    string
	Rating    int
	Body      string
	CreatedAt time.Time
}

type ReviewJSON struct {
	ID        uuid.UUID `json:"id"`
	Author    string    `json:"author"`
	Rating    int       `json:"rating"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"createdAt"`
}

type ReviewBody struct {
	Rating int    `json:"rating"`
	Body   string `json:"body"`
}

func (r Review) JSON() ReviewJSON {
	return ReviewJSON{
		ID:        r.ID,
		Author:    r.Author,
		Rating:    r.Rating,
		Body:      r.Body,
		CreatedAt: r.CreatedAt.UTC(),
	}
}

func ParseReview(body ReviewBody) (int, string, map[string]string) {
	errs := map[string]string{}
	if body.Rating < 1 || body.Rating > 5 {
		errs["rating"] = "Choose a rating from 1 to 5"
	}
	text := strings.TrimSpace(body.Body)
	if text == "" {
		errs["body"] = "Write a short review"
	} else if len([]rune(text)) < 8 {
		errs["body"] = "Use at least 8 characters"
	} else if len([]rune(text)) > 500 {
		errs["body"] = "Use at most 500 characters"
	}
	if len(errs) > 0 {
		return 0, "", errs
	}
	return body.Rating, text, nil
}
