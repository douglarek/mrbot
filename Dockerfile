FROM alpine:3.4

RUN apk add --no-cache ca-certificates

ADD mrbot /usr/local/bin/

EXPOSE 8080

ENTRYPOINT ["mrbot"]
