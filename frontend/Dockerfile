FROM node:22-alpine AS build
WORKDIR /src
COPY package.json package-lock.json ./
RUN npm ci
COPY . .
# API calls are proxied by nginx in production; the build itself needs no backend.
RUN BACKEND_HOST_PORT=8080 npm run build

FROM nginx:1.27-alpine
RUN addgroup -S app && adduser -S app -G app
COPY --from=build /src/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
HEALTHCHECK --interval=10s --timeout=3s --start-period=5s --retries=5 \
  CMD wget -qO- http://127.0.0.1/ >/dev/null || exit 1
