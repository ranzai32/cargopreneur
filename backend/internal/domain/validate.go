package domain

import (
	"regexp"
	"strings"
)

const (
	minPasswordLength = 8
	maxPasswordLength = 72
)

var emailPattern = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)

type RegisterInput struct {
	BusinessName string `json:"businessName"`
	ContactName  string `json:"contactName"`
	Email        string `json:"email"`
	Password     string `json:"password"`
	Region       string `json:"region"`
}

type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (in *RegisterInput) Normalize() {
	in.BusinessName = strings.TrimSpace(in.BusinessName)
	in.ContactName = strings.TrimSpace(in.ContactName)
	in.Email = strings.ToLower(strings.TrimSpace(in.Email))
	in.Region = strings.TrimSpace(in.Region)
}

func (in *LoginInput) Normalize() {
	in.Email = strings.ToLower(strings.TrimSpace(in.Email))
}

func (in RegisterInput) Validate() map[string]string {
	errs := map[string]string{}
	if in.BusinessName == "" {
		errs["businessName"] = "Business name is required"
	}
	if in.ContactName == "" {
		errs["contactName"] = "Contact name is required"
	}
	if in.Email == "" {
		errs["email"] = "Email is required"
	} else if !emailPattern.MatchString(in.Email) {
		errs["email"] = "Enter a valid email address"
	}
	if in.Password == "" {
		errs["password"] = "Password is required"
	} else if len(in.Password) < minPasswordLength {
		errs["password"] = "Password must be at least 8 characters"
	} else if len(in.Password) > maxPasswordLength {
		errs["password"] = "Password must be at most 72 characters"
	}
	if in.Region == "" {
		errs["region"] = "Region is required"
	}
	return errs
}

func (in LoginInput) Validate() map[string]string {
	errs := map[string]string{}
	if in.Email == "" {
		errs["email"] = "Email is required"
	}
	if in.Password == "" {
		errs["password"] = "Password is required"
	}
	return errs
}
