FROM registry.golang.scot/opensuse:latest

ENV APP_ENV prod

COPY liberty /liberty-proxy

RUN chmod +x /liberty-proxy

CMD ["/liberty-proxy", "serve"]
