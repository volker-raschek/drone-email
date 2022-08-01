FROM docker.io/library/golang:1.18.5-alpine3.16 AS build

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

FROM docker.io/library/alpine:3.16

RUN apk add --no-cache bash bash-completion ca-certificates tzdata

COPY --from=build /drone-email /

ENTRYPOINT [ "/usr/bin/drone-email" ]
