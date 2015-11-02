package xhub

import (
	"html/template"
	"os"
)

const STATIC_SRC = "/src/github.com/joyrexus/xhub/static"

var STATIC = os.Getenv("GOPATH") + STATIC_SRC
var templates = template.Must(template.ParseFiles(
    STATIC + "/study_view.html",
    STATIC + "/study_edit.html",
))
