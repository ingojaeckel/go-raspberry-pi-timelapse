package static

const Html = `
<!DOCTYPE html>
<html>
	<head>
		<meta charset="UTF-8">
		<title>Raspberry Pi Timelapse</title>
	</head>
	<body>
		<h1>Photos</h1>
		<ul><li><a href="/archive/zip">Download .zip archive ({{ .PhotosTotalSize }})<a/></li></ul>
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
		<h1>Preview</h1>
		<p><img src="/capture" alt="Preview" /></p>
		<h1>Administration</h1>
		<ul>
			<li>Configuration
				<table>
					<tr>
						<td>Time between capturing (minutes):</td>
						<td>
							<select id="frequency">
									<option>1</option>
									<option>5</option>
									<option>10</option>
									<option>15</option>
									<option>20</option>
									<option>25</option>
									<option selected="true">30</option>
									<option>35</option>
									<option>40</option>
									<option>45</option>
									<option>50</option>
									<option>55</option>
								</select>
						</td>
					</tr>
					<tr>
						<td>Time offset before capturing first picture (minutes):</td>
						<td>
							<select id="offset">
								<option value="-1">immediate</option>
								<option>0</option>
								<option>5</option>
								<option>10</option>
								<option selected="true">15</option>
								<option>20</option>
								<option>25</option>
								<option>30</option>
								<option>35</option>
								<option>40</option>
								<option>45</option>
								<option>50</option>
								<option>55</option>
							</select>
							<em>immediate</em> starts taking pictures right after turning on. This disables waiting for any offset.
						</td>
					</tr>
					<tr>
						<td>Resolution:</td>
						<td>
							<select id="resolution">
								<option value="0" selected="true">3280x2464 (8 MP)</option>
								<option value="1">2186x1642 (3 MP)</option>
								<option value="2">1640x1232 (2 MP)</option>
							</select>
						</td>
					</tr>
					<tr>
						<td>Rotate by (degrees):</td>
						<td>
							<select id="rotation">
								<option>0</option>
								<option>180</option>
							</select>
						</td>
					</tr>
					<tr>
						<td colspan="2">
							<input type="button" id="saveConfigBtn" value="Save" />
							<span id="changesSaved" style="display:none;">Changes applied. Restart at your convenience.</span>
						</td>
					</tr>
				</table>
			</li>
			<li><a href="/admin/clear">Delete existing photos</a></li>
			<li><a href="/admin/shutdown">Shutdown</a></li>
			<li><a href="/admin/restart">Restart</a></li>
			<li>More information: <a href="https://github.com/ingojaeckel/go-raspberry-pi-timelapse">github.com/ingojaeckel/go-raspberry-pi-timelapse</a></li>
		</ul>
	<script>
{{- .JQuerySource -}}
{{- .CustomJS -}}
	</script>
	</body>
</html>`
