ARG BUILD_FROM=ghcr.io/home-assistant/amd64-base:3.19

FROM node:20-alpine AS frontend-builder
WORKDIR /frontend
COPY frontend/package.json frontend/package-lock.json* ./
RUN if [ -f package-lock.json ]; then npm ci; else npm install; fi
COPY frontend/ ./
RUN npm run build

FROM golang:1.22-alpine AS builder
WORKDIR /src
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 go build -o /out/mikrotik-presence ./cmd/server

FROM ${BUILD_FROM}
SHELL ["/bin/ash", "-o", "pipefail", "-c"]

RUN apk add --no-cache ca-certificates
COPY --from=builder /out/mikrotik-presence /usr/local/bin/mikrotik-presence
COPY --from=frontend-builder /frontend/dist /app/frontend/dist

ENV HTTP_ADDR=:8080
ENV DB_PATH=/data/mikrotik_presence.db
ENV ADDON_OPTIONS_PATH=/data/options.json
ENV FRONTEND_DIST=/app/frontend/dist

CMD ["/usr/local/bin/mikrotik-presence"]
