{{ define "profile.html" }}
{{ template "header.html" . }}
<h3>Wellcome, {{ .user }}</h3>
<p>You are allowed to visit this pathes on this domain:</p>
<ul>
{{ range .placesAllowed }}
<li><a href="/{{ . }}/">{{ . }}</a></li>
{{ end }}
</ul>
<p><a href="{{ .profilePrefix }}/logout">Logout</a></p>
{{ template "footer.html" . }}
{{ end }}
