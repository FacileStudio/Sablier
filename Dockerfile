FROM oven/bun:1 AS client-build
WORKDIR /client
COPY apps/client/package.json apps/client/bun.lock* ./
RUN bun install --frozen-lockfile
COPY apps/client/ .
RUN bun run build

FROM golang:1.24-alpine AS api-build

ARG TARGETOS=linux
ARG TARGETARCH

WORKDIR /repo/apps/api

COPY apps/api/go.mod apps/api/go.sum ./
COPY apps/api/vendor ./vendor

COPY apps/api ./

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH:-amd64} \
    go build -mod=vendor -trimpath -ldflags="-s -w" -o bin/api .

FROM api-build AS dirs
RUN mkdir -p /data/avatars

FROM gcr.io/distroless/static-debian12

COPY --from=dirs /data /data
COPY --from=api-build /repo/apps/api/bin/api /api
COPY --from=client-build /client/build /client

# A distroless base can carry its own WorkingDir, which would make the relative
# ./client resolve where the SPA is not. Be explicit.
ENV CLIENT_DIR=/client

EXPOSE 4000
VOLUME ["/data"]

ENTRYPOINT ["/api"]
