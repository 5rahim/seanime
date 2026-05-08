FROM node:24-alpine AS frontend
WORKDIR /build
COPY seanime-web/package*.json ./
RUN npm ci
COPY seanime-web/ .
RUN npm run build

FROM golang:1.26-alpine AS builder
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=frontend /build/out ./web/
RUN CGO_ENABLED=0 go build -o seanime -trimpath -ldflags="-s -w"

FROM alpine:3.23
RUN apk add --no-cache ca-certificates ffmpeg
WORKDIR /app
COPY --from=builder /build/seanime .
EXPOSE 43211
VOLUME /data
ENTRYPOINT ["/app/seanime"]
CMD ["--datadir", "/data", "--host", "0.0.0.0", "--port", "43211"]
