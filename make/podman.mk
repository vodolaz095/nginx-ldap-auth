podman/up:
	podman-compose up --build --force-recreate --remove-orphans

podman/resource:
	podman-compose up -d jaeger ldap
	podman ps

podman/down:
	podman-compose down

podman/prune:
	podman system prune -a --volumes

podman/build:
	podman build -t dashboard:$(majorVersion).$(minorVersion).$(patchVersion) .
