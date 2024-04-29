FROM alpine:latest

WORKDIR /app

USER root
COPY seanime /app/seanime
COPY /web /app/web
ENTRYPOINT ["/app/seanime"]
