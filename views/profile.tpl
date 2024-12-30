{{ define "profile.html" }}
{{ template "header.html" . }}
<table>
  <tr>
    <td>
      <h3>General</h3>
      <ul>
        <li>Session expires at: {{.user.ExpiresAt.Format "15:04:05" }}</li>
        <li>DN: {{.user.DN}}</li>
        <li>UID: {{.user.UID}}</li>
      </ul>
    </td>
    <td>
      <h3>Personal</h3>
      <ul>
        {{if .user.GivenName}}<li>Given name: {{.user.GivenName}}</li>{{end}}
        {{if .user.CommonName}}<li>Common name: {{.user.CommonName}}</li>{{end}}
        {{if .user.Initials}}<li>Initials: {{.user.Initials}}</li>{{end}}
        {{if .user.Surname}}<li>Surname: {{.user.Surname}}</li>{{end}}
      </ul>
    </td>
  </tr>
  <tr>
    <td>
      <h3>Organizational</h3>
      <ul>
        {{ if .user.Organization }}<li>Organization: {{.user.Organization}}</li>{{end}}
        {{ if .user.OrganizationUnit }}<li>Unit: {{.user.OrganizationUnit}}</li>{{end}}
        {{ if .user.Title }}<li>Title: {{.user.Title}}</li>{{end}}
        {{ if .user.Description }}<li>Description: {{.user.Description}}</li>{{end}}
      </ul>
    </td>
    <td>
      <h3>Internet related</h3>
      <ul>
        {{ if .user.Website }}<li>Website: <a href="{{.user.Website}}">{{.user.Website}}</a></li>{{end}}
        {{ range $index, $element := .user.Emails }}
        <li>Email {{inc $index}}: <a href="mailto:{{$element}}">{{$element}}</a></li>
        {{ end }}
      </ul>
    </td>
  </tr>
  <tr>
    <td>
      <h3>Linux related</h3>
      <ul>
        {{if .user.UIDNumber}}<li>UID number: {{.user.UIDNumber}}</li>{{end}}
        {{if .user.GIDNumber}}<li>GID number: {{.user.GIDNumber}}</li>{{end}}
        {{if .user.HomeDirectory}}<li>Home directory: {{.user.HomeDirectory}}</li>{{end}}
        {{if .user.LoginShell}}<li>Login shell: {{.user.LoginShell}}</li>{{end}}
      </ul>
    </td>
    <td>
      <h3>Groups of {{.user.CommonName}}:</h3>
      <ul>
        {{ range .user.Groups }}
        <li><b>{{ .GID}} {{.Name}}</b> - {{.Description}}</li>
        {{ end }}
      </ul>
    </td>
  </tr>
</table>
<h3>You are allowed to visit this paths on this domain:</h3>
<ul>
  {{ range .placesAllowed }}
  <li><a href="/{{ . }}/">{{ . }}</a></li>
  {{ end }}
</ul>
<p><a href="{{ .profilePrefix }}/logout">Logout</a></p>
{{ template "footer.html" . }}
{{ end }}
