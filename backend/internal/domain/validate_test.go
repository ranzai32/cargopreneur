package domain

import "testing"

func TestRegisterValidate(t *testing.T) {
	input := RegisterInput{
		BusinessName: "  Harbor Shop  ",
		ContactName:  "Ada Lopez",
		Email:        "Ada@Shop.Example",
		Password:     "long-enough",
		Region:       "Grand Rapids, MI",
	}
	input.Normalize()
	if input.Email != "ada@shop.example" {
		t.Fatalf("email = %q", input.Email)
	}
	if errs := input.Validate(); len(errs) != 0 {
		t.Fatalf("unexpected errors: %#v", errs)
	}

	input.Password = "short"
	input.Email = "not-an-email"
	input.Region = " "
	input.Normalize()
	errs := input.Validate()
	if errs["password"] == "" || errs["email"] == "" || errs["region"] == "" {
		t.Fatalf("errors = %#v", errs)
	}
}

func TestRegisterRequiresFields(t *testing.T) {
	errs := RegisterInput{}.Validate()
	for _, field := range []string{"businessName", "contactName", "email", "password", "region"} {
		if errs[field] == "" {
			t.Fatalf("missing error for %s: %#v", field, errs)
		}
	}
}

func TestLoginValidate(t *testing.T) {
	input := LoginInput{Email: "  ADA@shop.example ", Password: "secret"}
	input.Normalize()
	if input.Email != "ada@shop.example" {
		t.Fatalf("email = %q", input.Email)
	}
	if errs := input.Validate(); len(errs) != 0 {
		t.Fatalf("unexpected errors: %#v", errs)
	}
	if errs := (LoginInput{}).Validate(); errs["email"] == "" || errs["password"] == "" {
		t.Fatalf("errors = %#v", errs)
	}
}
