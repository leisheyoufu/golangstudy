package main

import (
	"html/template"
	"os"
)

type ST struct {
	Status []StatusST
}

type StatusST struct {
	Status string `json:"status"`
	Num    int    `json:"sum"`
}

func main() {
	e := ST{
		Status: []StatusST{
			{Status: "ok", Num: 1},
			{Status: "error", Num: 2},
		},
	}

	tpl, err := template.New("test").Parse(`
{{ range .Status }}
	{{ $color := "red" }}
	{{ if eq .Status "ok" }} {{ $color = "green" }} {{end}}
	{{ if eq .Status "ok" }}<span style="color:{{ $color }}">{{.Num}}</span>
	{{ else if eq .Status "error" }}<span style="color:{{ $color }}">{{.Num}}</span>
	{{ end }}
{{ end }}`)
	if err != nil {
		panic(err)
	}
	err = tpl.Execute(os.Stdout, e)
	if err != nil {
		panic(err)
	}
}
