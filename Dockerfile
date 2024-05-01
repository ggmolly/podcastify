FROM golang:1.22.2-alpine

RUN mkdir -p /app

WORKDIR /app

# install build dependencies for go-sqlite3
RUN apk add --no-cache gcc musl-dev

COPY . /app/

ENV CGO_ENABLED=1

RUN go build -o 'podcastify' main.go

ENTRYPOINT [ "./podcastify" ]