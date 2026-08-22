FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /out/server ./cmd/server

FROM alpine:3.22
RUN addgroup -S app && adduser -S app -G app \
    && apk add --no-cache wget ca-certificates
USER app
COPY --from=build /out/server /usr/local/bin/server
EXPOSE 8080
HEALTHCHECK --interval=10s --timeout=3s --start-period=5s --retries=5 \
  CMD wget -qO- http://127.0.0.1:8080/api/v1/healthz >/dev/null || exit 1
ENTRYPOINT ["server"]
