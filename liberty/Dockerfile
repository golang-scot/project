FROM registry.golang.scot/opensuse:latest

ENV APP_ENV prod

COPY liberty /liberty-proxy

CMD ["/liberty-proxy", "serve"]
