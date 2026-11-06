FROM node:22-alpine AS base

WORKDIR /app

# pnpm comes from "packageManager" in package.json
RUN corepack enable

# Install dependencies (layer cached separately from source)
COPY package.json pnpm-lock.yaml pnpm-workspace.yaml ./
RUN pnpm install --frozen-lockfile

# Copy source (overridden at runtime by volume mount in dev)
COPY . .

# Test stage — run from CI with: docker build --target test .
FROM base AS test
RUN pnpm test

# Runtime stage (dev/local — builds at container start based on $ENV)
FROM base AS runtime
COPY docker-entrypoint.sh /docker-entrypoint.sh
RUN chmod +x /docker-entrypoint.sh

EXPOSE 80

ENTRYPOINT ["/docker-entrypoint.sh"]

FROM base AS build
RUN pnpm build

# Production stage — the built site served by nginx. Nothing is built or installed at start. The
# image's default build target, so it is what `docker build` and CI produce.
#
# SITE_NAME and SITE_DESCRIPTION (optional env) set the page's title and link preview text, see
# docker-site-meta.sh.
FROM nginx:1.29-alpine AS prod
COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=build /app/dist /usr/share/nginx/html
COPY --from=build /app/dist/index.html /usr/share/nginx/index.template.html
COPY --chmod=755 docker-site-meta.sh /docker-entrypoint.d/40-site-meta.sh

EXPOSE 80
