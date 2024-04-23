FROM alpine/git as cloner

WORKDIR /app

RUN git clone https://github.com/gptscript-ai/gptscript.git .

FROM oven/bun:latest as bun-builder

WORKDIR /app

COPY --from=cloner /app/ui /app
RUN bun install
RUN bun run generate

FROM golang:alpine as go-builder

WORKDIR /app

COPY --from=cloner /app /app
COPY --from=bun-builder /app/.output/public /app/static/ui
RUN CGO_ENABLED=0 go build -ldflags="-s -w" .

FROM alpine:latest

WORKDIR /scripts

COPY --from=cloner /app/examples /scripts/examples
COPY --from=go-builder /app/gptscript /usr/local/bin/gptscript

VOLUME ["/scripts"]

ENTRYPOINT ["gptscript"]

ENV GPTSCRIPT_LISTEN_ADDRESS=0.0.0.0:9090

CMD ["--server"]
