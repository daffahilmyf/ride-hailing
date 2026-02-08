#!/usr/bin/env bash
set -euo pipefail

if [ "$#" -lt 2 ]; then
  echo "usage: $0 <subject|uuid> <role> [scopes_csv]"
  echo "env: JWT_SECRET (required), JWT_ISSUER (optional), JWT_AUDIENCE (optional), JWT_TTL_SECONDS (optional)"
  exit 1
fi

SUBJECT="$1"
ROLE="$2"
SCOPES_CSV="${3:-}"

JWT_SECRET="${JWT_SECRET:-}"
if [ -z "$JWT_SECRET" ]; then
  echo "JWT_SECRET is required"
  exit 1
fi

JWT_ISSUER="${JWT_ISSUER:-ride-hailing}"
JWT_AUDIENCE="${JWT_AUDIENCE:-ride-hailing-clients}"
JWT_TTL_SECONDS="${JWT_TTL_SECONDS:-3600}"

if [ "$SUBJECT" = "uuid" ] || [ "$SUBJECT" = "auto" ]; then
  SUBJECT="$(python3 - <<'PY'
import uuid
print(uuid.uuid4())
PY
)"
  echo "generated subject: $SUBJECT" >&2
fi

python3 - <<'PY' "$SUBJECT" "$ROLE" "$SCOPES_CSV" "$JWT_SECRET" "$JWT_ISSUER" "$JWT_AUDIENCE" "$JWT_TTL_SECONDS"
import base64
import hmac
import hashlib
import json
import sys
import time

subject, role, scopes_csv, secret, issuer, audience, ttl = sys.argv[1:]
ttl = int(ttl)
now = int(time.time())

scopes = ""
if scopes_csv:
    scopes = " ".join([s.strip() for s in scopes_csv.split(",") if s.strip()])

header = {"alg": "HS256", "typ": "JWT"}
payload = {
    "sub": subject,
    "role": role,
    "scopes": scopes,
    "iss": issuer,
    "aud": audience,
    "iat": now,
    "exp": now + ttl,
}

def b64url(data: bytes) -> str:
    return base64.urlsafe_b64encode(data).rstrip(b"=").decode("utf-8")

header_b64 = b64url(json.dumps(header, separators=(",", ":"), ensure_ascii=False).encode())
payload_b64 = b64url(json.dumps(payload, separators=(",", ":"), ensure_ascii=False).encode())
signing_input = f"{header_b64}.{payload_b64}".encode()
sig = hmac.new(secret.encode(), signing_input, hashlib.sha256).digest()
token = f"{header_b64}.{payload_b64}.{b64url(sig)}"
print(token)
PY
