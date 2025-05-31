FROM scratch AS build

COPY drone-email-* /usr/bin/drone-email

ENTRYPOINT [ "/usr/bin/drone-email" ]