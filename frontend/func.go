package frontend

import (
	"html/template"
	"strings"

	"github.com/zackb/updog/pageview"
)

var Funcs template.FuncMap = template.FuncMap{
	"lower": strings.ToLower,
	"mul": func(a, b float64) float64 {
		return a * b
	},
	"RenderSVG": pageview.RenderSVG,
}
