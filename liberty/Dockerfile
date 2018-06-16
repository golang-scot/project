FROM registry.golang.scot/opensuse:latest

ENV APP_ENV prod

COPY liberty /liberty-proxy

RUN setcap 'cap_net_bind_service=+ep' /liberty-proxy

CMD ["/liberty-proxy", "serve"]
