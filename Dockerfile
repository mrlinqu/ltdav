FROM golang:1.24-alpine3.20 as bld
RUN apk add make git
WORKDIR /src
COPY ["./","./"]
RUN make build

FROM alpine:3.20 as release
COPY --from=bld /src/bin/ltdav .
RUN mkdir /data
ENV LTDAV_WORK_DIR=/data
LABEL org.opencontainers.image.source=https://github.com/mrlinqu/ltdav
LABEL org.opencontainers.image.description="ltdav - webdav server image"
LABEL org.opencontainers.image.licenses=MIT
ENTRYPOINT ["/ltdav"]
