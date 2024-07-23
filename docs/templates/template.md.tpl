{{- range .Files}}
{{- range .Messages}}
# --8<-- [start:{{.Name}}]
### {{ .Name }}

{{ .Description }}

| <div style="width:118px">Field Name</div> | Type | Description |
| --- | --- | --- |
{{- range .Fields }}
| `{{ .Name }}` | {{if eq .Label "repeated"}}[{{end}} {{if eq .Type "string" "bool"}}{{.Type}}{{else}}[{{ .Type }}](#{{.Type | lower}}){{end}} {{if eq .Label "repeated"}}]{{end}} {{if eq .Name "get" "post" "put" "delete" "patch"}}([RoutePattern](#routepattern)){{end}} | {{ .Description | replace "\n" "<br>" }} |
{{- end }}
# --8<-- [end:{{.Name}}]
{{- end }}
{{- end }}
