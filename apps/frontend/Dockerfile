FROM node:22-alpine AS base

WORKDIR /app

RUN corepack enable && corepack prepare pnpm@latest --activate

# Install dependencies (layer cached separately from source)
COPY package.json pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile

# Copy source (overridden at runtime by volume mount in dev)
COPY . .

# Test stage — run from CI with: docker build --target test .
FROM base AS test
RUN pnpm test

# Runtime stage
FROM base AS runtime
COPY docker-entrypoint.sh /docker-entrypoint.sh
RUN chmod +x /docker-entrypoint.sh

EXPOSE 80

ENTRYPOINT ["/docker-entrypoint.sh"]
