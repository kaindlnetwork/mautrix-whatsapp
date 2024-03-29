FROM dock.mau.dev/mautrix/whatsapp

LABEL org.opencontainers.image.description This is the mautrix-whatapp Container Image provided by Kaindl Network with added Healthcheck and higher security
LABEL org.opencontainers.image.authors Fabian Kaindl container@kaindlnetwork.de
LABEL org.opencontainers.image.source	https://github.com/kgncloud/mautrix-whatsapp/
LABEL org.opencontainers.image.documentation https://github.com/kgncloud/mautrix-whatsapp/
LABEL org.opencontainers.image.vendor Kaindl Network
#LABEL org.opencontainers.image.licenses	
# Add Healthcheck into the Image
# Maximum Retries are 5 times according to  CIS Docker Benchmark 1.4.0 aka Best Practices
HEALTHCHECK --interval=30s --timeout=3s --retries=5 --start-period=10s \
  CMD python3 /healthcheck.py

# We modified the Image to make it compliant with HTTP Standard Port
EXPOSE 80

# Copy our config into the Image
COPY /root /


# Get latest Security Updates
# This is usually not recommended, because it introduces errors to reproduce an image from the source parts but on the other hand does prevent that a part of the supply chain does not get updated fast enough
# --no-cache does produce an error when you try to later delete packages
RUN apk -U upgrade && \

# Remove not needed packages to make it distroless
# iputils = ping command and co
# apk-tools alpine-tools alpine-keys libc-utils -> remove apk command
# Nobudy should be able to install software inside an image!!!
# bash = We dont need a shell inside a production container
# Curl is needed for healthcheck and is a dependency from the application!
# Bash is a Dependency of the Application Developer but should not be in the production enviroment -> Could be blacklisted from the removal list
# ufw will be used as an internal firewall to ensure no unessesary ports can be used by default

apk add python3 ufw=0.36.1-r2 tzdata=2022f-r1 && \

python3 -m ensurepip --upgrade && \

pip3 install requests && \

apk del iputils apk-tools alpine-keys libc-utils wget bash && \
# Remove apk-tools entirely and every related files
rm -rf /var/cache/apk /lib/apk /etc/apk && \
# Remove any folders that are not needed to further shrink down image size and make the image simplified 
rm -rf /home /srv /media /root && \
# Remove commands that should not exist in this image -> This commands depends if the software what build nativly for Containers or if it is just ported
rm /sbin/reboot /sbin/poweroff /sbin/arp /sbin/fdisk /sbin/ifconfig
