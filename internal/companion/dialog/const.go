package dialog

const (
	tmplHistoryName = "tmpl-history"
	tmplHistoryText = `{{range .}}<tr hx-get={{printf "/dialog/%s" .ID }} hx-target="body">
	<td>{{.StartTime}}</td>
	<td>{{.FirstLine}}</td>
	</tr>{{end}}`

	tmplDialogName = "tmpl-dialog"
	// TODO add clear word button
	tmplDialogText = `<div id="word" style=""></div>
	<button hx-get="/proxy/woerter/abend" hx-target="#word">TEST</button><br>
	{{range .}}<tr>
	<td>{{.Role}}</td>
	<td>{{range .Words}}<button hx-get="/proxy/woerter/{{.}}" hx-target="#word" hx-swap="innerHTML">{{.}}</button> {{end}}</td>
	</tr>{{end}}`
)
