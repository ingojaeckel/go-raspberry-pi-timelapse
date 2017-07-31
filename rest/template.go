package rest

import (
	"fmt"
	"github.com/ingojaeckel/go-raspberry-pi-timelapse/admin"
	"github.com/ingojaeckel/go-raspberry-pi-timelapse/files"
	"html/template"
	"net/http"
	"net/url"
)

const HtmlTemplate = `
<!DOCTYPE html>
<html>
	<head>
		<meta charset="UTF-8">
		<title>Raspberry Pi Timelapse</title>
	</head>
	<body>
		<h1>Photos</h1>
		<a href="/archive">Download Archive ({{ .PhotosTotalSize }})<a/>

		<h2>Last Picture</h2>
		<a href="/file/last"><img src="/file/last" alt="last picture" width="300px" /></a>

		{{ .Screenshots }}

		<h1>Monitoring</h1>
		<table>
		<tr><td>Local Time</td><td>{{ .Time }}</td></tr>
		<tr><td>Free Disk Space</td><td>{{ .FreeDiskSpace }}</td></tr>
		<tr><td>Uptime</td><td>{{ .Uptime }}</td></tr>
		<tr><td>CPU Temperature</td><td>{{ .CpuTemperature }}</td></tr>
		<tr><td>GPU Temperature</td><td>{{ .GpuTemperature }}</td></tr>
		</table>

		<h1>Administration</h1>
		<a href="/file/delete">Delete existing photos</a>
		<a href="/admin/shutdown">Shutdown</a>
		<a href="/admin/restart">Restart</a>
	</body>
</html>`

func GetIndex(w http.ResponseWriter, _ *http.Request) {
	t, _ := template.New("index").Parse(HtmlTemplate)
	p := Page{
		Time:            getCommandHtml("/bin/date"),
		GpuTemperature:  getCommandHtml("/opt/vc/bin/vcgencmd", "measure_temp"),
		CpuTemperature:  getCommandHtml("/bin/cat", "/sys/class/thermal/thermal_zone0/temp"),
		Uptime:          getCommandHtml("/usr/bin/uptime"),
		FreeDiskSpace:   getCommandHtml("/bin/df", "-h"),
		Screenshots:     getScreenshotsHtml("timelapse-pictures"), // TODO pass in the folder name via the Timelapse type
		PhotosTotalSize: getTotalSizeHtml("timelapse-pictures"),
	}

	t.Execute(w, p)
}

func getScreenshotsHtml(folder string) template.HTML {
	files, _ := files.ListFiles(folder, true)
	screenshotsHtml := "<tr><td>Name</td><td>Date</td><td>Size</td><td></td></tr>"
	for _, f := range files {
		linkToName := fmt.Sprintf(`<a href="/file/%s">%s</a>`, url.PathEscape("timelapse-pictures/"+f.Name), f.Name)
		deleteLink := fmt.Sprintf(`<a href="foo">Delete</a>`)
		screenshotsHtml = screenshotsHtml + fmt.Sprintf("<tr><td>%s</td><td>%s</td><td>%s</td><td>%s</td></tr>", linkToName, f.ModTime, getHumanReadableSize(f.Bytes), deleteLink)
	}
	return template.HTML("<table>" + screenshotsHtml + "</table>")
}

func getTotalSizeHtml(folder string) template.HTML {
	files, _ := files.ListFiles(folder, true)
	sizeBytes := int64(0)
	for _, f := range files {
		sizeBytes = sizeBytes + f.Bytes
	}
	return template.HTML(getHumanReadableSize(sizeBytes))
}

func getHumanReadableSize(sizeInBytes int64) string {
	if sizeInBytes > 1024*1024 {
		// At least 1MB of files
		return fmt.Sprintf("%d MB", sizeInBytes/1024/1024)
	}
	if sizeInBytes > 1024 {
		// Size is between 1KB - 1MB
		return fmt.Sprintf("%d KB", sizeInBytes/1024)
	}
	// Size is between 0 - 1KB
	return fmt.Sprintf("%d bytes", sizeInBytes)
}

func getCommandHtml(cmd string, args ...string) template.HTML {
	output, err := admin.RunCommand(cmd, args...)
	if err != nil {
		return template.HTML(err.Error())
	}
	return template.HTML("<pre>" + output + "</pre>")
}

type Page struct {
	Time            template.HTML
	FreeDiskSpace   template.HTML
	CpuTemperature  template.HTML
	GpuTemperature  template.HTML
	Uptime          template.HTML
	Screenshots     template.HTML
	PhotosTotalSize template.HTML
}
