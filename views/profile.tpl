{{ define "profile.html" }}
{{ template "header.html" . }}
<h1>Profile</h1>
{{ range .flashes }}
<h3>{{ . }}</h3>
{{ end }}
<p><a href="/auth/logout">Logout</a></p>
{{ template "footer.html" . }}
{{ end }}
