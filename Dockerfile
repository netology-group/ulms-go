FROM golang:1.12-alpine AS build

WORKDIR /go/src/github.com/netology-group/ulms-go

COPY Gopkg.lock Gopkg.toml ./

RUN set -ex \
    && apk add --no-cache --virtual .build alpine-sdk \
    && curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh \
    && dep ensure -v -vendor-only \
    && go get -v github.com/rubenv/sql-migrate/... \
    && apk del .build

COPY . .

RUN set -ex \
    && go build ./cmd/main/... \
    && go build ./cmd/sql-migrate-config/...


FROM swaggerapi/swagger-ui AS swagger

RUN sed --in-place "s#https://petstore.swagger.io/v2/swagger.json#api/v1.yml#g" /usr/share/nginx/html/index.html || true


FROM alpine:latest

RUN set -ex \
    && apk add --no-cache \
        ca-certificates

WORKDIR /etc/ulms-go

VOLUME /etc/ulms-go/configs

EXPOSE 8000

COPY --from=build \
    /go/src/github.com/netology-group/ulms-go/main \
    /go/src/github.com/netology-group/ulms-go/cmd/migrate \
    /go/src/github.com/netology-group/ulms-go/sql-migrate-config \
    /go/bin/sql-migrate \
    /bin/
COPY --from=build /go/src/github.com/netology-group/ulms-go/configs ./configs
COPY --from=build /go/src/github.com/netology-group/ulms-go/migrations ./migrations

COPY --from=swagger /usr/share/nginx/html /swagger-ui
COPY --from=build /go/src/github.com/netology-group/ulms-go/api  /swagger-ui/api

CMD ["main"]
