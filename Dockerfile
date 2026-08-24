FROM oven/bun:1 AS client-build
WORKDIR /client
COPY apps/client/package.json apps/client/bun.lock* ./
RUN bun install --frozen-lockfile
COPY apps/client/ .
RUN bun run build

# Source maps leave the served tree. The build emits them 'hidden' — no
# sourceMappingURL comment, so no browser asks for one — but a file under
# /client is still reachable by guessing its name, and these carry the original
# sources. Journal reads them from here at boot and resolves stacks server-side.
RUN mkdir -p /sourcemaps \
    && find build -name '*.map' -exec mv {} /sourcemaps/ \; \
    && echo "source maps: $(find /sourcemaps -name '*.map' | wc -l)"

FROM golang:1.26-alpine AS api-build

ARG TARGETOS=linux
ARG TARGETARCH

WORKDIR /repo/apps/api

COPY apps/api/go.mod apps/api/go.sum ./
COPY apps/api/vendor ./vendor

COPY apps/api ./

# The stamp needs history. The .dockerignore re-includes /.git (last rule
# wins); without this copy git describe fails silently inside the assignment
# below and the binary ships an empty version string.
COPY .git /.git

RUN VERSION="$(git --git-dir=/.git describe --tags --always | sed 's/^v//')" \
    && rm -rf /.git

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH:-amd64} \
    go build -mod=vendor -trimpath -ldflags="-s -w -X main.version=$VERSION" -o bin/api .

FROM api-build AS dirs
RUN mkdir -p /data/avatars

FROM gcr.io/distroless/static-debian12

COPY --from=dirs /data /data
COPY --from=api-build /repo/apps/api/bin/api /api
COPY --from=client-build /client/build /client
COPY --from=client-build /sourcemaps /sourcemaps

# A distroless base can carry its own WorkingDir, which would make the relative
# ./client resolve where the SPA is not. Be explicit.
ENV CLIENT_DIR=/client
# Where the API finds the maps to upload. Outside CLIENT_DIR on purpose: nothing
# serves this path.
ENV SOURCEMAP_DIR=/sourcemaps

EXPOSE 4000
VOLUME ["/data"]

ENTRYPOINT ["/api"]
