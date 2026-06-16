package validator

import (
	"strings"
	"testing"
)

type TestInput struct {
	Name  string `json:"name" validate:"required,min=2,max=100"`
	Email string `json:"email" validate:"required,email"`
	Price int64  `json:"price" validate:"required,min=1"`
}

func TestValidate_ValidInput(t *testing.T) {
	v := New()

	input := TestInput{
		Name:  "Espresso",
		Email: "barista@coffee.test",
		Price: 25000,
	}

	if errs := v.Validate(input); errs != nil {
		t.Fatalf("expected nil, got %v", errs)
	}
}

func TestValidate_AllFieldsEmpty(t *testing.T) {
	v := New()

	errs := v.Validate(TestInput{})
	if errs == nil {
		t.Fatal("expected validation errors, got nil")
	}

	if len(errs) != 3 {
		t.Fatalf("expected 3 errors, got %d: %v", len(errs), errs)
	}

	for _, key := range []string{"name", "email", "price"} {
		if _, ok := errs[key]; !ok {
			t.Errorf("expected error for key %q, got keys: %v", key, errs)
		}
	}
}

func TestValidate_InvalidEmail(t *testing.T) {
	v := New()

	input := TestInput{
		Name:  "Latte",
		Email: "bukan-email",
		Price: 30000,
	}

	errs := v.Validate(input)
	if errs == nil {
		t.Fatal("expected validation errors, got nil")
	}

	msg, ok := errs["email"]
	if !ok {
		t.Fatalf("expected error for key \"email\", got keys: %v", errs)
	}

	if !strings.Contains(msg, "format email tidak valid") {
		t.Errorf("expected message to contain %q, got %q", "format email tidak valid", msg)
	}
}

func TestValidate_NameTooShort(t *testing.T) {
	v := New()

	input := TestInput{
		Name:  "A",
		Email: "barista@coffee.test",
		Price: 30000,
	}

	errs := v.Validate(input)
	if errs == nil {
		t.Fatal("expected validation errors, got nil")
	}

	msg, ok := errs["name"]
	if !ok {
		t.Fatalf("expected error for key \"name\", got keys: %v", errs)
	}

	if !strings.Contains(msg, "minimal 2 karakter") {
		t.Errorf("expected message to contain %q, got %q", "minimal 2 karakter", msg)
	}
}
