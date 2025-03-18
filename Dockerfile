# Stage 1 (Build)
FROM golang:1.23.7-alpine AS builder

ARG VERSION
# RUN apk add --update --no-cache git
WORKDIR /app/
COPY go.mod go.sum /app/
RUN go mod download
COPY . /app/
RUN CGO_ENABLED=0 go build \
    -ldflags="-s -w -X github.com/guardlight/server/system.Version=$VERSION" \
    -v \
    -trimpath \
    -o guardlight-server \
    guardlight-server.go
RUN echo "ID=\"distroless\"" > /etc/os-release

# Stage 2 (Final)
FROM gcr.io/distroless/static:latest
COPY --from=builder /etc/os-release /etc/os-release
COPY --from=builder /etc/mime.types /etc/mime.types

COPY --from=builder /app/guardlight-server /usr/bin/

ENTRYPOINT ["/usr/bin/guardlight-server"]
CMD ["--config", "/etc/guardlight/config.yml"]

EXPOSE 6842