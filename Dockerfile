# build app
FROM golang:1.18-alpine AS builder

RUN apk update && apk add --no-cache git build-base

WORKDIR /root
COPY . .

RUN CGO_ENABLED=0 go build -o /root/shortener /root/cmd/shortener/ 

# build small image
FROM scratch

COPY --from=builder /root/shortener /shortener

ENTRYPOINT ["/shortener"]
