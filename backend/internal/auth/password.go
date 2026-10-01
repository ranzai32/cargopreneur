package auth

import (
	"sync"

	"golang.org/x/crypto/bcrypt"
)

type Hasher struct {
	cost  int
	dummy []byte
	once  sync.Once
	err   error
}

func NewHasher() *Hasher {
	return &Hasher{cost: bcrypt.DefaultCost}
}

func NewTestHasher() *Hasher {
	return &Hasher{cost: bcrypt.MinCost}
}

func (h *Hasher) Hash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), h.cost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func (h *Hasher) Compare(hash, password string) error {
	if hash == "" {
		dummy, err := h.placeholder()
		if err != nil {
			return err
		}
		return bcrypt.CompareHashAndPassword(dummy, []byte(password))
	}
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

func (h *Hasher) placeholder() ([]byte, error) {
	h.once.Do(func() {
		h.dummy, h.err = bcrypt.GenerateFromPassword([]byte("not-a-real-password"), h.cost)
	})
	return h.dummy, h.err
}
