{{- if eq .Name "main" -}}
	{{- header .Level .Dirname -}}
{{- else -}}
	{{- header .Level .Name -}}
{{- end -}}
{{- spacer -}}

{{- template "import" . -}}
{{- spacer -}}

**Note**: This file documents the use of this package when built with a explicit build tag to put it into "structured mode"

```
go build -tags=structuredlogs
```

[Documentation for the build without this tag is available here](./standard.md).
{{- spacer -}}

{{- if len .Doc.Blocks -}}
	{{- template "doc" .Doc -}}
	{{- spacer -}}
{{- end -}}

{{- range (iter .Examples) -}}
	{{- template "example" .Entry -}}
	{{- spacer -}}
{{- end -}}

{{- header (add .Level 1) "Index" -}}
{{- spacer -}}

{{- template "index" . -}}

{{- if len .Consts -}}
	{{- spacer -}}

	{{- header (add .Level 1) "Constants" -}}
	{{- spacer -}}

	{{- range (iter .Consts) -}}
		{{- template "value" .Entry -}}
		{{- if (not .Last) -}}{{- spacer -}}{{- end -}}
	{{- end -}}

{{- end -}}

{{- if len .Vars -}}
	{{- spacer -}}

	{{- header (add .Level 1) "Variables" -}}
	{{- spacer -}}

	{{- range (iter .Vars) -}}
		{{- template "value" .Entry -}}
		{{- if (not .Last) -}}{{- spacer -}}{{- end -}}
	{{- end -}}

{{- end -}}

{{- if len .Funcs -}}
	{{- spacer -}}

	{{- range (iter .Funcs) -}}
		{{- template "func" .Entry -}}
		{{- if (not .Last) -}}{{- spacer -}}{{- end -}}
	{{- end -}}
{{- end -}}

{{- if len .Types -}}
	{{- spacer -}}

	{{- range (iter .Types) -}}
		{{- template "type" .Entry -}}
		{{- if (not .Last) -}}{{- spacer -}}{{- end -}}
	{{- end -}}
{{- end -}}
