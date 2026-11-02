#!/bin/sh
set -e

if [ "$ENV" = "dev" ]; then
    echo "Starting Vite dev server..."
    exec pnpm dev --host 0.0.0.0 --port 80
else
    echo "Building frontend for production..."
    pnpm build
    echo "Starting preview server..."
    exec pnpm preview --host 0.0.0.0 --port 80
fi
