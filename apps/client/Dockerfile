# Use full Node.js Debian image for robust networking
FROM node:22 AS build

WORKDIR /app

# npm configuration
ENV NPM_CONFIG_MAXSOCKETS=1
ENV NPM_CONFIG_AUDIT=false
ENV NPM_CONFIG_FUND=false
ENV NPM_CONFIG_FETCH_RETRIES=5

# Copy package files
COPY package.json package-lock.json* ./

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
