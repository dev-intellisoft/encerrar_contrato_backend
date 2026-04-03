#!/usr/bin/env bash
set -euo pipefail

ENV_FILE="${ENV_FILE:-/Users/wellington/Developer/encerrar/backend/.env}"
COMMAND=""

show_help() {
  cat <<'USAGE'
Asaas Customers Utility

Usage:
  asaas_customers.sh <command>

Commands:
  help             Show this help
  fill-ids         Fetch Asaas customers and fill customers.asaas_id by externalReference match
  create-from-db   Create Asaas customers from DB (latest solicitation address), avoid duplicates by externalReference
  backup-delete    Backup all Asaas customers then delete them (with confirmation)

Environment:
  ENV_FILE         Path to env file (default: /Users/wellington/Developer/encerrar/backend/.env)

Examples:
  ./scripts/asaas_customers.sh fill-ids
  ./scripts/asaas_customers.sh create-from-db
  ./scripts/asaas_customers.sh backup-delete
  ENV_FILE=/Users/wellington/Developer/encerrar/backend/.env.development ./scripts/asaas_customers.sh create-from-db
USAGE
}

# No args -> show help
if [[ $# -eq 0 ]]; then
  show_help
  exit 0
fi

for arg in "$@"; do
  case "$arg" in
    help)
      show_help
      exit 0
      ;;
    fill-ids|backup-delete|create-from-db)
      COMMAND="$arg"
      ;;
    -h|--help)
      show_help
      exit 0
      ;;
  esac
done

if [[ -z "$COMMAND" ]]; then
  show_help
  exit 1
fi

if [[ ! -f "$ENV_FILE" ]]; then
  echo "Env file not found: $ENV_FILE" >&2
  exit 1
fi

get_env() {
  local key="$1"
  local line
  line=$(grep -E "^${key}=" "$ENV_FILE" | tail -n 1 || true)
  if [[ -z "$line" ]]; then
    echo ""
    return
  fi
  local val="${line#*=}"
  if [[ "$val" == '"'*'"' ]]; then
    val="${val#\"}"
    val="${val%\"}"
  elif [[ "$val" == "'"*"'" ]]; then
    val="${val#\'}"
    val="${val%\'}"
  fi
  echo "$val"
}

require_jq() {
  if ! command -v jq >/dev/null 2>&1; then
    echo "jq is required but not installed." >&2
    exit 1
  fi
}

ASAAS_URL="$(get_env ASAAS_URL)"
ASAAS_TOKEN="$(get_env ASAAS_TOKEN)"
ASAAS_BASE="${ASAAS_URL%/}/v1"

if [[ -z "$ASAAS_URL" || -z "$ASAAS_TOKEN" ]]; then
  echo "ASAAS_URL or ASAAS_TOKEN missing in $ENV_FILE" >&2
  exit 1
fi

require_jq

fetch_all_customers() {
  local limit=100
  local offset=0
  local has_more=true
  local outfile="$1"

  printf '{"object":"list","data":[]}' > "$outfile"

  while $has_more; do
    echo "Fetching Asaas customers: limit=$limit offset=$offset..."
    local page
    page=$(curl -sS --max-time 30 --get \
      -H "Content-Type: application/json" \
      -H "access_token: $ASAAS_TOKEN" \
      --data-urlencode "limit=$limit" \
      --data-urlencode "offset=$offset" \
      "$ASAAS_BASE/customers")

    local tmp_merge
    tmp_merge=$(mktemp)
    jq -s '{object:"list", data: (.[0].data + .[1].data)}' \
      "$outfile" <(echo "$page") > "$tmp_merge"
    mv "$tmp_merge" "$outfile"

    has_more=$(echo "$page" | jq -r '.hasMore // false')
    offset=$((offset + limit))
  done
}

case "$COMMAND" in
  fill-ids)
    DB_HOST="$(get_env DB_HOST)"
    DB_NAME="$(get_env DB_NAME)"
    DB_USER="$(get_env DB_USER)"
    DB_PASS="$(get_env DB_PASS)"
    DB_PORT="$(get_env DB_PORT)"

    if [[ -z "$DB_HOST" || -z "$DB_NAME" || -z "$DB_USER" || -z "$DB_PASS" || -z "$DB_PORT" ]]; then
      echo "DB_* variables missing in $ENV_FILE" >&2
      exit 1
    fi

    export PGPASSWORD="$DB_PASS"

    readarray -t rows < <(
      psql \
        -h "$DB_HOST" \
        -p "$DB_PORT" \
        -U "$DB_USER" \
        -d "$DB_NAME" \
        -At \
        -F $'\t' \
        -c "select id from customers where (asaas_id is null or asaas_id = '');"
    )

    if [[ ${#rows[@]} -eq 0 ]]; then
      echo "No customers missing asaas_id."
      exit 0
    fi

    echo "Found ${#rows[@]} customers missing asaas_id."

    ids_need=()
    for row in "${rows[@]}"; do
      ids_need+=("$row")
    done

    tmp_all=$(mktemp)
    fetch_all_customers "$tmp_all"

    declare -A match_map
    while IFS=$'\t' read -r ext_ref asaas_id; do
      if [[ -n "$ext_ref" && -n "$asaas_id" ]]; then
        match_map["$ext_ref"]="$asaas_id"
      fi
    done < <(jq -r '.data[] | [.customer.externalReference, .customer.id] | @tsv' "$tmp_all")

    rm -f "$tmp_all"

    updated=0
    for cid in "${ids_need[@]}"; do
      asaas_id="${match_map[$cid]:-}"
      if [[ -z "$asaas_id" ]]; then
        echo "No Asaas match for customer id $cid"
        continue
      fi

      psql \
        -h "$DB_HOST" \
        -p "$DB_PORT" \
        -U "$DB_USER" \
        -d "$DB_NAME" \
        -c "update customers set asaas_id = '$asaas_id' where id = '$cid';"

      echo "Updated customer $cid with asaas_id=$asaas_id"
      updated=$((updated + 1))
      sleep 0.1
    done

    echo "Done. Updated $updated customers."
    ;;

  create-from-db)
    DB_HOST="$(get_env DB_HOST)"
    DB_NAME="$(get_env DB_NAME)"
    DB_USER="$(get_env DB_USER)"
    DB_PASS="$(get_env DB_PASS)"
    DB_PORT="$(get_env DB_PORT)"

    if [[ -z "$DB_HOST" || -z "$DB_NAME" || -z "$DB_USER" || -z "$DB_PASS" || -z "$DB_PORT" ]]; then
      echo "DB_* variables missing in $ENV_FILE" >&2
      exit 1
    fi

    export PGPASSWORD="$DB_PASS"

    # Load existing Asaas customers by externalReference
    tmp_all=$(mktemp)
    fetch_all_customers "$tmp_all"

    declare -A existing_map
    while IFS=$'\t' read -r ext_ref asaas_id; do
      if [[ -n "$ext_ref" && -n "$asaas_id" ]]; then
        existing_map["$ext_ref"]="$asaas_id"
      fi
    done < <(jq -r '.data[] | [.customer.externalReference, .customer.id] | @tsv' "$tmp_all")

    rm -f "$tmp_all"

    # Latest solicitation address per customer
    readarray -t rows < <(
      psql \
        -h "$DB_HOST" \
        -p "$DB_PORT" \
        -U "$DB_USER" \
        -d "$DB_NAME" \
        -At \
        -F $'\t' \
        -c "select distinct on (c.id)
            c.id, c.name, c.cpf, c.email, c.phone,
            a.street, a.number, a.complement, a.neighborhood, a.city, a.state, a.country, a.zip_code
            from solicitations s
            inner join customers c on s.customer_id = c.id
            inner join addresses a on s.address_id = a.id
            order by c.id, s.created_at desc nulls last;"
    )

    if [[ ${#rows[@]} -eq 0 ]]; then
      echo "No customers found from solicitations."
      exit 0
    fi

    created=0
    skipped=0

    for row in "${rows[@]}"; do
      IFS=$'\t' read -r cid name cpf email phone street number complement neighborhood city state country zip_code <<< "$row"

      # Skip if already exists in Asaas by externalReference
      if [[ -n "${existing_map[$cid]:-}" ]]; then
        asaas_id="${existing_map[$cid]}"
        echo "Asaas customer exists for $cid -> $asaas_id. Updating DB."
        psql \
          -h "$DB_HOST" \
          -p "$DB_PORT" \
          -U "$DB_USER" \
          -d "$DB_NAME" \
          -c "update customers set asaas_id = '$asaas_id' where id = '$cid';"
        skipped=$((skipped + 1))
        continue
      fi

      cpf_clean=$(echo "$cpf" | tr -cd '0-9')
      phone_clean=$(echo "$phone" | tr -cd '0-9')
      zip_clean=$(echo "$zip_code" | tr -cd '0-9')

      payload=$(jq -n \
        --arg name "$name" \
        --arg cpf "$cpf_clean" \
        --arg email "$email" \
        --arg phone "$phone_clean" \
        --arg street "$street" \
        --arg number "$number" \
        --arg complement "$complement" \
        --arg neighborhood "$neighborhood" \
        --arg city "$city" \
        --arg state "$state" \
        --arg zip "$zip_clean" \
        --arg country "$country" \
        --arg ext "$cid" \
        '{name:$name, cpfCnpj:$cpf, email:$email, mobilePhone:$phone, address:$street, addressNumber:$number, complement:$complement, province:$neighborhood, city:$city, state:$state, postalCode:$zip, country:$country, externalReference:$ext}')

      response=$(curl -sS --max-time 30 -X POST \
        -H "Content-Type: application/json" \
        -H "access_token: $ASAAS_TOKEN" \
        -d "$payload" \
        "$ASAAS_BASE/customers")

      asaas_id=$(echo "$response" | jq -r '.id // empty')
      if [[ -z "$asaas_id" ]]; then
        echo "Failed to create Asaas customer for $cid ($email)"
        echo "Response: $response"
        continue
      fi

      psql \
        -h "$DB_HOST" \
        -p "$DB_PORT" \
        -U "$DB_USER" \
        -d "$DB_NAME" \
        -c "update customers set asaas_id = '$asaas_id' where id = '$cid';"

      echo "Created Asaas customer for $cid -> $asaas_id"
      created=$((created + 1))
      sleep 0.1
    done

    echo "Done. Created $created customers. Updated existing $skipped customers."
    ;;

  backup-delete)
    backup_root="/Users/wellington/Developer/encerrar/backend/backups"
    run_stamp=$(date +"%Y-%m-%d_%H-%M-%S")
    backup_dir="$backup_root/$run_stamp"
    mkdir -p "$backup_dir"

    all_file="$backup_dir/asaas_customers_all.json"
    fetch_all_customers "$all_file"

    echo "Backup saved to $all_file"

    count_total=$(jq -r '.data | length' "$all_file")
    if [[ "$count_total" == "0" ]]; then
      echo "No customers to delete."
      exit 0
    fi

    echo "About to DELETE $count_total Asaas customers."
    read -r -p "Type DELETE to confirm: " confirm
    if [[ "$confirm" != "DELETE" ]]; then
      echo "Cancelled."
      exit 0
    fi

    ids=$(jq -r '.data[].customer.id // .data[].id // empty' "$all_file")
    count=0

    for id in $ids; do
      echo "Deleting Asaas customer $id..."
      http_code=$(curl -sS -o /dev/null -w "%{http_code}" -X DELETE \
        -H "Content-Type: application/json" \
        -H "access_token: $ASAAS_TOKEN" \
        "$ASAAS_BASE/customers/$id")

      if [[ "$http_code" != "200" && "$http_code" != "204" ]]; then
        echo "  Failed to delete $id (HTTP $http_code)"
        continue
      fi

      count=$((count + 1))
      sleep 0.1
    done

    echo "Done. Deleted $count customers."
    ;;

  *)
    show_help
    exit 1
    ;;

esac
