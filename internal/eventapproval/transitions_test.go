package eventapproval

import (
	"testing"

	"be-eventgate/internal/models"
)

func TestCanTransition_ValidPaths(t *testing.T) {
	cases := []struct{ from, to string }{
		{models.EventStatusDraft, models.EventStatusPendingApproval},
		{models.EventStatusPendingApproval, models.EventStatusApproved},
		{models.EventStatusPendingApproval, models.EventStatusRejected},
		{models.EventStatusPendingApproval, models.EventStatusRevisionRequested},
		{models.EventStatusRevisionRequested, models.EventStatusPendingApproval},
		{models.EventStatusApproved, models.EventStatusPublished},
		{models.EventStatusPublished, models.EventStatusDraft},
	}
	for _, c := range cases {
		if !CanTransition(c.from, c.to) {
			t.Errorf("expected transition %q -> %q to be valid", c.from, c.to)
		}
	}
}

func TestCanTransition_InvalidPaths(t *testing.T) {
	cases := []struct{ from, to string }{
		{models.EventStatusDraft, models.EventStatusApproved},            // lompat, harus lewat pending_approval
		{models.EventStatusDraft, models.EventStatusPublished},           // lompat jauh
		{models.EventStatusRejected, models.EventStatusPendingApproval},  // rejected = final
		{models.EventStatusRejected, models.EventStatusDraft},            // rejected = final
		{models.EventStatusApproved, models.EventStatusRejected},         // approved tidak bisa balik ke rejected
		{models.EventStatusPendingApproval, models.EventStatusPublished}, // harus lewat approved dulu
		{models.EventStatusPublished, models.EventStatusPublished},       // publish dua kali
	}
	for _, c := range cases {
		if CanTransition(c.from, c.to) {
			t.Errorf("expected transition %q -> %q to be INVALID", c.from, c.to)
		}
	}
}

func TestValidateTransition_ReturnsError(t *testing.T) {
	err := ValidateTransition(models.EventStatusDraft, models.EventStatusApproved)
	if err == nil {
		t.Fatal("expected an error for an invalid transition")
	}
	if _, ok := err.(*ErrInvalidTransition); !ok {
		t.Errorf("expected error type *ErrInvalidTransition, got %T", err)
	}
}

func TestValidateTransition_NilForValidPath(t *testing.T) {
	if err := ValidateTransition(models.EventStatusApproved, models.EventStatusPublished); err != nil {
		t.Errorf("expected no error for a valid transition, got: %v", err)
	}
}
