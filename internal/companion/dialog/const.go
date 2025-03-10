package dialog

const (
	tmplHistoryName = "tmpl-history"
	tmplHistoryText = `{{range .}}<tr hx-get={{printf "/dialog/%s" .ID }} hx-target="body">
	<td>{{.StartTime}}</td>
	<td>{{.FirstLine}}</td>
	</tr>{{end}}`

	tmplDialogName = "tmpl-dialog"
	tmplDialogText = `{{range .}}<tr>
	<td>{{.Role}}</td>
	<td>{{.Content}}</td>
	</tr>{{end}}`
)
