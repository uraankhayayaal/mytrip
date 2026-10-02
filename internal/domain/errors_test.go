package domain

import (
	"testing"
)

func TestAppError_Error(t *testing.T) {
	e := NewNotFound("trip not found")
	want := "not_found: trip not found"
	if got := e.Error(); got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}

func TestNewValidation(t *testing.T) {
	details := map[string]any{"field": "email"}
	e := NewValidation("invalid email", details)
	if e.Code != ErrCodeValidation {
		t.Errorf("Code = %q, want %q", e.Code, ErrCodeValidation)
	}
	if e.HTTPStatus != 400 {
		t.Errorf("HTTPStatus = %d, want 400", e.HTTPStatus)
	}
	if e.Message != "invalid email" {
		t.Errorf("Message = %q", e.Message)
	}
	if e.Details["field"] != "email" {
		t.Errorf("Details = %v", e.Details)
	}
}

func TestNewNotFound(t *testing.T) {
	e := NewNotFound("nope")
	if e.Code != ErrCodeNotFound || e.HTTPStatus != 404 {
		t.Errorf("got code=%q status=%d", e.Code, e.HTTPStatus)
	}
}

func TestNewUnauthorized(t *testing.T) {
	e := NewUnauthorized("no token")
	if e.Code != ErrCodeUnauthorized || e.HTTPStatus != 401 {
		t.Errorf("got code=%q status=%d", e.Code, e.HTTPStatus)
	}
}

func TestNewConflict(t *testing.T) {
	e := NewConflict("exists")
	if e.Code != ErrCodeConflict || e.HTTPStatus != 409 {
		t.Errorf("got code=%q status=%d", e.Code, e.HTTPStatus)
	}
}

func TestNewInternal(t *testing.T) {
	e := NewInternal()
	if e.Code != ErrCodeInternal || e.HTTPStatus != 500 {
		t.Errorf("got code=%q status=%d", e.Code, e.HTTPStatus)
	}
	if e.Message != "internal server error" {
		t.Errorf("Message = %q", e.Message)
	}
}
