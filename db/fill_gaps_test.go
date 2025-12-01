package db

import (
	"testing"
	"time"

	"github.com/zackb/updog/pageview"
)

func TestFillGaps(t *testing.T) {
	start := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2023, 1, 1, 3, 0, 0, 0, time.UTC)
	step := func(t time.Time) time.Time {
		return t.Add(time.Hour)
	}

	// Case 1: Empty input
	t.Run("Empty input", func(t *testing.T) {
		var input []*pageview.AggregatedPoint
		filled := fillGaps(input, start, end, step)

		if len(filled) != 4 { // 0, 1, 2, 3
			t.Errorf("Expected 4 points, got %d", len(filled))
		}
		for i, p := range filled {
			expectedTime := start.Add(time.Duration(i) * time.Hour)
			if !p.Time.Equal(expectedTime) {
				t.Errorf("Point %d: expected time %v, got %v", i, expectedTime, p.Time)
			}
			if p.Count != 0 {
				t.Errorf("Point %d: expected count 0, got %d", i, p.Count)
			}
		}
	})

	// Case 2: Sparse input
	t.Run("Sparse input", func(t *testing.T) {
		input := []*pageview.AggregatedPoint{
			{Time: start, Count: 10},
			{Time: start.Add(2 * time.Hour), Count: 20},
		}
		filled := fillGaps(input, start, end, step)

		if len(filled) != 4 {
			t.Errorf("Expected 4 points, got %d", len(filled))
		}

		// 0:00 - present (10)
		if filled[0].Count != 10 {
			t.Errorf("Point 0: expected count 10, got %d", filled[0].Count)
		}
		// 1:00 - missing (0)
		if filled[1].Count != 0 {
			t.Errorf("Point 1: expected count 0, got %d", filled[1].Count)
		}
		// 2:00 - present (20)
		if filled[2].Count != 20 {
			t.Errorf("Point 2: expected count 20, got %d", filled[2].Count)
		}
		// 3:00 - missing (0)
		if filled[3].Count != 0 {
			t.Errorf("Point 3: expected count 0, got %d", filled[3].Count)
		}
	})
}
