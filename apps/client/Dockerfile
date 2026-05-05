# Use full Node.js Debian image for the most robust networking stack
FROM node:22 AS build

WORKDIR /app

# Copy package files
COPY package.json package-lock.json* ./

# npm configuration
ENV NPM_CONFIG_MAXSOCKETS=1
ENV NPM_CONFIG_AUDIT=false
ENV NPM_CONFIG_FUND=false

# FIX: Overwrite the broken Tailscale DNS with a public one for the install step
# We use a subshell to ensure the DNS is set just before npm runs
RUN echo "nameserver 8.8.8.8" > /etc/resolv.conf && \
    echo "nameserver 1.1.1.1" >> /etc/resolv.conf && \
    npm install

COPY . .

# Build-time argument for the API URL
ARG VITE_API_BASE_URL
ENV VITE_API_BASE_URL=$VITE_API_BASE_URL

RUN npm run build

FROM nginx:alpine

COPY --from=build /app/build /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80
