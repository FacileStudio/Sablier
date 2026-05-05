# Use full Node.js Debian image (not slim, not alpine) for the most robust networking/DNS stack
FROM node:22 AS build

WORKDIR /app

# Diagnostic: Check DNS configuration and connectivity
RUN cat /etc/resolv.conf && \
    (getent hosts registry.npmjs.org || echo "DNS Resolution Failed")

# Conservative npm configuration
# - maxsockets=1: Limit to one connection to avoid overwhelming weak host DNS/networks
# - fetch-retries=5: Be more patient with transient DNS failures
ENV NPM_CONFIG_MAXSOCKETS=1
ENV NPM_CONFIG_FETCH_RETRIES=5
ENV NPM_CONFIG_FETCH_RETRY_MINTIMEOUT=20000
ENV NPM_CONFIG_FETCH_RETRY_MAXTIMEOUT=120000
ENV NPM_CONFIG_AUDIT=false
ENV NPM_CONFIG_FUND=false

# Copy package files
COPY package.json package-lock.json* ./

# Run install - the full debian image handles DNS much better than alpine in restricted VPCs
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
