package pageview

import (
	"strings"
	"testing"
	"time"
)

func TestRenderSVG(t *testing.T) {
	now := time.Now()
	stats := []*AggregatedPoint{
		{
			Time:           now,
			Count:          100,
			UniqueVisitors: 50,
			BounceRate:     0.5,
		},
		{
			Time:           now.Add(time.Hour),
			Count:          200,
			UniqueVisitors: 150,
			BounceRate:     0.3,
		},
	}

	html := RenderSVG(stats)
	svg := string(html)

	if !strings.Contains(svg, "<svg") {
		t.Error("Expected SVG tag")
	}

	if !strings.Contains(svg, "class=\"bar-total\"") {
		t.Error("Expected bar-total class")
	}

	if !strings.Contains(svg, "class=\"bar-unique\"") {
		t.Error("Expected bar-unique class")
	}

	if !strings.Contains(svg, "Total Views:") {
		t.Error("Expected tooltip label 'Total Views:'")
	}
	if !strings.Contains(svg, ">100<") {
		t.Error("Expected tooltip value '100'")
	}
}
