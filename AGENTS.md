# nginx-ldap-auth - Agent Skills

This project implements a **microservice for LDAP authentication integration with NGINX**, using OpenLDAP as a backend. It enables authentication via two mechanisms:

1. **Basic Authorization** (via HTTP `Authorization: Basic` header)
2. **Cookie Session** (using HTTP cookies with encrypted user sessions)

## Project Overview

The `nginx-ldap-auth` service acts as an intermediary between NGINX (configured with `auth_request` module) and OpenLDAP. When NGINX receives requests for protected paths, it makes an internal subrequest to this service to verify user credentials before allowing access.

### Authentication Strategies

- **Basic Authorization**: Extracts username/password from `Authorization: Basic` headers. The credentials are used to perform LDAP BIND operations, similar to basic auth in other services.
- **Cookie Session**: Stores encrypted user session information (user ID, roles, email) in HTTP cookies. Users authenticate once via a login form, then can access protected resources throughout their session.

### How It Works

1. NGINX receives client request for a protected path
2. NGINX makes an internal subrequest to `nginx-ldap-auth`
3. `nginx-ldap-auth` queries OpenLDAP to verify the user
4. If authenticated, returns 200; otherwise returns 401/403
5. NGINX uses the response to allow or deny the client's request

---

## Available Agent Skills

### `vuln` - Dependency Vulnerability Scanning

**Purpose**: Check Go project dependencies for known vulnerabilities and provide remediation steps.

**Use Cases**:
- Run `go mod graph | vulncheck` to scan dependencies
- Identify vulnerable packages in `go.mod`
- Generate security patches for outdated dependencies
- Review CVE databases (e.g., `nvd-api`, `govulncheck`)

**Typical Commands**:
```bash
# Check for known vulnerabilities
go mod vendor
govulncheck ./...

# Review vulnerable dependencies
go list -m all | grep -E "(golang.org/x|github.com|gopkg.in)"
```

---

### `golint` - Go Style & Syntax Linting

**Purpose**: Enforce Go code style, catch syntax errors, and ensure code quality.

**Use Cases**:
- Run `gofmt`, `staticcheck`, or `golint` to check code style
- Identify linting violations in `main.go`, `config/`, `endpoints/`
- Ensure proper indentation, formatting, and naming conventions
- Validate Go best practices across the codebase

**Typical Commands**:
```bash
# Format and check code
gofmt -s -l .
staticcheck ./...

# Check specific files
staticcheck main.go config/...
```

---

### `go-test` - Go Unit Testing with Coverage

**Purpose**: Run Go unit tests with detailed coverage reports and explanations.

**Use Cases**:
- Run `go test -cover ./...` to test all packages
- Execute `go test -coverprofile=coverage.out ./...` for coverage analysis
- Get detailed test failures with line-by-line explanations
- Validate test coverage for `config/`, `endpoints/`, `middlewares/`

**Typical Commands**:
```bash
# Run tests with coverage
go test -v -cover ./...

# Get coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run specific package tests
go test -v github.com/vodolaz095/nginx-ldap-auth/config
go test -v github.com/vodolaz095/nginx-ldap-auth/endpoints
```

---

## Project Structure

```
nginx-ldap-auth/
├── main.go                      # Entry point, config loading, tracing setup
├── config/
│   ├── config.go                # Configuration structs (WebServer, Authenticator, etc.)
│   ├── load.go                  # Config file loading
│   └── permissions.go           # Permission struct definition
├── endpoints/
│   ├── api.go                   # Main API router, HTTP server startup
│   ├── api_subrequest_basic.go # Basic auth subrequest handler
│   ├── api_subrequest_session.go # Cookie session subrequest handler
│   ├── api_profile.go           # User profile endpoint
│   ├── api_check_permissions.go # Permission checking endpoint
│   ├── injectTemplates.go       # Template injection for views
│   └── md5.go                   # MD5 helper functions
├── middlewares/
│   ├── sessions.go              # Session management middleware
│   ├── emulatePHP.go            # give false positive headers to make attacker to believe there is old and vulnerable PHP environment
│   └── CSRF.go                  # CSRF protection middleware
├── views/
│   ├── layout.tpl               # Main HTML layout
│   ├── login.tpl                # Login form template
│   ├── profile.tpl              # User profile template
│   └── views.go                 # Template file management
├── public/
│   ├── assets/                  # Static assets (CSS, JS)
│   ├── favicon.ico              # Site favicon
│   ├── public.go                # Assets bundling
│   └── script.js                # Frontend JavaScript
│   └── style.css                # Frontend CSS
├── ldif/                        # LDAP configuration files
├── Dockerfile_development       # Development Docker image
├── docker-compose.yaml          # Development environment
├── Makefile                     # Build and test targets
├── go.mod                       # Go module definition
└── README.md                    # Project documentation
```

---

## Key Configuration Fields

### `webserver`
- `network`: Network type (`"unix"` or `"tcp"`)
- `listen`: Listen address (`"/var/run/.../sock"` or `"0.0.0.0:3000"`)
- `cookie_name`: Cookie name for session storage
- `session_secret`: Secret key for session encryption
- `session_max_age`: Session lifetime (e.g., `"30m"`)
- `subrequest_basic`: Internal path for basic auth requests
- `subrequest_session`: Internal path for session auth requests
- `profile_prefix`: Prefix for user profile URLs (e.g., `/auth`)
- `trusted_proxies`: List of trusted IP ranges

### `authenticator`
- `ttl`: Time-to-live for cached user profiles
- `connection_string`: LDAP server address (`ldap://ldap.example.org:389`)
- `start_tls`: Enable STARTTLS for LDAP connection
- `insecure_tls`: Skip TLS certificate verification
- `readonly_dn`: LDAP readonly user DN (e.g., `"cn=readonly,dc=example,dc=org"`)
- `readonly_passwd`: Password for readonly user
- `user_base_tpl`: LDAP base search template (e.g., `"uid=%s,ou=people,dc=example,dc=org"`)
- `groups_ou`: Organization unit for group search (e.g., `"ou=groups,dc=example,dc=org"`)

### `tracing`
- `protocol`: Tracing protocol (`"udp"`)
- `host`: Jaeger/OTLP collector address
- `port`: Collector port
- `ratio`: Sampling ratio for spans (e.g., `0.01` for 1%)

### `permissions`
- `host`: Domain/hostname for permissions
- `prefix`: Path prefix to protect (e.g., `/private`)
- `uids`: List of allowed user UIDs
- `gids`: List of allowed group GIDs

### `log`
- `level`: Logging level (`"info"`, `"debug"`, `"trace"`)
- `to_journald`: Output to systemd journal

---

## Example Usage

### Basic Authorization Example

**Use Case**: Protect `/private/` directory with basic auth using `designers` group members

**NGINX Config**:
```nginx
location /private/ {
    auth_request /subrequest/basic;
    root /srv/www/site/;
}

location = /subrequest/basic {
    internal;
    proxy_pass http://nginx_ldap_auth/subrequest/basic;
    proxy_set_header Host $http_host;
}
```

### Cookie Session Example

**Use Case**: Allow users to authenticate via login form, then access backend with cookie

**NGINX Config**:
```nginx
location /auth/ {
    proxy_pass http://nginx_ldap_auth/auth/;
}

location /subrequest/session {
    internal;
    proxy_pass http://nginx_ldap_auth/subrequest/session;
    proxy_set_header Host $http_host;
}
```

---

## Testing & Validation

### Running Tests
```bash
# Full test suite with coverage
go test -v -cover ./...

# Test specific packages
go test -v github.com/vodolaz095/nginx-ldap-auth/config
go test -v github.com/vodolaz095/nginx-ldap-auth/endpoints
```

### Common Tasks

| Task | Command |
|------|---------|
| Format Go code | `gofmt -s -l .` |
| Check lint issues | `staticcheck ./...` |
| Run tests | `go test -v ./...` |
| Generate coverage | `go test -coverprofile=coverage.out ./...` |
| Build Docker image | `make docker-build` |
| Run locally | `go run main.go config/ldap_auth.yaml` |

---

## Notes

- This is a **Go-based microservice** using Gin framework
- Uses OpenLDAP as authentication backend via `github.com/vodolaz095/ldap4gin`
- Supports OpenTelemetry tracing for distributed tracing
- Integrates with NGINX via `auth_request` subrequest module
- Default session TTL: 30 minutes, user profile cache TTL: 180 seconds
- Supports both Unix socket and TCP network listeners
