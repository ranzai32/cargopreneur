#!/bin/sh
set -eu
dir="${UPLOAD_DIR:-/var/lib/cargopreneur/uploads}"
mkdir -p "$dir"
chown nobody:nobody "$dir"
exec su-exec nobody /api
