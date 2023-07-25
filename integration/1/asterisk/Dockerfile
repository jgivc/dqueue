FROM andrius/asterisk:alpine-18.15.1


RUN apk add --update-cache \
    asterisk-sounds-en \
    asterisk-sounds-moh \
    && rm -rf /var/cache/apk/*

COPY beep.gsm /var/lib/asterisk/sounds/
COPY ./conf/* /etc/asterisk/

