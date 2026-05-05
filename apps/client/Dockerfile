# Switch to node:alpine for the build stage as its DNS handling sometimes differs on Dokploy
FROM node:22-alpine AS build

WORKDIR /app

# Conservative npm configuration to prevent socket exhaustion and networking hangs
# - maxsockets=1: Limit to one connection at a time
# - loglevel=verbose: Show exactly where it hangs
# - audit/fund=false: Reduce extra network requests
ENV NPM_CONFIG_MAXSOCKETS=1
ENV NPM_CONFIG_LOGLEVEL=verbose
ENV NPM_CONFIG_AUDIT=false
ENV NPM_CONFIG_FUND=false

# Copy package files
COPY package.json package-lock.json* ./

# Run install with verbose output to identify the exact point of failure
RUN npm install

COPY . .

# Build-time argument for the API URL
ARG VITE_API_BASE_URL
ENV VITE_API_BASE_URL=$VITE_API_BASE_URL

RUN npm run build

FROM nginx:alpine

COPY --from=build /app/build /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80
