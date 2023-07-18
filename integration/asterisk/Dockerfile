FROM andrius/asterisk:alpine-18.15.1

COPY log_sender.go /
WORKDIR /
RUN apk add --update-cache \
    asterisk-sounds-en \
    asterisk-sounds-moh \
    go \
    && rm -rf /var/cache/apk/* \
    && go build log_sender.go

COPY beep.gsm /var/lib/asterisk/sounds/
COPY ./conf/* /etc/asterisk/

COPY entrypoint.sh /

ENTRYPOINT [ "" ]
CMD [ "/entrypoint.sh" ]

# FROM andrius/asterisk:alpine-18.15.1

# RUN apk add --update-cache \
#     asterisk-sounds-en \
#     asterisk-sounds-moh \
#     && rm -rf /var/cache/apk/*

# COPY beep.gsm /var/lib/asterisk/sounds/
# COPY ./conf/* /etc/asterisk/
# COPY entrypoint.sh /

# ENTRYPOINT [ "" ]
# CMD [ "/entrypoint.sh" ]
