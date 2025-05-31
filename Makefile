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

# DCMERGE_IMAGE_REGISTRY_NAME
# Defines the name of the new container to be built using several variables.
DCMERGE_IMAGE_REGISTRY_NAME:=git.cryptic.systems
DCMERGE_IMAGE_REGISTRY_USER:=volker.raschek

DCMERGE_IMAGE_NAMESPACE?=${DCMERGE_IMAGE_REGISTRY_USER}
DCMERGE_IMAGE_NAME:=${EXECUTABLE}
DCMERGE_IMAGE_VERSION?=latest
DCMERGE_IMAGE_FULLY_QUALIFIED=${DCMERGE_IMAGE_REGISTRY_NAME}/${DCMERGE_IMAGE_NAMESPACE}/${DCMERGE_IMAGE_NAME}:${DCMERGE_IMAGE_VERSION}

# BIN
# ==============================================================================
dcmerge:
	CGO_ENABLED=0 \
	GOPROXY=$(shell go env GOPROXY) \
		go build -ldflags "-X 'main.version=${VERSION}'" -o ${@} main.go

# CLEAN
# ==============================================================================
PHONY+=clean
clean:
	rm --force --recursive dcmerge

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
install: dcmerge
	install --directory ${DESTDIR}/etc/bash_completion.d
	./dcmerge completion bash > ${DESTDIR}/etc/bash_completion.d/${EXECUTABLE}

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
		--tag ${DCMERGE_IMAGE_FULLY_QUALIFIED} \
		--tag ${DCMERGE_IMAGE_UNQUALIFIED} \
		.

# DELETE CONTAINER IMAGE
# ==============================================================================
PHONY:=container-image/delete
container-image/delete:
	- ${CONTAINER_RUNTIME} image rm ${DCMERGE_IMAGE_FULLY_QUALIFIED}

# PUSH CONTAINER IMAGE
# ==============================================================================
PHONY+=container-image/push
container-image/push:
	echo ${DCMERGE_IMAGE_REGISTRY_PASSWORD} | ${CONTAINER_RUNTIME} login ${DCMERGE_IMAGE_REGISTRY_NAME} --username ${DCMERGE_IMAGE_REGISTRY_USER} --password-stdin
	${CONTAINER_RUNTIME} push ${DCMERGE_IMAGE_FULLY_QUALIFIED}

# PHONY
# ==============================================================================
# Declare the contents of the PHONY variable as phony.  We keep that information
# in a variable so we can use it in if_changed.
.PHONY: ${PHONY}