FROM golang:1.17.3-bullseye AS build

COPY go.mod go.sum /app/
WORKDIR /app
RUN go mod download

COPY Makefile favicon.ico *.go /app/
COPY cmd /app/cmd
COPY esa /app/esa
COPY templates /app/templates
COPY utils /app/utils

RUN make

FROM debian:bullseye-slim

RUN apt-get update && \
  apt-get install -y \
  gettext-base \
  ca-certificates

COPY --from=build /app/yktr2 /
COPY dockerfiles /

ENTRYPOINT ["/entrypoint.sh"]
