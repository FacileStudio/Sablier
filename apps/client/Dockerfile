FROM oven/bun:1-alpine AS build

WORKDIR /app

# Copy lockfile and package.json for caching
COPY package.json bun.lock ./
RUN bun install --frozen-lockfile

COPY . .

# Build-time argument for the API URL
ARG VITE_API_BASE_URL
ENV VITE_API_BASE_URL=$VITE_API_BASE_URL

RUN bun run build

FROM nginx:alpine

COPY --from=build /app/build /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80
