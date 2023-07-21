FROM alpine:3.18.2

RUN apk add --update-cache \
    kamailio \
    kamailio-db \
    kamailio-json \
    kamailio-dbtext \
    kamailio-outbound \
    kamailio-debugger \
    python3 \
    && rm -rf /var/cache/apk/*

ENTRYPOINT [ "kamailio", "-DD",  "-E" ]