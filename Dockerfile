FROM golang:alpine AS builder
WORKDIR /app
ADD ./ /app/
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -tags musl ./server/main.go

FROM alpine
COPY --from=builder /app/main /usr/local/bin/server
ENTRYPOINT [ "/usr/local/bin/server" ]
