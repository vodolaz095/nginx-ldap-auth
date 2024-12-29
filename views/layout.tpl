{{define "header.html"}}
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>{{.title}}</title>
  <meta name="robots" content="index,follow"/>
  <link href="{{ .profilePrefix }}/assets/style.css" rel="stylesheet">
</head>
<body>
<h1>{{.title}}</h1>
<hr/>
{{ range .flashes }}
<p>{{ . }}</h3>
{{ end }}
{{end}}

{{define "footer.html"}}
<p><a href="/">Back</a></p>
<script type="application/javascript" src="{{ .profilePrefix }}/assets/script.js"></script>
</body>
</html>
{{end}}
