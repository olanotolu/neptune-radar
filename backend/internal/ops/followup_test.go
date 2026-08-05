package ops

import (
	"testing"
	"time"

	"neptune-social-radar/backend/internal/store"
)

func TestIsDueForFollowUp(t *testing.T) {
	past := time.Now().UTC().AddDate(0, 0, -20)
	future := time.Now().UTC().AddDate(0, 0, 5)
	mailedAt := time.Now().UTC().AddDate(0, 0, -20)
	sentAt := time.Now().UTC().AddDate(0, 0, -1)
	now := time.Now().UTC()

	tests := []struct {
		name string
		kit  store.CongratulateKit
		want bool
	}{
		{
			name: "due: follow_up_at in past, not sent, count 0, conf 0.60",
			kit: store.CongratulateKit{
				Status:            "mailed",
				FollowUpAt:        &past,
				FollowUpSentAt:    nil,
				FollowUpCount:     0,
				AddressConfidence: 0.60,
				MailedAt:          &mailedAt,
			},
			want: true,
		},
		{
			name: "not due: follow_up_at in future",
			kit: store.CongratulateKit{
				FollowUpAt:        &future,
				AddressConfidence: 0.60,
			},
			want: false,
		},
		{
			name: "not due: already sent (FollowUpSentAt set)",
			kit: store.CongratulateKit{
				FollowUpAt:        &past,
				FollowUpSentAt:    &sentAt,
				AddressConfidence: 0.60,
			},
			want: false,
		},
		{
			name: "not due: count >= 2 (max reached)",
			kit: store.CongratulateKit{
				FollowUpAt:        &past,
				FollowUpCount:     2,
				AddressConfidence: 0.60,
			},
			want: false,
		},
		{
			name: "not due: confidence below 0.50 floor",
			kit: store.CongratulateKit{
				FollowUpAt:        &past,
				AddressConfidence: 0.49,
			},
			want: false,
		},
		{
			name: "not due: FollowUpAt nil (no follow-up scheduled)",
			kit: store.CongratulateKit{
				AddressConfidence: 0.60,
			},
			want: false,
		},
		{
			name: "due: confidence exactly 0.50 (boundary)",
			kit: store.CongratulateKit{
				FollowUpAt:        &past,
				AddressConfidence: 0.50,
				MailedAt:          &mailedAt,
			},
			want: true,
		},
		{
			name: "due: count 1 (second follow-up allowed)",
			kit: store.CongratulateKit{
				FollowUpAt:        &past,
				FollowUpCount:     1,
				AddressConfidence: 0.55,
				MailedAt:          &mailedAt,
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsDueForFollowUp(tt.kit, now)
			if got != tt.want {
				t.Errorf("IsDueForFollowUp() = %v, want %v", got, tt.want)
			}
		})
	}
}
