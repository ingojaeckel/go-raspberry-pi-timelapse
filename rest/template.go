package rest

import (
	"html/template"
	"net/http"
)

const HtmlTemplate = `
<!DOCTYPE html>
<html>
	<head>
		<meta charset="UTF-8">
		<title>Raspberry Pi Timelapse</title>
	</head>
	<body>
		<h1>hello world</h1>
	</body>
</html>`

func GetIndex(w http.ResponseWriter, _ *http.Request) {
	t, _ := template.New("index").Parse(HtmlTemplate)
	t.Execute(w, nil)
}
