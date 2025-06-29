#!/bin/bash

CONFIG_DIR="$HOME/.cloudflared"
TEMPLATE_CONFIG="$CONFIG_DIR/template.yml"
mkdir -p "$CONFIG_DIR/pids"

# Cargar variables desde .env
if [[ -f .env ]]; then
  export $(grep -v '^#' .env | xargs)
fi

if [[ -z "$CF_DOMAIN" ]]; then
  echo -e "\033[0;31m[!] Variable CF_DOMAIN no está definida en .env\033[0m"
  exit 1
fi

# Colores
COLOR_RED="\033[0;31m"
COLOR_GREEN="\033[0;32m"
COLOR_YELLOW="\033[1;33m"
COLOR_BLUE="\033[0;34m"
COLOR_CYAN="\033[0;36m"
COLOR_RESET="\033[0m"

print_line() {
  printf '%*s\n' "${COLUMNS:-$(tput cols)}" '' | tr ' ' -
}

extract_port() {
  local config_file="$1"
  if [[ -f "$config_file" ]]; then
    grep -oE 'http://[^:]+:([0-9]+)' "$config_file" | grep -oE '[0-9]+$' | head -1
  fi
}

generate_pet_name() {
  adjectives=(
    "morbid" "sarcastic" "chaotic" "deranged" "manic"
    "nihilistic" "twisted" "grotesque" "unhinged" "macabre"
    "sinister" "eerie" "hysterical" "toxic" "bleak"
    "lunatic" "cryptic" "damned" "grim" "volatile"
    "hollow" "rancid" "spiteful" "vile" "decaying"
    "neurotic" "obsessive" "tormented" "malignant" "ashen"
    "dreadful" "morose" "feral" "cursed" "ghastly"
    "seething" "rotting" "gory" "haunted" "infernal"
    "sickly" "chained" "fractured" "stained" "withering"
    "maddened" "coldblooded" "ravaged" "scorched" "crimson"
  )

  nouns=(
    "ghost" "zombie" "vampire" "skeleton" "demon"
    "witch" "phantom" "wraith" "specter" "goblin"
    "banshee" "ghoul" "mutant" "shade" "reaper"
    "fiend" "creep" "entity" "poltergeist" "devil"
    "succubus" "lich" "abomination" "incubus" "jackal"
    "ravager" "howler" "lurker" "serpent" "fallen"
  )

  adj1="${adjectives[$RANDOM % ${#adjectives[@]}]}"
  adj2="${adjectives[$RANDOM % ${#adjectives[@]}]}"
  noun="${nouns[$RANDOM % ${#nouns[@]}]}"

  echo "$adj1-$adj2-$noun"
}


crear_tunel() {
  echo -n "Subdomain (e.g., home) [leave blank for random name or type 'cancel' to exit]: "
  read SUB
  [[ "$SUB" == "exit" || "$SUB" == "cancel" ]] && echo -e "${COLOR_YELLOW}[-] Operación cancelada por el usuario.${COLOR_RESET}" && sleep 1 && return

  if [[ -z "$SUB" ]]; then
    SUB=$(generate_pet_name)
    echo -e "${COLOR_CYAN}[*] Generated random subdomain: $SUB${COLOR_RESET}"
  fi

  echo -n "Local port (e.g., 3000) [or type 'cancel' to abort]: "
  read PORT
  [[ "$PORT" == "exit" || "$PORT" == "cancel" || -z "$PORT" ]] && echo -e "${COLOR_YELLOW}[-] Operación cancelada por el usuario.${COLOR_RESET}" && sleep 1 && return

  echo -e "${COLOR_CYAN}[*] Creating tunnel: $SUB.$CF_DOMAIN → localhost:$PORT ...${COLOR_RESET}"
  
  TUNNEL_OUTPUT=$(cloudflared tunnel create "$SUB-tunnel" 2>&1)
  TUNNEL_ID=$(echo "$TUNNEL_OUTPUT" | grep -oE '[0-9a-f\-]{36}' | head -n 1)

  if [[ -z "$TUNNEL_ID" ]]; then
    echo -e "${COLOR_RED}[!] Could not create tunnel. Review: ${TUNNEL_OUTPUT}${COLOR_RESET}"
    return
  fi

  echo -e "${COLOR_GREEN}[+] Tunnel created with ID: $TUNNEL_ID${COLOR_RESET}"

  CONFIG_PATH="$CONFIG_DIR/$SUB-config.yml"
  CRED_PATH="$CONFIG_DIR/$TUNNEL_ID.json"

  cat > "$CONFIG_PATH" <<EOF
tunnel: "$TUNNEL_ID"
credentials-file: "$CRED_PATH"
ingress:
  - hostname: "$SUB.$CF_DOMAIN"
    icmp: false
    service: "http://0.0.0.0:$PORT"
  - service: http_status:404
EOF

  echo -e "${COLOR_GREEN}[+] Config generated: $CONFIG_PATH${COLOR_RESET}"
  echo -e "${COLOR_YELLOW}[!] Point DNS to: CNAME → $TUNNEL_ID.cfargotunnel.com${COLOR_RESET}"

  echo -n "Create DNS record automatically? (y/n): "
  read create_dns
  if [[ "$create_dns" == "y" ]]; then
    export CF_DNS_SUB="$SUB"
    export CF_DNS_TARGET="$TUNNEL_ID.cfargotunnel.com"
    SCRIPT_DIR="$(dirname "$0")"
    "$SCRIPT_DIR/cfdns.sh" --auto-tunnel
    echo -e "${COLOR_GREEN}[+] DNS record created automatically.${COLOR_RESET}"
  fi

  echo -n "Do you want to start the tunnel now? (y/n): "
  read start
  if [[ "$start" == "y" ]]; then
    cloudflared tunnel --config "$CONFIG_PATH" run &
    echo -e "${COLOR_GREEN}[+] Tunnel started in background.${COLOR_RESET}"
  fi

  sleep 1
}

eliminar_tunel() {
  NAME="$1"
  if [[ -z "$NAME" ]]; then
    echo -n "Tunnel name to delete: "
    read NAME
  fi

  CONFIG_PATH="$CONFIG_DIR/$NAME-config.yml"

  if [[ -f "$CONFIG_PATH" ]]; then
    CRED_FILE=$(grep 'credentials-file:' "$CONFIG_PATH" | awk '{print $2}' | tr -d '"')
    [[ -n "$CRED_FILE" && -f "$CRED_FILE" ]] && rm -f "$CRED_FILE"
  fi

  cloudflared tunnel delete "$NAME"
  rm -f "$CONFIG_PATH" "$CONFIG_DIR/pids/$NAME.pid"
  echo -e "${COLOR_RED}[-] Tunnel $NAME deleted.${COLOR_RESET}"
}

start_tunel() {
  NAME="$1"
  CONFIG_PATH="$CONFIG_DIR/$NAME-config.yml"
  PID_PATH="$CONFIG_DIR/pids/$NAME.pid"

  if [[ ! -f "$CONFIG_PATH" ]]; then
    echo -e "${COLOR_RED}[!] Tunnel not found: $NAME${COLOR_RESET}"
    return
  fi

  if [[ -f "$PID_PATH" ]]; then
    PID=$(cat "$PID_PATH")
    kill -0 "$PID" &>/dev/null && kill "$PID" && sleep 1
  fi

  nohup cloudflared tunnel --config "$CONFIG_PATH" run > /dev/null 2>&1 &
  echo "$!" > "$PID_PATH"
  echo -e "${COLOR_GREEN}[+] $NAME started in background.${COLOR_RESET}"
}

stop_tunel() {
  NAME="$1"
  PID_PATH="$CONFIG_DIR/pids/$NAME.pid"

  if [[ ! -f "$PID_PATH" ]]; then
    echo -e "${COLOR_RED}[!] No PID registered for $NAME.${COLOR_RESET}"
    return
  fi

  PID=$(cat "$PID_PATH")
  kill -0 "$PID" &>/dev/null && kill "$PID"
  rm -f "$PID_PATH"
  echo -e "${COLOR_YELLOW}[-] $NAME stopped.${COLOR_RESET}"
}

edit_tunel() {
  NAME="$1"
  [[ -z "$NAME" ]] && read -p "Tunnel name to edit: " NAME

  CONFIG_PATH="$CONFIG_DIR/$NAME-config.yml"
  CRED_FILE=$(grep 'credentials-file:' "$CONFIG_PATH" | awk '{print $2}' | tr -d '"')

  [[ ! -f "$CONFIG_PATH" ]] && echo -e "${COLOR_RED}[!] Tunnel not found: $NAME${COLOR_RESET}" && return

  echo -e "${COLOR_BLUE}[*] Editing YAML config: $CONFIG_PATH${COLOR_RESET}"
  read -p "Press Enter to continue..."
  ${EDITOR:-nano} "$CONFIG_PATH"

  if [[ -f "$CRED_FILE" ]]; then
    echo -e "${COLOR_BLUE}[*] Editing JSON credentials: $CRED_FILE${COLOR_RESET}"
    read -p "Press Enter to continue..."
    ${EDITOR:-nano} "$CRED_FILE"
  else
    echo -e "${COLOR_RED}[!] Credentials file not found: $CRED_FILE${COLOR_RESET}"
  fi
}

status_tuneles() {
  while true; do
    clear
    echo -e "${COLOR_CYAN}==== Cloudflared Tunnel Dashboard ====${COLOR_RESET}"
    print_line
    printf "%-25s %-10s %-10s %-10s %-10s\n" "Tunnel" "Status" "Port" "CPU%" "MEM%"
    print_line

    TUNNELS=($(find "$CONFIG_DIR" -name '*-config.yml'))
    if [[ ${#TUNNELS[@]} -eq 0 ]]; then
      echo -e "${COLOR_RED}No tunnels configured.${COLOR_RESET}"
    else
      for config in "${TUNNELS[@]}"; do
        NAME=$(basename "$config" | sed 's/-config.yml//')
        PID_FILE="$CONFIG_DIR/pids/$NAME.pid"
        PORT=$(extract_port "$config")
        PORT=${PORT:-N/A}
        CPU="N/A"
        MEM="N/A"

        if [[ -f "$PID_FILE" ]]; then
          PID=$(cat "$PID_FILE")
          if kill -0 "$PID" &>/dev/null; then
            STATUS="${COLOR_GREEN}RUNNING${COLOR_RESET}"
            USAGE=$(ps -p "$PID" -o %cpu,%mem --no-headers)
            CPU=$(echo "$USAGE" | awk '{print $1}')
            MEM=$(echo "$USAGE" | awk '{print $2}')
          else
            STATUS="${COLOR_RED}DEAD${COLOR_RESET}"
            rm -f "$PID_FILE"
          fi
        else
          STATUS="${COLOR_YELLOW}STOPPED${COLOR_RESET}"
        fi

    # Trim nombre a 23 chars para no romper tabla
        NAME_TRIMMED=$(echo "$NAME" | cut -c1-23)
        NAME_PADDED=$(printf "%-25s" "$NAME_TRIMMED")
        STATUS_PLAIN=$(echo -e "$STATUS" | sed 's/\x1B\[[0-9;]*[JKmsu]//g')
        STATUS_PADDED=$(printf "%-10s" "$STATUS_PLAIN")
        PORT_PADDED=$(printf "%-10s" "$PORT")
        CPU_PADDED=$(printf "%-10s" "$CPU")
        MEM_PADDED=$(printf "%-10s" "$MEM")

      printf "%s  %b  %s  %s  %s\n" "$NAME_PADDED" "$STATUS" "$PORT_PADDED" "$CPU_PADDED" "$MEM_PADDED"
      done
    fi

    print_line
    echo -e "${COLOR_BLUE}Commands:${COLOR_RESET} start <name> | stop <name> | delete <name> | edit <name>"
    echo -e "          start-all | stop-all | delete-all | refresh | back"
    print_line
    echo -n ">> "
    read action param

    case $action in
      start) start_tunel "$param" ;;
      stop) stop_tunel "$param" ;;
      delete) eliminar_tunel "$param" ;;
      edit) edit_tunel "$param" ;;
      start-all)
        for conf in "${TUNNELS[@]}"; do
          NAME=$(basename "$conf" | sed 's/-config.yml//')
          start_tunel "$NAME"
        done ;;
      stop-all)
        for conf in "${TUNNELS[@]}"; do
          NAME=$(basename "$conf" | sed 's/-config.yml//')
          stop_tunel "$NAME"
        done ;;
      delete-all)
        for conf in "${TUNNELS[@]}"; do
          NAME=$(basename "$conf" | sed 's/-config.yml//')
          eliminar_tunel "$NAME"
        done ;;
      refresh) continue ;;
      back) break ;;
      *) echo -e "${COLOR_RED}[!] Invalid command${COLOR_RESET}" ; sleep 1 ;;
    esac
    sleep 1
  done
}

menu() {
  while true; do
    clear
    echo -e "${COLOR_BLUE}Cloudflared Tunnel Manager${COLOR_RESET}"
    echo "1. Create new tunnel"
    echo "2. View tunnel status"
    echo "3. Exit"
    print_line
    echo -n "Elige opción: "
    read opt

    case $opt in
      1) crear_tunel ;;
      2) status_tuneles ;;
      3) exit ;;
      *) echo -e "${COLOR_RED}[!] Invalid option${COLOR_RESET}" ; sleep 1 ;;
    esac
  done
}

menu
