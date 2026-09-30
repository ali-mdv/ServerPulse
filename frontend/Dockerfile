# syntax=docker/dockerfile:1.6
FROM node:22-alpine AS builder
WORKDIR /app

RUN npm install -g pnpm@10.19.0

COPY package.json pnpm-lock.yaml .npmrc ./
RUN pnpm install --frozen-lockfile

COPY . .

ARG VITE_SERVER_ADDRESS
ARG VITE_API_BASE_URL
ARG VITE_API_TIME_OUT
ARG VITE_ENVIRONMENT
ARG VITE_APP_NAME
ARG VITE_APP_VERSION
ARG VITE_PUBLIC_BUILDER_KEY
ENV VITE_SERVER_ADDRESS=${VITE_SERVER_ADDRESS} \
    VITE_API_BASE_URL=${VITE_API_BASE_URL} \
    VITE_API_TIME_OUT=${VITE_API_TIME_OUT} \
    VITE_ENVIRONMENT=${VITE_ENVIRONMENT} \
    VITE_APP_NAME=${VITE_APP_NAME} \
    VITE_APP_VERSION=${VITE_APP_VERSION} \
    VITE_PUBLIC_BUILDER_KEY=${VITE_PUBLIC_BUILDER_KEY}

RUN pnpm build

FROM nginx:1.27-alpine
COPY --from=builder /app/dist/spa /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
HEALTHCHECK --interval=30s --timeout=3s --retries=3 \
    CMD wget -qO- http://127.0.0.1/ >/dev/null || exit 1

CMD ["nginx", "-g", "daemon off;"]
