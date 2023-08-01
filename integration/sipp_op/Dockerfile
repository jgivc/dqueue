FROM alpine:3.18.2
RUN apk add --update-cache sipp

COPY operator_accounts.csv register.xml uas.xml entrypoint.sh /sipp/
WORKDIR /sipp

ENTRYPOINT ["./entrypoint.sh"]
