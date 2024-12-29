{{ define "login.html" }}
{{ template "header.html" . }}
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
