FROM node:22-alpine

WORKDIR /app

COPY build ./build
COPY package.json ./package.json

EXPOSE 3000

ENV HOST=0.0.0.0
ENV PORT=3000
ENV NODE_ENV=production

CMD ["node", "build"]
