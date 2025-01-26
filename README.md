nginx-ldap-auth
=====================
Separate microservice to implement [subrequest](https://nginx.org/en/docs/http/ngx_http_auth_request_module.html#auth_request)
[authentication in nginx](https://docs.nginx.com/nginx/admin-guide/security-controls/configuring-subrequest-authentication/)
using openldap as backend.

How it works?
=====================
NGINX acts as reverse proxy, passing requests from client to upstream. But, before allowing request to pass,
NGINX makes sub request to `nginx-ldap-auth` service. It is GET request with headers (and cookies) inherited
from clients request with some extra headers added - namely, the `Host` and `X-Original-URI` with original URI opened
by client on server. Sub request has body omitted. It is worth notice we can extract a lot of authorization information 
from headers, including these two bits: 
- First one is content of `Authorization: Basic ...` header, used in [Basic Authorization Strategy](https://developer.mozilla.org/en-US/docs/Web/HTTP/Authentication).
  Username and password extracted from `Authorization:` header can be used to perform BIND authorization like explained
  in this article - https://vodolaz095.ru/nodejs-openldap/.
- Second one is cookies, being sent by client with every request. We can store encrypted information about users' sessions
  in them including users identification, roles, email addresses and so on.
If `nginx-ldap-auth` service responds with 200th status code on sub-request, nginx will pass clients request to backend,
and if `nginx-ldap-auth` responds with 4xx ones, nginx will restrict access.

So, `nginx-ldap-auth` service runs independently of nginx, allowing it to either pass or restrict clients requests 
to backend while acting as reverse proxy.



Example configuration 1 - basis authorization to limit access to subdirectory
=====================
Use case - we serve static files from directory `/srv/www/site/` using nginx, and we want to implement 
[basic authentication](https://developer.mozilla.org/en-US/docs/Web/HTTP/Authentication) 
in order to grant access to subdirectory `/srv/www/site/private` for members of `designers` group organization unit and to 
few other users: `user1`, `user2`, `user3`. In this example `nginx-ldap-auth` listens on socket in `/var/run/nginx_ldap_auth`.

Nginx config `/etc/nginx/sites/site.example.org.conf`

```

upstream nginx_ldap_auth {
  server unix:/var/run/nginx_ldap_auth/nginx_ldap_auth.sock;
}

server {
    listen       80;
    listen  [::]:80;
    server_name site.example.org;

    proxy_buffer_size   128k;
    proxy_buffers   4 256k;
    proxy_busy_buffers_size   256k;

    # serving public content
    location / {
        root /srv/www/site;
    }

    # serving private content protected by basic authorization
    location /private/ {
        auth_request     /subrequest/basic;
        auth_request_set $auth_status $upstream_status;

        root /srv/www/site/;
    }

    # internal endpoints for authorization subrequests
    location = /subrequest/basic {
        internal;
        proxy_pass              http://nginx_ldap_auth/subrequest/basic;
        proxy_pass_request_body off;
        proxy_set_header        Host $http_host;
        proxy_set_header        Content-Length "";
        proxy_set_header        X-Original-URI $request_uri;
    }
}
```

nginx-ldap-auth configuration - `/etc/ldap_auth.yaml`

```yaml

realm: site.example.org

webserver:
  network: "unix"
  listen: "/var/run/nginx_ldap_auth/nginx_ldap_auth.sock"
  cookie_name: "nginx_ldap_auth"
  session_secret: "long-random-string-333--221"
  session_max_age: "30m"
  subrequest_basic: /subrequest/basic

authenticator:
  ttl: 180s
  connection_string: "ldap://ldap.example.org:389"
  start_tls: false
  insecure_tls: false
  readonly_dn: "cn=readonly,dc=example,dc=org"
  readonly_passwd: "readonly"
  user_base_tpl: "uid=%s,ou=people,dc=example,dc=org"
  groups_ou: "ou=groups,dc=example,dc=org"

tracing:
  protocol: udp
  host: jaeger.example.org
  port: 6831
  ratio: 0.01 # 1% of spans will be recorded


permissions:
  - host: site.example.org # for hostname
    prefix: /private # in order to access all paths under /private, for example /private/image.jpg and so on
    uids: #  user should have uid from this list 
      - user1 # "uid=user1,ou=people,dc=example,dc=org"
      - user2 # "uid=user2,ou=people,dc=example,dc=org"
      - user3 # "uid=user3,ou=people,dc=example,dc=org"
    gids: # or user should be member of this groups
      - designers # "cn=designers,ou=groups,dc=example,dc=org" 

log:
  level: "info"
  to_journald: true

```

Example configuration 2 - cookie session based authorization
==============================================
Use case - consider you have some backend service lacking any authentication mechanisms, and you want to expose it 
to your company members. So, they can open login page in browser, provide credentials and then - access backend.
Some users should be able to access it `user1`, `user2` and`user3` - and same user group - `designers`.
But, in this case `nginx-ldap-auth` is deployed on separate server accessible via http on non standard port 3000.
So, firstly users visits https://backend.example.org/auth to perform authentication using username and password, and then
he/she can access backend on https://backend.example.org/ with cookie basic session. 

Nginx config `/etc/nginx/sites/backend.example.org.conf`

```
# nginx-ldap-auth listens http on 3000 port
upstream nginx_ldap_auth {
  server nginx-ldap-auth.example.org:3000;
}

# very important backend expose privately accessible webui on 8080 port
upstream very_important_backend {
  server very_important_backend.example.org:8080;
}

# redirect from http to https
server {
    listen       80;
    listen  [::]:80;
    server_name  backend.example.org;
    location / {
        add_header Cache-Control "private, max-age=10";
        expires 10;
        rewrite ^ https://backend.example.org$request_uri? permanent;
    }
}
# serving reverse-proxied very_important_backend with cookie based authorization from nginx-ldap-auth
server {
    listen       443 ssl;
    listen  [::]:443 ssl;
    server_name backend.example.org.conf;
    keepalive_timeout 60;

    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_prefer_server_ciphers on;
    ssl_certificate     /etc/ssl/nginx/backend.pem;
    ssl_certificate_key /etc/ssl/nginx/backend.pem;


    proxy_buffer_size   128k;
    proxy_buffers   4 256k;
    proxy_busy_buffers_size   256k;

    # serving very important backend with cookie based session provided by nginx-ldap-auth
    location / {
        auth_request     /subrequest/session;
        auth_request_set $auth_status $upstream_status;

        proxy_pass              http://very_important_backend/;
        proxy_set_header        Host $http_host;
        proxy_set_header        X-Real-IP $remote_addr;
        proxy_set_header        X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header        X-Forwarded-Proto https;
        proxy_read_timeout      30;
        proxy_buffer_size       128k;
        proxy_buffers 4         256k;
        proxy_busy_buffers_size 256k;
    }

    # exposing authorization form from nginx-ldap-auth
    location /auth/ {
        proxy_pass              http://nginx_ldap_auth/auth/;
        proxy_set_header        Host $http_host;
        proxy_set_header        X-Real-IP $remote_addr;
        proxy_set_header        X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header        X-Forwarded-Proto https;
        proxy_read_timeout      30;
        proxy_buffer_size       128k;
        proxy_buffers 4         256k;
        proxy_busy_buffers_size 256k;
    }

   # internal endpoints for authorization subrequests 
   location = /subrequest/session {
        internal;
        proxy_pass              http://nginx_ldap_auth/subrequest/session;
        proxy_pass_request_body off;
        proxy_set_header        Host $http_host;
        proxy_set_header        Content-Length "";
        proxy_set_header        X-Original-URI $request_uri;
    }
}

```


nginx-ldap-auth configuration - `/etc/ldap_auth.yaml`

```yaml

realm: site.example.org

webserver:
  network: "tcp"
  listen: "0.0.0.0:3000"
  # it is important to choose cookie name to be unique in order 
  # to not break cookie based sessions of backends
  cookie_name: "nginx_ldap_auth"
  session_secret: "long-random-string-333--221"
  session_max_age: "30m"
  profile_prefix: /auth
  subrequest_session: /subrequest/session


authenticator:
  ttl: 180s
  connection_string: "ldap://ldap.example.org:389"
  start_tls: false
  insecure_tls: false
  readonly_dn: "cn=readonly,dc=example,dc=org"
  readonly_passwd: "readonly"
  user_base_tpl: "uid=%s,ou=people,dc=example,dc=org"
  groups_ou: "ou=groups,dc=example,dc=org"

tracing:
  protocol: udp
  host: jaeger.example.org
  port: 6831
  ratio: 0.01 # 1% of spans will be recorded

permissions:
  - host: site.example.org # for hostname
    prefix: /private # in order to access all path under /private, for example /private/image.jpg and so on
    uids: #  user should have uid from this list 
      - user1 # "uid=user1,ou=people,dc=example,dc=org"
      - user2 # "uid=user2,ou=people,dc=example,dc=org"
      - user3 # "uid=user3,ou=people,dc=example,dc=org"
    gids: # or user should be member of this groups
      - designers # "cn=designers,ou=groups,dc=example,dc=org" 


log:
  level: "info"
  to_journald: true

```

Example configuration 3 - mixed case.
=============================================

In this case basic authorization is required for visiting `/basic` directory, cookie based session - for visiting
`/private` directory under same domain `files.example.org`.


Nginx config in `/etc/nginx/sites/files.example.org.conf`
```
upstream nginx_ldap_auth {
  server nginx-ldap-auth.example.org:3000;
}

server {
    listen       80;
    listen  [::]:80;
    server_name files.example.org;

    proxy_buffer_size   128k;
    proxy_buffers   4 256k;
    proxy_busy_buffers_size   256k;

    # serving public content, but direcories of 
    # /usr/share/nginx/html/basic
    # /usr/share/nginx/html/private are not accessible
    location / {
        expires -1;
        add_header Cache-Control "no-cache, no-store, must-revalidate";

        root /usr/share/nginx/html/;
    }

    # serving private content protected by basic authorization
    # session from /usr/share/nginx/html/basic directory
    location /basic/ {
        auth_request     /subrequest/basic;
        auth_request_set $auth_status $upstream_status;

        expires -1;
        add_header Cache-Control "no-cache, no-store, must-revalidate";
        root /usr/share/nginx/html/;
    }

    # serving private content protected by cookie based 
    # session from /usr/share/nginx/html/private directory
    location /private/ {
        auth_request     /subrequest/session;
        auth_request_set $auth_status $upstream_status;

        expires -1;
        add_header Cache-Control "no-cache, no-store, must-revalidate";
        root /usr/share/nginx/html/;
    }

    # exposing authorization form on http://files.example.org/auth/
    location /auth/ {
        proxy_pass              http://nginx_ldap_auth/auth/;
        proxy_set_header        Host $http_host;
        proxy_set_header        X-Real-IP $remote_addr;
        proxy_set_header        X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header        X-Forwarded-Proto https;
        proxy_read_timeout      30;
        proxy_buffer_size       128k;
        proxy_buffers 4         256k;
        proxy_busy_buffers_size 256k;
    }

    # internal endpoints for authorization subrequests
    location = /subrequest/basic {
        internal;
        proxy_pass              http://nginx_ldap_auth/subrequest/basic;
        proxy_pass_request_body off;
        proxy_set_header        Host $http_host;
        proxy_set_header        Content-Length "";
        proxy_set_header        X-Original-URI $request_uri;
    }

   location = /subrequest/session {
        internal;
        proxy_pass              http://nginx_ldap_auth/subrequest/session;
        proxy_pass_request_body off;
        proxy_set_header        Host $http_host;
        proxy_set_header        Content-Length "";
        proxy_set_header        X-Original-URI $request_uri;
    }
}

```

nginx-ldap-auth config

```yaml

realm: oldcity

webserver:
  network: "tcp"
  listen: "0.0.0.0:3000"
  cookie_name: "nginx_ldap_auth"
  session_secret: "secret"
  session_max_age: "30m"
  profile_prefix: /auth
  subrequest_basic: /subrequest/basic
  subrequest_session: /subrequest/session

authenticator:
  ttl: 180s
  connection_string: "ldap://ldap:389"
  start_tls: false
  insecure_tls: false
  readonly_dn: "cn=readonly,dc=example,dc=org"
  readonly_passwd: "readonly"
  user_base_tpl: "uid=%s,ou=people,dc=example,dc=org"
  groups_ou: "ou=groups,dc=example,dc=org"

tracing:
  protocol: udp
  host: jaeger.example.org
  port: 6831
  ratio: 0.01 # 1% of spans will be recorded

permissions:
  - host: files.example.org
    prefix: /basic # in order to access all path under /basic, for example /basic/image.jpg and so on
    uids: #  user should have uid from this list 
      - user1 # "uid=user1,ou=people,dc=example,dc=org"
      - user2 # "uid=user2,ou=people,dc=example,dc=org"
      - user3 # "uid=user3,ou=people,dc=example,dc=org"
    gids: # or user should be member of this groups
      - designers # "cn=designers,ou=groups,dc=example,dc=org" 

  - host: files.example.org
    prefix: /private # in order to access all path under /private, for example /private/image.jpg and so on
    uids: #  user should have uid from this list 
      - user1 # "uid=user1,ou=people,dc=example,dc=org"
      - user2 # "uid=user2,ou=people,dc=example,dc=org"
      - user3 # "uid=user3,ou=people,dc=example,dc=org"
    gids: # or user should be member of this groups
      - designers # "cn=designers,ou=groups,dc=example,dc=org" 

log:
  level: "trace"
  to_journald: false

```

Tracing options
=============================
Tracing options are explained here
https://github.com/vodolaz095/pkg/blob/master/tracing/all.go


Logging options
=============================
Logging options are explained here
https://github.com/vodolaz095/pkg/blob/master/zerologger/config.go
In general, `info` level is ok for production, while `debug` can help fine tune and debug.
