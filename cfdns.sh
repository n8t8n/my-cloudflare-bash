#!/bin/bash

# Colores
RED="\033[0;31m"
GREEN="\033[0;32m"
CYAN="\033[0;36m"
YELLOW="\033[1;33m"
RESET="\033[0m"

# Cargar variables desde .env si existe
if [ -f "$(dirname "$0")/.env" ]; then
  source "$(dirname "$0")/.env"
fi

# Validar variables obligatorias
: "${CF_API_TOKEN:?Falta CF_API_TOKEN}"
: "${CF_ZONE_ID:?Falta CF_ZONE_ID}"
: "${CF_API_BASE:=https://api.cloudflare.com/client/v4}"
: "${CF_DOMAIN:?Falta CF_DOMAIN}"

create_record() {
  if [[ "$1" == "--auto-tunnel" ]]; then
    if [[ -z "$CF_DNS_SUB" || -z "$CF_DNS_TARGET" ]]; then
      echo -e "${RED}[!] Missing environment variables for automatic mode.${RESET}"
      exit 1
    fi
    
    SUB="$CF_DNS_SUB"
    TYPE="CNAME"
    TARGET="$CF_DNS_TARGET"
    PROXY_ENABLED=true
    
    echo -e "${CYAN}[*] Automatic mode: Creating CNAME for $SUB.$CF_DOMAIN → $TARGET with proxy enabled${RESET}"
  else
    echo -e "${CYAN}=== Create subdomain in Cloudflare ===${RESET}"
    echo -n "Subdomain (e.g., app): "
    read SUB
    echo -n "Record type (A/CNAME): "
    read TYPE
    TYPE=$(echo "$TYPE" | tr '[:lower:]' '[:upper:]')

    if [[ "$TYPE" == "A" ]]; then
      echo -n "Destination IP (e.g., 192.168.1.10): "
      read TARGET
    elif [[ "$TYPE" == "CNAME" ]]; then
      echo -n "Destination host (e.g., mytunnel.cfargotunnel.com): "
      read TARGET
    else
      echo -e "${RED}[!] Invalid type. Only A or CNAME.${RESET}"
      exit 1
    fi

    echo -n "Enable Cloudflare proxy? (y/n): "
    read PROXY_OPTION
    PROXY_ENABLED=false
    PROXY_OPTION_LOWER=$(echo "$PROXY_OPTION" | tr '[:upper:]' '[:lower:]')
    if [[ "$PROXY_OPTION_LOWER" == "y" || "$PROXY_OPTION_LOWER" == "yes" || "$PROXY_OPTION_LOWER" == "si" || "$PROXY_OPTION_LOWER" == "sí" ]]; then
      PROXY_ENABLED=true
      echo -e "${CYAN}[*] Cloudflare proxy will be enabled (orange cloud)${RESET}"
    else
      echo -e "${CYAN}[*] Cloudflare proxy will not be enabled (gray cloud)${RESET}"
    fi
  fi

  FULL_NAME="$SUB.$CF_DOMAIN"

  echo -e "${CYAN}[*] Checking if record already exists...${RESET}"
  CHECK_RESPONSE=$(curl -s -X GET "$CF_API_BASE/zones/$CF_ZONE_ID/dns_records?name=$FULL_NAME" \
    -H "Authorization: Bearer $CF_API_TOKEN" \
    -H "Content-Type: application/json")
  
  if echo "$CHECK_RESPONSE" | grep -q "\"name\":\"$FULL_NAME\""; then
    echo -e "${YELLOW}[!] A record already exists for $FULL_NAME${RESET}"
    exit 0
  fi

  echo -e "${CYAN}[*] Creating $TYPE record for $FULL_NAME → $TARGET ...${RESET}"

  RESPONSE=$(curl -s -X POST "$CF_API_BASE/zones/$CF_ZONE_ID/dns_records" \
    -H "Authorization: Bearer $CF_API_TOKEN" \
    -H "Content-Type: application/json" \
    --data "{
      \"type\": \"$TYPE\",
      \"name\": \"$FULL_NAME\",
      \"content\": \"$TARGET\",
      \"ttl\": 1,
      \"proxied\": $PROXY_ENABLED
    }")

  if echo "$RESPONSE" | grep -q '"success":true'; then
    PROXY_STATUS="gray cloud (proxy disabled)"
    if [ "$PROXY_ENABLED" = true ]; then
      PROXY_STATUS="orange cloud (proxy enabled)"
    fi
    echo -e "${GREEN}[✓] Subdomain created successfully: $FULL_NAME → $TARGET${RESET}"
    echo -e "${GREEN}[✓] Proxy status: $PROXY_STATUS${RESET}"
  else
    echo -e "${RED}[✗] Error creating record.${RESET}"
    echo "$RESPONSE"
  fi
}

list_records() {
  echo -e "${CYAN}=== Listado de registros DNS en zona $CF_DOMAIN ===${RESET}"
  
  RESPONSE=$(curl -s -X GET "$CF_API_BASE/zones/$CF_ZONE_ID/dns_records?per_page=100" \
    -H "Authorization: Bearer $CF_API_TOKEN" \
    -H "Content-Type: application/json")

  echo "$RESPONSE" | grep -oP '"id":"\K[^"]+' | while read -r id; do
    name=$(echo "$RESPONSE" | grep -oP '"id":"'"$id"'"[^}]*"name":"\K[^"]+')
    echo -e "ID: $id\tName: $name"
  done
}

delete_record() {
  echo -n "Ingrese el ID del registro DNS que desea eliminar: "
  read RECORD_ID

  if [[ -z "$RECORD_ID" ]]; then
    echo -e "${RED}[!] ID inválido.${RESET}"
    return
  fi

  echo -e "${YELLOW}[-] Eliminando registro DNS ID: $RECORD_ID ...${RESET}"
  RESPONSE=$(curl -s -X DELETE "$CF_API_BASE/zones/$CF_ZONE_ID/dns_records/$RECORD_ID" \
    -H "Authorization: Bearer $CF_API_TOKEN" \
    -H "Content-Type: application/json")

  if echo "$RESPONSE" | grep -q '"success":true'; then
    echo -e "${GREEN}[✓] Registro eliminado correctamente.${RESET}"
  else
    echo -e "${RED}[✗] Error al eliminar registro.${RESET}"
    echo "$RESPONSE"
  fi
}

menu_dns() {
  while true; do
    echo -e "${CYAN}=== Menú DNS ===${RESET}"
    echo "1) Crear subdominio"
    echo "2) Listar registros"
    echo "3) Eliminar registro"
    echo "4) Salir"
    echo -n "Elija opción: "
    read opt

    case $opt in
      1) create_record ;;
      2) list_records ;;
      3) delete_record ;;
      4) exit 0 ;;
      *) echo -e "${RED}[!] Opción inválida.${RESET}" ;;
    esac
  done
}

# Si pasás --auto-tunnel, crea el registro directamente, sino menú
if [[ "$1" == "--auto-tunnel" ]]; then
  create_record "$1"
else
  menu_dns
fi
