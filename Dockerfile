FROM golang:1.21.5-alpine

# Iranian Alpine mirror
RUN sed -i 's|dl-cdn.alpinelinux.org|mirror.arvancloud.ir|g' /etc/apk/repositories
RUN apk update

WORKDIR /app
