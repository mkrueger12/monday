FROM docker:20.10 AS docker

FROM golang:1.24-alpine
RUN apk add --no-cache git curl bash

COPY --from=docker /usr/local/bin/docker /usr/local/bin/docker

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o /usr/local/bin/monday .

ENTRYPOINT ["monday"]
CMD ["--help"]
