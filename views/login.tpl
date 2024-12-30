{{ define "login.html" }}
{{ template "header.html" . }}
<form id="loginForm" class="nla-body" action="{{ .profilePrefix }}/login" method="post">
  <input type="hidden" name="_csrf" value="{{.csrf}}">
  <input type="text" name="username" class="nla-body__email" id="username"
         autocomplete="true" size="128" required placeholder="Username"/>
  <input type="password" name="password" id="password"
         class="nla-body__password" autocomplete="off"
         size="128" required placeholder="Secret"/>
  <input type="submit" value="Login to {{.realm}}">
  <input type="reset" value="Reset">
</form>
{{ template "footer.html" . }}
{{ end }}
