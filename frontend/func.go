package frontend

import (
	"html/template"
	"strings"
)

var Funcs template.FuncMap = template.FuncMap{
	"lower": strings.ToLower,
	"mul": func(a, b float64) float64 {
		return a * b
	},
}
