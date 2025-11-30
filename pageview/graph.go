package pageview

import (
	"fmt"
	"html/template"
)

func RenderSVG(stats []*AggregatedPoint) template.HTML {
	if len(stats) == 0 {
		return ""
	}

	const (
		barWidth   = 20
		gap        = 10
		maxH       = 120
		topPadding = 60
	)

	// Find max count to scale bars
	var maxVal float64
	for _, s := range stats {
		if float64(s.Count) > maxVal {
			maxVal = float64(s.Count)
		}
	}
	if maxVal == 0 {
		maxVal = 1
	}

	width := len(stats)*(barWidth+gap) + gap

	svg := fmt.Sprintf(`<svg width="%d" height="%d" xmlns="http://www.w3.org/2000/svg">`, width, maxH+40)

	// Add embedded CSS for tooltip
	svg += `
    <style>
    .bar-group .tooltip {
        display: none;
        font-size: 12px;
        pointer-events: none;
        text-anchor: middle;
        fill: black;
    }
    .bar-group:hover .tooltip {
        display: block;
    }
    .tooltip-bg {
        fill: white;
        stroke: black;
        stroke-width: 0.5;
        rx: 3;
        ry: 3;
    }
    </style>
    `

	for i, s := range stats {
		total := float64(s.Count)
		unique := float64(s.UniqueVisitors)

		totalH := int((total / maxVal) * maxH)
		uniqueH := int((unique / maxVal) * maxH)

		x := gap + i*(barWidth+gap)

		yTotal := topPadding + maxH - totalH
		yUnique := topPadding + maxH - uniqueH
		tooltipY := yTotal - 10
		tooltipX := x + barWidth/2

		svg += `<g class="bar-group">`

		// TOTAL (base bar)
		svg += fmt.Sprintf(`<rect x="%d" y="%d" width="%d" height="%d" fill="#ccc"/>`, x, yTotal, barWidth, totalH)

		// UNIQUE (stacked bar)
		svg += fmt.Sprintf(`<rect x="%d" y="%d" width="%d" height="%d" fill="#4caf50"/>`, x, yUnique, barWidth, uniqueH)

		// Tooltip background (rect)
		svg += fmt.Sprintf(
			`<rect class="tooltip-bg tooltip" x="%d" y="%d" width="110" height="50"/>`,
			tooltipX-55, tooltipY-45,
		)

		// Tooltip text (multi-line)
		lines := []string{
			s.Time.Format("2006-01-02 15:04"),
			fmt.Sprintf("Pageviews: %d", s.Count),
			fmt.Sprintf("Unique: %d", s.UniqueVisitors),
			fmt.Sprintf("Bounce: %.1f%%", s.BounceRate*100),
		}

		for j, line := range lines {
			svg += fmt.Sprintf(
				`<text class="tooltip" x="%d" y="%d">%s</text>`,
				tooltipX,
				tooltipY-30+j*12,
				line,
			)
		}

		// Label below
		svg += fmt.Sprintf(
			`<text x="%d" y="%d" font-size="10" text-anchor="middle">%s</text>`,
			x+barWidth/2, maxH+12,
			s.Time.Format("15:04"),
		)

		svg += `</g>`
	}

	svg += `</svg>`
	return template.HTML(svg)
}
