package frontend

import (
	"html/template"
	"strings"
)

var Funcs template.FuncMap = template.FuncMap{
	"lower": strings.ToLower,
}
