FROM alpine:latest

WORKDIR /app

USER root
COPY seanime /app/seanime
ENTRYPOINT ["/app/seanime"]
