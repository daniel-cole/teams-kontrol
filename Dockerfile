FROM alpine:3.7

COPY ./bin/teams-kontrol /
RUN chmod +x /teams-kontrol
RUN adduser -D -h / -u 1000 teams-kontrol

ENTRYPOINT /teams-kontrol