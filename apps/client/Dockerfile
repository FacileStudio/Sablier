# Use Debian-based Bun image for better stability and glibc support
FROM oven/bun:1 AS build

WORKDIR /app

# Set network concurrency and timeout to prevent ConnectionRefused/Hangs in Docker
ENV BUN_CONFIG_NETWORK_CONCURRENCY=1
ENV BUN_CONFIG_NETWORK_TIMEOUT=300000

# Copy lockfile and package.json for caching
COPY package.json bun.lock ./

# Clear cache and install with explicit registry to ensure connectivity
RUN bun install --frozen-lockfile --registry=https://registry.npmjs.org

COPY . .

# Build-time argument for the API URL
ARG VITE_API_BASE_URL
ENV VITE_API_BASE_URL=$VITE_API_BASE_URL

RUN bun run build

FROM nginx:alpine

COPY --from=build /app/build /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80
