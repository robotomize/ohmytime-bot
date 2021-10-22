FROM golang:1.17 AS builder

RUN apt-get -qq update && apt-get -yqq install upx bzip2 tzdata

ENV GO111MODULE=on \
  CGO_ENABLED=0 \
  GOOS=linux \
  GOARCH=amd64

WORKDIR /src

COPY . .

ENV BUILD_INFO_PACKAGE=github.com/ohmytime-bot/internal/buildinfo
ENV BUILD_NAME=ohmytime-bot

RUN go build \
  -trimpath \
  -ldflags "-s -w -X ${BUILD_INFO_PACKAGE}.BuildTag=$(git describe --tags --abbrev=0) -X ${BUILD_INFO_PACKAGE}.Time=$(date -u '+%Y-%m-%d-%H:%M') -X ${BUILD_INFO_PACKAGE}.Name=${BUILD_NAME} -extldflags '-static'" \
  -installsuffix cgo \
  -o /bin/ohmytime-bot \
  ./cmd/ohmytime-bot

RUN cd ./internal/index/assets && tar -xvvf ./cities.idx.tar.gz
RUN strip /bin/ohmytime-bot
RUN upx -q -9 /bin/ohmytime-bot

FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /bin/ohmytime-bot /bin/ohmytime-bot
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /src/internal/index/assets /bin/

VOLUME /data

ENTRYPOINT ["/bin/ohmytime-bot"]
