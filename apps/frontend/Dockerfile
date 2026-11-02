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

# Production stage — pre-built for k8s; no runtime build step
FROM base AS prod
RUN pnpm build

EXPOSE 80

CMD ["pnpm", "preview", "--host", "0.0.0.0", "--port", "80"]
