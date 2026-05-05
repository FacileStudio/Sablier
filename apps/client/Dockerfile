# Use Node.js for building as its network stack is more stable in restricted environments like Dokploy
FROM node:22-slim AS build

WORKDIR /app

# Copy package.json and package-lock.json if it exists
COPY package.json package-lock.json* ./

# npm is much more tolerant of network issues/Docker BuildKit edge cases than Bun's Zig-based stack
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
