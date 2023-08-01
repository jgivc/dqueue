FROM alpine:3.18.2
RUN apk add --update-cache sipp

COPY client_accounts.csv uac.xml /sipp/
WORKDIR /sipp

ENTRYPOINT ["sipp"]
