podman/up:
	podman-compose up -d
	podman ps

podman/resource:
	podman-compose up -d jaeger ldap lam nginx
	podman ps

podman/down:
	podman-compose down

podman/prune:
	podman system prune -a --volumes

podman/build:
	podman build -t dashboard:$(majorVersion).$(minorVersion).$(patchVersion) .
