VERSION?=$(shell git describe --abbrev=0)+$(shell date +'%Y%m%d_%H%M%S')

EXECUTABLE:=drone-email

DESTDIR?=
PREFIX?=/usr/local

# CONTAINER_RUNTIME
CONTAINER_RUNTIME?=$(shell which podman)

# DRONEEMAIL_IMAGE
DRONEEMAIL_IMAGE_REGISTRY_HOST?=docker.io
DRONEEMAIL_IMAGE_REPOSITORY=volkerraschek/${EXECUTABLE}
DRONEEMAIL_IMAGE_VERSION?=latest
DRONEEMAIL_IMAGE_FULLY_QUALIFIED=${DRONEEMAIL_IMAGE_REGISTRY_HOST}/${DRONEEMAIL_IMAGE_REPOSITORY}:${DRONEEMAIL_IMAGE_VERSION}
DRONEEMAIL_IMAGE_UNQUALIFIED=${DRONEEMAIL_IMAGE_REPOSITORY}:${DRONEEMAIL_IMAGE_VERSION}

# BINARIES
# ==============================================================================
EXECUTABLES := ${EXECUTABLE}
EXECUTABLES += $(addsuffix .sh, ${EXECUTABLE})
EXECUTABLES += $(addsuffix .fish, ${EXECUTABLE})
EXECUTABLES += $(addsuffix .zsh, ${EXECUTABLE})

all: ${EXECUTABLES}

${EXECUTABLE}:
	CGO_ENABLED=0 \
	GONOPROXY=$(shell go env GONOPROXY) \
	GONOSUMDB=$(shell go env GONOSUMDB) \
	GOPRIVATE=$(shell go env GOPRIVATE) \
	GOPROXY=$(shell go env GOPROXY) \
		go build -ldflags "-X main.version=${VERSION:v%=%}" -o ${@}

${EXECUTABLE}.sh: ${EXECUTABLE}
	./${EXECUTABLE} completion bash > ${EXECUTABLE}.sh

${EXECUTABLE}.fish: ${EXECUTABLE}
	./${EXECUTABLE} completion fish > ${EXECUTABLE}.fish

${EXECUTABLE}.zsh: ${EXECUTABLE}
	./${EXECUTABLE} completion zsh > ${EXECUTABLE}.zsh

# UN/INSTALL
# ==============================================================================
PHONY+=install
install: all
	install --directory ${DESTDIR}${PREFIX}/bin
	install --mode 755 ${EXECUTABLE} ${DESTDIR}${PREFIX}/bin/${EXECUTABLE}

	install --directory ${DESTDIR}/etc/bash_completion.d
	install --mode 644 ${EXECUTABLE}.sh ${DESTDIR}/etc/bash_completion.d/${EXECUTABLE}.sh

	install --directory ${DESTDIR}${PREFIX}/share/fish/vendor_completions.d
	install --mode 644 ${EXECUTABLE}.fish ${DESTDIR}${PREFIX}/share/fish/vendor_completions.d/${EXECUTABLE}.fish

	install --directory ${DESTDIR}${PREFIX}/share/zsh/site-functions
	install --mode 644 ${EXECUTABLE}.zsh ${DESTDIR}${PREFIX}/share/zsh/site-functions/_${EXECUTABLE}.zsh

	install --directory ${DESTDIR}${PREFIX}/licenses/${EXECUTABLE}
	install --mode 644 LICENSE ${DESTDIR}${PREFIX}/licenses/${EXECUTABLE}/LICENSE

PHONY+=uninstall
uninstall:
	-rm --recursive --force \
		${DESTDIR}${PREFIX}/bin/${EXECUTABLE} \
		${DESTDIR}/etc/bash_completion.d/${EXECUTABLE}.sh \
		${DESTDIR}${PREFIX}/share/fish/vendor_completions.d/${EXECUTABLE}.fish \
		${DESTDIR}${PREFIX}/share/zsh/site-functions/_${EXECUTABLE}.zsh \
		${DESTDIR}${PREFIX}/licenses/${EXECUTABLE}/LICENSE

# CLEAN
# ==============================================================================
PHONY+=clean
clean:
	-rm -rf ${EXECUTABLE}*

# TEST
# ==============================================================================
PHONY+=test/unit
test/unit:
	go test -v -race -coverprofile=coverage.txt -covermode=atomic -timeout 600s -count=1 ./...

PHONY+=test/coverage
test/coverage: test/unit
	go tool cover -html=coverage.txt

# GOLANGCI-LINT
# ==============================================================================
PHONY+=golangci-lint
golangci-lint:
	golangci-lint run --concurrency=$(shell nproc)

# CONTAINER-IMAGE
# ==============================================================================
PHONY+=container-image/build
container-image/build:
	${CONTAINER_RUNTIME} build \
		--build-arg GONOPROXY=${GOPROXY} \
		--build-arg GONOSUMDB=${GONOSUMDB} \
		--build-arg GOPRIVATE=${GOPRIVATE} \
		--build-arg GOPROXY=${GOPROXY} \
		--build-arg VERSION=${VERSION} \
		--file ./Dockerfile \
		--no-cache \
		--tag ${DRONEEMAIL_IMAGE_UNQUALIFIED} \
		--tag ${DRONEEMAIL_IMAGE_FULLY_QUALIFIED} \
		.

PHONY+=container-image/push
container-image/push: container-image/build
	${CONTAINER_RUNTIME} push ${DRONEEMAIL_IMAGE_FULLY_QUALIFIED}

# PHONY
# ==============================================================================
# Declare the contents of the PHONY variable as phony.  We keep that information
# in a variable so we can use it in if_changed.
.PHONY: ${PHONY}