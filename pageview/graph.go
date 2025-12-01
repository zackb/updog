package pageview

import (
	"bytes"
	"fmt"
	"html/template"
	texttemplate "text/template"
)

func RenderSVG(stats []*AggregatedPoint) template.HTML {
	if len(stats) == 0 {
		return ""
	}

	const (
		barWidth      = 24
		gap           = 12
		maxH          = 150
		topPadding    = 40
		bottomPadding = 30
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
	height := maxH + topPadding + bottomPadding

	type Bar struct {
		X            int
		YTotal       int
		HeightTotal  int
		YUnique      int
		HeightUnique int
		TimeLabel    string
		FullDate     string
		Total        int64
		Unique       int64
		BounceRate   string
	}

	var bars []Bar
	for i, s := range stats {
		total := float64(s.Count)
		unique := float64(s.UniqueVisitors)

		totalH := int((total / maxVal) * maxH)
		uniqueH := int((unique / maxVal) * maxH)

		// Ensure at least 1px height if value > 0
		if total > 0 && totalH == 0 {
			totalH = 1
		}
		if unique > 0 && uniqueH == 0 {
			uniqueH = 1
		}

		x := gap + i*(barWidth+gap)
		yTotal := topPadding + maxH - totalH
		yUnique := topPadding + maxH - uniqueH

		bars = append(bars, Bar{
			X:            x,
			YTotal:       yTotal,
			HeightTotal:  totalH,
			YUnique:      yUnique,
			HeightUnique: uniqueH,
			TimeLabel:    s.Time.Format("15:04"),
			FullDate:     s.Time.Format("Jan 02, 2006 15:04"),
			Total:        s.Count,
			Unique:       s.UniqueVisitors,
			BounceRate:   fmt.Sprintf("%.1f%%", s.BounceRate*100),
		})
	}

	data := struct {
		Width  int
		Height int
		Bars   []Bar
	}{
		Width:  width,
		Height: height,
		Bars:   bars,
	}

	tmpl := texttemplate.Must(texttemplate.New("svg").Parse(svgTemplate))
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return template.HTML(fmt.Sprintf("<!-- Error rendering SVG: %v -->", err))
	}

	return template.HTML(buf.String())
}

const svgTemplate = `<svg width="{{.Width}}" height="{{.Height}}" xmlns="http://www.w3.org/2000/svg" class="analytics-graph">
    <style>
        .analytics-graph {
            font-family: 'Outfit', sans-serif;
        }
        .bar-total {
            fill: #58a6ff;
            opacity: 0.3;
            transition: opacity 0.2s;
        }
        .bar-unique {
            fill: #238636;
            opacity: 0.8;
            transition: opacity 0.2s;
        }
        .bar-group:hover .bar-total {
            opacity: 0.5;
        }
        .bar-group:hover .bar-unique {
            opacity: 1;
        }
        .axis-line {
            stroke: #30363d;
            stroke-width: 1;
        }
        .axis-label {
            fill: #8b949e;
            font-size: 10px;
            text-anchor: middle;
        }
        
        /* Tooltip */
        .tooltip-group {
            opacity: 0;
            pointer-events: none;
            transition: opacity 0.1s;
        }
        .bar-group:hover .tooltip-group {
            opacity: 1;
        }
        .tooltip-bg {
            fill: #161b22;
            stroke: #30363d;
            stroke-width: 1;
            filter: drop-shadow(0 4px 6px rgba(0,0,0,0.3));
        }
        .tooltip-text {
            fill: #c9d1d9;
            font-size: 12px;
        }
        .tooltip-header {
            fill: #ffffff;
            font-weight: 600;
        }
        .tooltip-value-total {
            fill: #58a6ff;
        }
        .tooltip-value-unique {
            fill: #238636;
        }
    </style>

    <!-- Grid lines could go here -->

    {{range .Bars}}
    <g class="bar-group">
        <!-- Invisible hit area for easier hovering -->
        <rect x="{{.X}}" y="0" width="24" height="100%" fill="transparent" />

        <!-- Total Views Bar (Background) -->
        <rect class="bar-total" x="{{.X}}" y="{{.YTotal}}" width="24" height="{{.HeightTotal}}" rx="2" />

        <!-- Unique Views Bar (Foreground) -->
        <rect class="bar-unique" x="{{.X}}" y="{{.YUnique}}" width="24" height="{{.HeightUnique}}" rx="2" />

        <!-- X Axis Label -->
        <text class="axis-label" x="{{.X}}" y="{{$.Height}}" dx="12" dy="-5">{{.TimeLabel}}</text>

        <!-- Tooltip -->
        <g class="tooltip-group" transform="translate({{if lt .X 100}}{{.X}}{{else}}{{if gt .X 500}}{{.X}}{{else}}{{.X}}{{end}}{{end}}, 0)">
             <!-- Logic to shift tooltip if too close to edge could be added, simplified here to center on bar -->
             <!-- Using a fixed position relative to the bar top or fixed at top of graph -->
             
             <g transform="translate({{if gt .X 400}}-130{{else}}10{{end}}, 10)">
                <rect class="tooltip-bg" x="0" y="0" width="140" height="85" rx="6" />
                
                <text class="tooltip-text tooltip-header" x="10" y="20">{{.FullDate}}</text>
                
                <text class="tooltip-text" x="10" y="40">Total Views:</text>
                <text class="tooltip-text tooltip-value-total" x="130" y="40" text-anchor="end">{{.Total}}</text>
                
                <text class="tooltip-text" x="10" y="60">Unique Visitors:</text>
                <text class="tooltip-text tooltip-value-unique" x="130" y="60" text-anchor="end">{{.Unique}}</text>
                
                <text class="tooltip-text" x="10" y="80">Bounce Rate:</text>
                <text class="tooltip-text" x="130" y="80" text-anchor="end">{{.BounceRate}}</text>
             </g>
        </g>
    </g>
    {{end}}
</svg>`
