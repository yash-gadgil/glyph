#!/usr/bin/env bash
set -euo pipefail

NS="${1:-glyph}"

echo "==> Resetting all data in namespace '$NS'"

echo "--> Dropping userdb schema"
kubectl exec -n "$NS" deploy/userdb -- \
  psql -U user-db-user -d user_db \
  -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;"

echo "--> Dropping orderdb schema"
kubectl exec -n "$NS" deploy/orderdb -- \
  psql -U order-db-user -d order_db \
  -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;"

echo "--> Restarting services (migration init containers re-run)"
kubectl rollout restart -n "$NS" deploy/user deploy/order deploy/order-book

kubectl rollout status -n "$NS" deploy/user --timeout=120s
kubectl rollout status -n "$NS" deploy/order --timeout=120s

echo "==> Done. Fresh schema, every account starts at \$100,000."
