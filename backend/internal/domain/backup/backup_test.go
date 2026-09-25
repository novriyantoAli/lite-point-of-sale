package backup

import (
	"testing"
	"time"
)

func TestKeep(t *testing.T) {
	now := time.Date(2026, time.January, 15, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name string
		at   time.Time
		days int
		keep bool
	}{
		{
			name: "a fresh backup is kept",
			at:   now,
			days: 7,
			keep: true,
		},
		{
			name: "a backup on the last day is kept",
			at:   now.Add(-7 * 24 * time.Hour),
			days: 7,
			keep: true,
		},
		{
			name: "a backup just past the window is pruned",
			at:   now.Add(-7*24*time.Hour - time.Second),
			days: 7,
			keep: false,
		},
		{
			name: "a clock skewed into the future is kept rather than pruned",
			at:   now.Add(24 * time.Hour),
			days: 7,
			keep: true,
		},
		{
			name: "a zero-day window keeps nothing",
			at:   now,
			days: 0,
			keep: false,
		},
		{
			name: "a negative window keeps nothing",
			at:   now,
			days: -3,
			keep: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := Keep(test.at, now, test.days); got != test.keep {
				t.Fatalf("Keep(%v, %v, %d) = %v, want %v", test.at, now, test.days, got, test.keep)
			}
		})
	}
}
