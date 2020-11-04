FROM golang:1.14.3 as build

WORKDIR /build
COPY . .
ENV GOPROXY=https://goproxy.cn
RUN go build -ldflags "-linkmode external -extldflags '-static' -s -w" ./cmd/swagger-ui

FROM busybox:1.31.1-glibc

WORKDIR /swagger
COPY --from=build /build/swagger-ui .
COPY --from=build /build/ui ./ui
COPY --from=build /build/*.html ./
VOLUME [ "/data" ]

ENTRYPOINT [ "/swagger/swagger-ui", "-db", "/data/db.sqlite3" ]
