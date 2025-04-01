FROM docker.io/library/golang:1.24.2-alpine3.21 AS build

ARG GONOPROXY
ARG GONOSUMDB
ARG GOPRIVATE
ARG GOPROXY
ARG VERSION

COPY . /workspace

WORKDIR /workspace

RUN set -ex && \
    apk update && \
    apk add git make && \
    make install DESTDIR=/drone-email PREFIX=/usr VERSION=${VERSION}

###############################################################################

FROM docker.io/library/alpine:3.21

RUN apk add --no-cache bash bash-completion ca-certificates tzdata

COPY --from=build /drone-email /

ENTRYPOINT [ "/usr/bin/drone-email" ]
