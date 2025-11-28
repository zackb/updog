#!/usr/bin/env bash
set -euo pipefail

if [[ -z "${MAXMIND_LICENSE_KEY:-}" ]]; then
  echo "ERROR: MAXMIND_LICENSE_KEY is not set."
  exit 1
fi

OUT_DIR="maxmind-dl"
EDITION="GeoLite2-City"

mkdir -p "$OUT_DIR"

URL="https://download.maxmind.com/app/geoip_download?edition_id=${EDITION}&license_key=${MAXMIND_LICENSE_KEY}&suffix=tar.gz"

echo "Downloading MaxMind ${EDITION}..."
curl -sSL "$URL" -o "${OUT_DIR}/${EDITION}.tar.gz"

echo "Extracting..."
tar -xzf "${OUT_DIR}/${EDITION}.tar.gz" -C "${OUT_DIR}"

MMDB_PATH=$(find "${OUT_DIR}" -name "*.mmdb" | head -n 1)
mv "$MMDB_PATH" "${EDITION}.mmdb"

rm -Rf "$OUT_DIR"
