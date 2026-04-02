#!/usr/bin/env bash
set -euo pipefail

ENV_FILE="${ENV_FILE:-/Users/wellington/Developer/encerrar/backend/.env}"

for arg in "$@"; do
  case "$arg" in
    --dev)
      ENV_FILE="/Users/wellington/Developer/encerrar/backend/.env.development"
      ;;
    --prod)
      ENV_FILE="/Users/wellington/Developer/encerrar/backend/.env.production"
      ;;
  esac
done

if [[ ! -f "$ENV_FILE" ]]; then
  echo "Env file not found: $ENV_FILE" >&2
  exit 1
fi

# Read a KEY from env file without eval (handles values with $)
get_env() {
  local key="$1"
  local line
  line=$(grep -E "^${key}=" "$ENV_FILE" | tail -n 1 || true)
  if [[ -z "$line" ]]; then
    echo ""
    return
  fi
  # strip leading KEY=
  local val="${line#*=}"
  # strip surrounding quotes if present
  if [[ "$val" == '"'*'"' ]]; then
    val="${val#\"}"
    val="${val%\"}"
  elif [[ "$val" == "'"*"'" ]]; then
    val="${val#\'}"
    val="${val%\'}"
  fi
  echo "$val"
}

ASAAS_URL="$(get_env ASAAS_URL)"
ASAAS_TOKEN="$(get_env ASAAS_TOKEN)"
DB_HOST="$(get_env DB_HOST)"
DB_NAME="$(get_env DB_NAME)"
DB_USER="$(get_env DB_USER)"
DB_PASS="$(get_env DB_PASS)"
DB_PORT="$(get_env DB_PORT)"

if [[ -z "$ASAAS_URL" || -z "$ASAAS_TOKEN" ]]; then
  echo "ASAAS_URL or ASAAS_TOKEN missing in $ENV_FILE" >&2
  exit 1
fi

if [[ -z "$DB_HOST" || -z "$DB_NAME" || -z "$DB_USER" || -z "$DB_PASS" || -z "$DB_PORT" ]]; then
  echo "DB_* variables missing in $ENV_FILE" >&2
  exit 1
fi

if ! command -v jq >/dev/null 2>&1; then
  echo "jq is required but not installed." >&2
  exit 1
fi

ASAAS_BASE="${ASAAS_URL%/}"

export PGPASSWORD="$DB_PASS"

readarray -t rows < <(
  psql \
    -h "$DB_HOST" \
    -p "$DB_PORT" \
    -U "$DB_USER" \
    -d "$DB_NAME" \
    -At \
    -F $'\t' \
    -c "select id, name, cpf, email, phone, birth_date from customers where (asaas_id is null or asaas_id = '') and email is not null and email <> '';"
)

if [[ ${#rows[@]} -eq 0 ]]; then
  echo "No customers missing asaas_id."
  exit 0
fi

echo "Found ${#rows[@]} customers missing asaas_id."

for row in "${rows[@]}"; do
  id="${row%%$'\t'*}"
  rest="${row#*$'\t'}"
  name="${rest%%$'\t'*}"
  rest="${rest#*$'\t'}"
  cpf="${rest%%$'\t'*}"
  rest="${rest#*$'\t'}"
  email="${rest%%$'\t'*}"
  rest="${rest#*$'\t'}"
  phone="${rest%%$'\t'*}"
  birth_date="${rest#*$'\t'}"

  echo "Looking up Asaas customer for $email ($id)..."

  response=$(curl -sS --get \
    -H "Content-Type: application/json" \
    -H "access_token: $ASAAS_TOKEN" \
    --data-urlencode "email=$email" \
    "$ASAAS_BASE/customers")

  asaas_id=$(echo "$response" | jq -r '.data[0].id // empty')

  if [[ -z "$asaas_id" ]]; then
    echo "  Not found. Creating in Asaas..."

    # Normalize fields
    cpf_clean=$(echo "$cpf" | tr -cd '0-9')
    phone_clean=$(echo "$phone" | tr -cd '0-9')

    create_payload=$(jq -n \
      --arg name "$name" \
      --arg cpf "$cpf_clean" \
      --arg email "$email" \
      --arg phone "$phone_clean" \
      '{name: $name, cpfCnpj: $cpf, email: $email, mobilePhone: $phone}')

    # capture status + body
    tmp_body=$(mktemp)
    http_code=$(curl -sS -o "$tmp_body" -w "%{http_code}" -X POST \
      -H "Content-Type: application/json" \
      -H "access_token: $ASAAS_TOKEN" \
      -d "$create_payload" \
      "$ASAAS_BASE/customers" || echo "000")

    create_response=$(cat "$tmp_body")
    rm -f "$tmp_body"

    if [[ "$http_code" != "200" && "$http_code" != "201" ]]; then
      echo "  Failed to create Asaas customer for $email (HTTP $http_code)"
      echo "  Payload: $create_payload"
      echo "  Response: $create_response"
      continue
    fi

    asaas_id=$(echo "$create_response" | jq -r '.id // empty')

    if [[ -z "$asaas_id" ]]; then
      echo "  Failed to read Asaas id for $email"
      echo "  Response: $create_response"
      continue
    fi

    echo "  Created Asaas customer id=$asaas_id"
  fi

  psql \
    -h "$DB_HOST" \
    -p "$DB_PORT" \
    -U "$DB_USER" \
    -d "$DB_NAME" \
    -v cid="$id" \
    -v asaas="$asaas_id" \
    -c "update customers set asaas_id = :'asaas' where id = :'cid';"

  echo "  Updated $email with asaas_id=$asaas_id"
  sleep 0.2

done

echo "Done."
