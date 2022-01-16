FROM golang:latest AS builder
LABEL maintainer="Remmelt <remmelt@gmail.com>"

WORKDIR /project
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o main main.go


FROM alpine:latest

RUN apk add -Uu --no-cache --purge mpc ca-certificates tzdata \
    && rm -rf /var/cache/apk/* /tmp/*

WORKDIR /project
COPY --from=builder /project/main .

ENTRYPOINT ["./main"]
