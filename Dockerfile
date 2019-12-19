FROM golang:latest AS builder
LABEL maintainer="Remmelt <remmelt@gmail.com>"

WORKDIR /project
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .


FROM alpine:latest

RUN apk --no-cache add ca-certificates
RUN apk add -Uu --no-cache --purge mpc \
    && rm -rf /var/cache/apk/* /tmp/*

WORKDIR /project
COPY --from=builder /project/main .

ENTRYPOINT ["./main"]
