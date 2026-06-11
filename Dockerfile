# syntax=docker/dockerfile:1
#
# Build Pounce with ZERO local toolchain — only Docker required.
#   docker build -t pounce:local .
#   docker run --rm -p 7766:7766 pounce:local
# Then open http://localhost:7766

# --- Stage 1: build the dashboard ---
FROM node:20-alpine AS dashboard
WORKDIR /app/dashboard
COPY dashboard/package.json ./
RUN npm install
COPY dashboard/ ./
RUN npm run build

# --- Stage 2: build the engine (static binary) ---
FROM golang:1.22-alpine AS engine
WORKDIR /app/engine
COPY engine/go.mod ./
RUN go mod download
COPY engine/ ./
RUN CGO_ENABLED=0 go build -ldflags "-s -w" -o /out/pounce ./cmd/pounce

# --- Stage 3: minimal runtime ---
FROM alpine:3.20
RUN apk add --no-cache ca-certificates && adduser -D -u 10001 pounce
COPY --from=engine /out/pounce /usr/local/bin/pounce
COPY --from=dashboard /app/dashboard/dist /opt/pounce/web
USER pounce
WORKDIR /home/pounce
EXPOSE 7766
VOLUME ["/home/pounce/.pounce"]
ENTRYPOINT ["pounce"]
CMD ["--addr", "0.0.0.0:7766", "--static", "/opt/pounce/web"]
