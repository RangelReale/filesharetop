{{template "header" .}}
	
Home<br/>

<div id="body">

	<div class="categories">
		<a href="/">ALL</a>
		{{range .categories}}
		<a href="/?category={{.}}">{{.}}</a>
		{{end}}
	</div>

{{.Body}}

	<div class="paging">
		{{ if .page_first }}
		<a href="{{ .page_first }}">First</a>
		{{ end }}
		{{ if .page_prev }}
		<a href="{{ .page_prev }}">Previous</a>
		{{ end }}
		<div>Page {{ .page }}</div>
		{{ if .page_next }}
		<a href="{{ .page_next }}">Next</a>
		{{ end }}
		{{ if .page_last }}
		<a href="{{ .page_last }}">Last</a>
		{{ end }}
	</div>

</div>
	
{{template "footer" .}}