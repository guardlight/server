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
    -o guardlight \
    guardlight.go
RUN echo "ID=\"distroless\"" > /etc/os-release

# Stage 2 (Final)
FROM gcr.io/distroless/static:latest

COPY --from=builder /app/guardlight /usr/bin/

ENTRYPOINT ["/usr/bin/guardlight"]

EXPOSE 6842