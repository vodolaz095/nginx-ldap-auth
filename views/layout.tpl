{{define "header.html"}}
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>{{.title}}</title>
  <meta name="robots" content="noindex,nofollow"/>
  <link rel="shortcut icon" type="image/x-icon" href="{{ .profilePrefix }}/assets/favicon.ico"/>
  <link rel="icon" type="image/x-icon" href="{{ .profilePrefix }}/assets/favicon.ico"/>
  <link href="{{ .profilePrefix }}/assets/style.css" rel="stylesheet">
</head>
<body>
<div class="nla-card">
  <h1 class="nla-header">{{.title}}</h1>
  {{ range .flashes }}
  <p class="error">{{ . }}</p>
  {{ end }}
{{end}}

{{define "footer.html"}}
<p><a href="/">Main page</a></p>
<script type="application/javascript" src="{{ .profilePrefix }}/assets/script.js"></script>
</div>
</body>
</html>
{{end}}
