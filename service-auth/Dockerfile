FROM --platform=linux/amd64 harbor.infra.yandex.astral-dev.ru/astral-edo/go/edo-golang-builder:v2.0.6 as builder

WORKDIR /build
COPY . .
RUN --mount=type=cache,id=edicore,target=/go/pkg/mod/ \
    --mount=type=bind,source=go.sum,target=go.sum \
    --mount=type=bind,source=go.mod,target=go.mod \
     go mod download
RUN --mount=type=cache,id=edicore,target=/go/pkg/mod \
    GOOS=linux GOARCH=amd64 go build -v -o ./package-creator cmd/app/main.go

FROM --platform=linux/amd64 harbor.infra.yandex.astral-dev.ru/proxy-hub.docker.com/library/alpine:3.19.1
RUN apk add --no-cache gcompat
WORKDIR /app
COPY --from=builder /build/package-creator ./
COPY --from=builder /build/docs/swagger ./docs/swagger

CMD ["/app/package-creator"]
