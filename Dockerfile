FROM dock.mau.dev/mautrix/whatsapp

MAINTAINER Fabian Kaindl <github-docker@fabiankaindl.de>
# Add Healthcheck into the Image
HEALTHCHECK --interval=30s --timeout=3s \
  CMD code_http=$(curl -sI -o /dev/null -w %{http_code} http://localhost:80); if [ $code_http != 404 ];then exit 1;else exit 0; fi

# Get latest Security Updates
# --no-cache does produce an error when you try to later delete packages
RUN apk -U upgrade

# Remove not needed packages to make it distroless
# iputils = ping command and co
# apk-tools alpine-tools alpine-keys libc-utils -> remove apk command
# Nobudy should be able to install software inside an image!!!
# bash = We dont need a shell inside a production container
# Curl is needed for healthcheck and is a dependency from the application!
# Bash is a Dependency of the Application Developer but should not be in the production enviroment -> Could be blacklisted from the removal list

RUN apk del iputils apk-tools alpine-keys libc-utils wget bash

# Remove apk-tools entirely and every related files
RUN rm -rf /var/cache/apk /lib/apk /etc/apk
