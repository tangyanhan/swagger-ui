FROM golang:1.15 as build

WORKDIR /build
COPY . .
RUN go build ./cmd/swagger-ui

FROM busybox:1.31.1-glibc

WORKDIR /swagger
COPY --from=build /build/swagger-ui .
COPY --from=build /build/ui ./
COPY --from=build /build/index.html ./

ENTRYPOINT [ "/swagger/swagger-ui" ]
