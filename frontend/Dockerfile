# Build stage
# The output is static files, so the build always runs on the build platform:
# a multi-arch image only re-wraps the same dist/ in an nginx layer per arch.
FROM --platform=$BUILDPLATFORM node:20-alpine AS builder

WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci

COPY . .
RUN npm run build

# Runtime stage
FROM nginx:alpine

ARG VERSION=dev
ARG COMMIT=unknown

LABEL org.opencontainers.image.title="reqmango-frontend" \
      org.opencontainers.image.description="Reqmango web UI" \
      org.opencontainers.image.source="https://github.com/vinthuy/reqmango" \
      org.opencontainers.image.url="https://github.com/vinthuy/reqmango" \
      org.opencontainers.image.documentation="https://github.com/vinthuy/reqmango/blob/master/README.md" \
      org.opencontainers.image.licenses="MIT" \
      org.opencontainers.image.version="${VERSION}" \
      org.opencontainers.image.revision="${COMMIT}"

COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=builder /app/dist /usr/share/nginx/html

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
