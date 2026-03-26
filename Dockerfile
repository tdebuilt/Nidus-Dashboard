# Stage 1: Build Svelte frontend
FROM node:24-alpine AS frontend
WORKDIR /app/web
COPY web/package.json web/package-lock.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

# Stage 2: Build Go binary
FROM golang:1.24-alpine AS backend
WORKDIR /app
RUN go install github.com/swaggo/swag/cmd/swag@latest
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=frontend /app/web/static ./web/static
RUN swag init -g cmd/nidus/main.go -o docs/ --parseDependency --parseInternal
ARG VERSION=dev
RUN CGO_ENABLED=0 go build -ldflags="-s -w -X main.Version=${VERSION}" -o nidus ./cmd/nidus/

# Stage 3: Final image
FROM alpine:3.20
ARG TARGETARCH=amd64
ARG GO2RTC_VERSION=1.9.9
RUN apk add --no-cache ca-certificates su-exec \
    && wget -O /usr/local/bin/go2rtc \
       "https://github.com/AlexxIT/go2rtc/releases/download/v${GO2RTC_VERSION}/go2rtc_linux_${TARGETARCH}" \
    && chmod +x /usr/local/bin/go2rtc
WORKDIR /app
RUN adduser -D -u 1000 nidus
COPY --from=backend /app/nidus .
COPY docker-entrypoint.sh .
RUN mkdir -p /data && chown -R nidus:nidus /app /data && chmod 700 /data
EXPOSE 3777 1984
ENV NIDUS_DB_PATH=/data/nidus.db
ENTRYPOINT ["./docker-entrypoint.sh"]
