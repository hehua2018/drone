FROM alpine:latest

COPY drone /

WORKDIR /

ENTRYPOINT ["./drone"]
