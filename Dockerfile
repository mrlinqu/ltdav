FROM golang:1.22-alpine3.20 as bld
RUN apk add make git
WORKDIR /src
COPY ["./","./"]
RUN make build

FROM alpine:3.20 as release
COPY --from=bld /src/bin/ltdav .
RUN mkdir /data
ENV DAV_SERVER_WORKDIR=/data
ENTRYPOINT ["/ltdav"]
