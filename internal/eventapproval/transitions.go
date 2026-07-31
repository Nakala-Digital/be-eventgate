// Package eventapproval berisi aturan bisnis inti (state machine) untuk
// workflow approval event, dipisah dari handler HTTP supaya bisa di-unit-test
// tanpa perlu DB/HTTP sama sekali.
package eventapproval

import (
	"fmt"

	"be-eventgate/internal/models"
)

// AllowedTransitions adalah peta (map) yang mendefinisikan aturan pergerakan status event.
// Aturan transisi status ini secara ketat membatasi dari satu status tertentu hanya bisa
// berpindah ke status tujuan yang telah diizinkan. Berikut rutenya:
//
//	draft               -> pending_approval (saat panitia mensubmit event)
//	pending_approval    -> approved | rejected | revision_requested (keputusan dari reviewer)
//	revision_requested  -> pending_approval (saat panitia telah merevisi dan mensubmit ulang)
//	approved            -> published (saat super admin mempublikasikan event)
//	published           -> draft (saat event ditarik kembali / unpublish)
//	rejected            -> (status final, tidak ada transisi lanjutan yang diizinkan)
var AllowedTransitions = map[string][]string{
	models.EventStatusDraft:             {models.EventStatusPendingApproval},
	models.EventStatusPendingApproval:   {models.EventStatusApproved, models.EventStatusRejected, models.EventStatusRevisionRequested},
	models.EventStatusRevisionRequested: {models.EventStatusPendingApproval},
	models.EventStatusApproved:          {models.EventStatusPublished},
	models.EventStatusPublished:         {models.EventStatusDraft},
}

// CanTransition mengecek apakah perpindahan status from -> to diperbolehkan.
func CanTransition(from, to string) bool {
	for _, allowed := range AllowedTransitions[from] {
		if allowed == to {
			return true
		}
	}
	return false
}

// ErrInvalidTransition dikembalikan kalau transisi status tidak sah.
type ErrInvalidTransition struct {
	From string
	To   string
}

func (e *ErrInvalidTransition) Error() string {
	return fmt.Sprintf("cannot transition event from %q to %q", e.From, e.To)
}

// ValidateTransition mengembalikan error kalau from -> to tidak sah.
func ValidateTransition(from, to string) error {
	if !CanTransition(from, to) {
		return &ErrInvalidTransition{From: from, To: to}
	}
	return nil
}
