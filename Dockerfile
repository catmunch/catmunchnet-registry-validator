FROM golang:1.18-alpine as builder
RUN apk update && apk add --no-cache git upx ca-certificates
WORKDIR /src/
COPY . .
RUN go get -d -v
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o validator .
RUN upx --lzma validator

FROM alpine
COPY --from=builder /src/validator /app/validator
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
CMD ["/app/validator"]
ENTRYPOINT "/app/validator"
