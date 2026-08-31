package contacts

import (
	"testing"
	"time"

	"github.com/coreaxissoftware/talkex_business/internal/database"
)

func mustPtr(t time.Time) *time.Time { return &t }

func TestEvaluateStage(t *testing.T) {
	now := time.Date(2026, 8, 31, 12, 0, 0, 0, time.UTC)

	cases := []struct {
		name string
		c    Contact
		want string
	}{
		{
			name: "opted-out is always churned",
			c:    Contact{OptedIn: false},
			want: StageChurned,
		},
		{
			name: "brand new, no inbound",
			c: Contact{
				OptedIn: true,
				Base:    database.Base{CreatedAt: now.Add(-3 * 24 * time.Hour)},
			},
			want: StageNew,
		},
		{
			name: "old, no inbound is dormant",
			c: Contact{
				OptedIn: true,
				Base:    database.Base{CreatedAt: now.Add(-40 * 24 * time.Hour)},
			},
			want: StageDormant,
		},
		{
			name: "inbound within 30d is active",
			c: Contact{
				OptedIn:       true,
				LastInboundAt: mustPtr(now.Add(-10 * 24 * time.Hour)),
			},
			want: StageActive,
		},
		{
			name: "inbound 45d ago is dormant",
			c: Contact{
				OptedIn:       true,
				LastInboundAt: mustPtr(now.Add(-45 * 24 * time.Hour)),
			},
			want: StageDormant,
		},
		{
			name: "inbound 120d ago is churned",
			c: Contact{
				OptedIn:       true,
				LastInboundAt: mustPtr(now.Add(-120 * 24 * time.Hour)),
			},
			want: StageChurned,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := EvaluateStage(&tc.c, now)
			if got != tc.want {
				t.Fatalf("got=%s want=%s", got, tc.want)
			}
		})
	}
}
