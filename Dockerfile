FROM golang:1.22 as buildgo
WORKDIR /go/src/github.com/shieldoo/shieldoo-mesh-lighthouse/
COPY go.mod .
COPY go.sum .
RUN go get ./...
COPY *.go .
COPY dns/ ./dns/
COPY main/ ./main/
RUN go build -o out/shieldoo-mesh-lighthouse ./main

FROM alpine:latest as final
RUN apk --no-cache add ca-certificates
RUN apk add --no-cache libc6-compat gcompat
RUN apk add --update iptables 
RUN rm -rf /tmp/* /var/tmp/* /var/cache/apk/* /var/cache/distfiles/*
WORKDIR /app/
COPY --from=buildgo /go/src/github.com/shieldoo/shieldoo-mesh-lighthouse/out/ .
COPY start.sh /app/start.sh
RUN chmod 550 /app/start.sh
COPY shieldoo.html /app/shieldoo.html
WORKDIR /app

ENTRYPOINT [ "/bin/sh", "/app/start.sh" ]
