# Kaindl Network build of mautrix-whatsapp.
# Built from the source in this repository (upstream mautrix/whatsapp + Kaindl call extensions)
# instead of re-tagging the upstream image, so that our patches are actually part of the binary.
FROM golang:1-alpine3.24 AS builder

RUN apk add --no-cache git ca-certificates build-base su-exec olm-dev

COPY . /build
WORKDIR /build
RUN ./build.sh

FROM alpine:3.24

LABEL org.opencontainers.image.description="mautrix-whatsapp container image provided by Kaindl Network with healthcheck, hardening and WhatsApp call bridging"
LABEL org.opencontainers.image.authors="Fabian Kaindl <container@kaindlnetwork.de>"
LABEL org.opencontainers.image.source="https://github.com/kaindlnetwork/mautrix-whatsapp"
LABEL org.opencontainers.image.documentation="https://github.com/kaindlnetwork/mautrix-whatsapp"
LABEL org.opencontainers.image.vendor="Kaindl Network"
LABEL org.opencontainers.image.licenses="AGPL-3.0-or-later"

ENV UID=1337 \
    GID=1337 \
    BRIDGE_PORT=29318

# Get latest security updates and install runtime dependencies.
# bash and yq are needed by docker-run.sh, curl by the healthcheck.
RUN apk upgrade --no-cache && \
    apk add --no-cache ffmpeg su-exec ca-certificates olm bash jq curl yq-go lottieconverter tzdata && \
    # Remove package management so nobody can install software inside the running container.
    apk del --no-cache apk-tools alpine-keys && \
    rm -rf /var/cache/apk /lib/apk /etc/apk /home /srv /media && \
    rm -f /sbin/reboot /sbin/poweroff /sbin/arp /sbin/fdisk /sbin/ifconfig

COPY --from=builder /build/mautrix-whatsapp /usr/bin/mautrix-whatsapp
COPY --from=builder /build/docker-run.sh /docker-run.sh
VOLUME /data

# Maximum retries are 5 according to CIS Docker Benchmark 1.4.0.
# /_matrix/mau/live is served by the appservice HTTP server of every mautrix bridge.
HEALTHCHECK --interval=30s --timeout=3s --retries=5 --start-period=10s \
  CMD curl -fsS "http://localhost:${BRIDGE_PORT}/_matrix/mau/live" || exit 1

EXPOSE 29318

CMD ["/docker-run.sh"]
