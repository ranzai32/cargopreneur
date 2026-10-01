package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

const (
	RoleBusiness = "business"
	RoleAdmin    = "admin"
)

var (
	ErrNotFound    = errors.New("not found")
	ErrEmailTaken  = errors.New("email taken")
)

type User struct {
	ID           uuid.UUID
	BusinessName string
	ContactName  string
	Email        string
	PasswordHash string
	Region       string
	Role         string
	CreatedAt    time.Time
}

type NewUser struct {
	BusinessName string
	ContactName  string
	Email        string
	PasswordHash string
	Region       string
	Role         string
}

type UserJSON struct {
	ID           uuid.UUID `json:"id"`
	BusinessName string    `json:"businessName"`
	ContactName  string    `json:"contactName"`
	Email        string    `json:"email"`
	Region       string    `json:"region"`
	Role         string    `json:"role"`
	CreatedAt    time.Time `json:"createdAt"`
}

func (u User) JSON() UserJSON {
	return UserJSON{
		ID:           u.ID,
		BusinessName: u.BusinessName,
		ContactName:  u.ContactName,
		Email:        u.Email,
		Region:       u.Region,
		Role:         u.Role,
		CreatedAt:    u.CreatedAt,
	}
}
