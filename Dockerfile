FROM dock.mau.dev/mautrix/whatsapp

MAINTAINER Fabian Kaindl <github-docker@fabiankaindl.de>
# Add Healthcheck into the Image
HEALTHCHECK --interval=30s --timeout=3s \
  CMD code_http=$(curl -sI -o /dev/null -w %{http_code} http://localhost:80); if [ $code_http != 404 ];then exit 1;else exit 0; fi

# Get latest Security Updates
RUN apk -U upgrade

# Remove not needed packages to make it distroless

RUN apk del iputils
