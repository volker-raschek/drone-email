EXECUTABLE=drone-email
VERSION?=$(shell git describe --abbrev=0)+hash.$(shell git rev-parse --short HEAD)

# Destination directory and prefix to place the compiled binaries, documentaions
# and other files.
DESTDIR?=
PREFIX?=/usr/local

# CONTAINER_RUNTIME
# The CONTAINER_RUNTIME variable will be used to specified the path to a
# container runtime. This is needed to start and run a container image.
CONTAINER_RUNTIME?=$(shell which podman)

# DRONE_EMAIL_IMAGE_REGISTRY_NAME
# Defines the name of the new container to be built using several variables.
DRONE_EMAIL_IMAGE_REGISTRY_NAME:=git.cryptic.systems
DRONE_EMAIL_IMAGE_REGISTRY_USER:=volker.raschek

DRONE_EMAIL_IMAGE_NAMESPACE?=${DRONE_EMAIL_IMAGE_REGISTRY_USER}
DRONE_EMAIL_IMAGE_NAME:=${EXECUTABLE}
DRONE_EMAIL_IMAGE_VERSION?=latest
DRONE_EMAIL_IMAGE_FULLY_QUALIFIED=${DRONE_EMAIL_IMAGE_REGISTRY_NAME}/${DRONE_EMAIL_IMAGE_NAMESPACE}/${DRONE_EMAIL_IMAGE_NAME}:${DRONE_EMAIL_IMAGE_VERSION}

# BIN
# ==============================================================================
drone-email:
	CGO_ENABLED=0 \
	GOPROXY=$(shell go env GOPROXY) \
		go build -ldflags "-X 'main.version=${VERSION}'" -o ${@} main.go

# CLEAN
# ==============================================================================
PHONY+=clean
clean:
	rm --force --recursive drone-email

# TESTS
# ==============================================================================
PHONY+=test/unit
test/unit:
	CGO_ENABLED=0 \
	GOPROXY=$(shell go env GOPROXY) \
		go test -v -p 1 -coverprofile=coverage.txt -covermode=count -timeout 1200s ./pkg/...

PHONY+=test/integration
test/integration:
	CGO_ENABLED=0 \
	GOPROXY=$(shell go env GOPROXY) \
		go test -v -p 1 -count=1 -timeout 1200s ./it/...

PHONY+=test/coverage
test/coverage: test/unit
	CGO_ENABLED=0 \
	GOPROXY=$(shell go env GOPROXY) \
		go tool cover -html=coverage.txt

# GOLANGCI-LINT
# ==============================================================================
PHONY+=golangci-lint
golangci-lint:
	golangci-lint run --concurrency=$(shell nproc)

# INSTALL
# ==============================================================================
PHONY+=uninstall
install: drone-email
	install --directory ${DESTDIR}/etc/bash_completion.d
	./drone-email completion bash > ${DESTDIR}/etc/bash_completion.d/${EXECUTABLE}

	install --directory ${DESTDIR}${PREFIX}/bin
	install --mode 0755 ${EXECUTABLE} ${DESTDIR}${PREFIX}/bin/${EXECUTABLE}

	install --directory ${DESTDIR}${PREFIX}/share/licenses/${EXECUTABLE}
	install --mode 0644 LICENSE ${DESTDIR}${PREFIX}/share/licenses/${EXECUTABLE}/LICENSE

# UNINSTALL
# ==============================================================================
PHONY+=uninstall
uninstall:
	-rm --force --recursive \
		${DESTDIR}/etc/bash_completion.d/${EXECUTABLE} \
		${DESTDIR}${PREFIX}/bin/${EXECUTABLE} \
		${DESTDIR}${PREFIX}/share/licenses/${EXECUTABLE}

# BUILD CONTAINER IMAGE
# ==============================================================================
PHONY+=container-image/build
container-image/build:
	${CONTAINER_RUNTIME} build \
		--build-arg VERSION=${VERSION} \
		--file Dockerfile \
		--no-cache \
		--pull \
		--tag ${DRONE_EMAIL_IMAGE_FULLY_QUALIFIED} \
		--tag ${DRONE_EMAIL_IMAGE_UNQUALIFIED} \
		.

# DELETE CONTAINER IMAGE
# ==============================================================================
PHONY:=container-image/delete
container-image/delete:
	- ${CONTAINER_RUNTIME} image rm ${DRONE_EMAIL_IMAGE_FULLY_QUALIFIED}

# PUSH CONTAINER IMAGE
# ==============================================================================
PHONY+=container-image/push
container-image/push:
	echo ${DRONE_EMAIL_IMAGE_REGISTRY_PASSWORD} | ${CONTAINER_RUNTIME} login ${DRONE_EMAIL_IMAGE_REGISTRY_NAME} --username ${DRONE_EMAIL_IMAGE_REGISTRY_USER} --password-stdin
	${CONTAINER_RUNTIME} push ${DRONE_EMAIL_IMAGE_FULLY_QUALIFIED}

# PHONY
# ==============================================================================
# Declare the contents of the PHONY variable as phony.  We keep that information
# in a variable so we can use it in if_changed.
.PHONY: ${PHONY}