{{ define "profile.html" }}
{{ template "header.html" . }}
<h1>Profile</h1>
{{ range .flashes }}
<h3>{{ . }}</h3>
{{ end }}
<h1>Wellcome, {{ .user }}</h1>
<p>Your are allowed to visit this pathes on this domain:</p>
<ul>
{{ range .placesAllowed }}
<li><a href="/{{ . }}/">{{ . }}</a></li>
{{ end }}
</ul>
<p><a href="{{ .profilePrefix }}/logout">Logout</a></p>
{{ template "footer.html" . }}
{{ end }}
