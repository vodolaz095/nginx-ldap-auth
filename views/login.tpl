{{ define "login.html" }}
{{ template "header.html" . }}
<h1>Authorization required</h1>
{{ range .flashes }}
<h3>{{ . }}</h3>
{{ end }}
<form action="{{ .profilePrefix }}/login" method="post">
  <input name="_csrf" type="hidden" value="{{ .csrf }}"/>
  <label for="username">Username:</label>
  <input id="username" name="username" value="" type="text" placeholder="Svetlana"/>
  <label for="password">Password:</label>
  <input id="password" name="password" value="" type="password" placeholder="secret"/>
  <input type="submit" value="Login">
  <input type="reset" value="Reset">
</form>
{{ template "footer.html" . }}
{{ end }}
