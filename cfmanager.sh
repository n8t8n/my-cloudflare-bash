#!/bin/bash

CONFIG_DIR="$HOME/.cloudflared"
TEMPLATE_CONFIG="$CONFIG_DIR/template.yml"
mkdir -p "$CONFIG_DIR/pids"

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

# Function to generate random pet names
generate_pet_name() {
  adjectives=("happy" "silly" "quick" "lazy" "brave" "calm" "eager" "gentle" "jolly" "kind"
              "morbid" "sarcastic" "cynical" "chaotic" "deranged" "manic" "pessimistic" "nihilistic" "absurd" "grotesque"
              "unhinged" "bizarre" "macabre" "twisted" "sinister" "wicked" "ghoulish" "eerie" "spooky" "ominous")
  nouns=("cat" "dog" "bird" "fish" "tiger" "lion" "bear" "wolf" "fox" "rabbit"
         "ghost" "zombie" "vampire" "skeleton" "demon" "witch" "goblin" "phantom" "wraith" "specter")
  
  # Select 2 random adjectives and 2 random nouns
  name_parts=(
    "${adjectives[$RANDOM % ${#adjectives[@]}]}"
    "${adjectives[$RANDOM % ${#adjectives[@]}]}"
    "${nouns[$RANDOM % ${#nouns[@]}]}"
    "${nouns[$RANDOM % ${#nouns[@]}]}"
  )
  
  # Join with dashes
  IFS='-' eval 'echo "${name_parts[*]}"'
}

crear_tunel() {
  echo -n "Subdomain (e.g., home) [leave blank for random name]: "
  read SUB
  if [[ -z "$SUB" ]]; then
    SUB=$(generate_pet_name)
    echo -e "${COLOR_CYAN}[*] Generated random subdomain: $SUB${COLOR_RESET}"
  fi
  echo -n "Local port (e.g., 3000): "
  read PORT

  echo -e "${COLOR_CYAN}[*] Creating tunnel: $SUB.neptuno.uno → localhost:$PORT ...${COLOR_RESET}"
  
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
  - hostname: "$SUB.neptuno.uno"
    icmp: false
    service: "http://0.0.0.0:$PORT"
  - service: http_status:404
EOF

  echo -e "${COLOR_GREEN}[+] Config generated: $CONFIG_PATH${COLOR_RESET}"
  echo -e "${COLOR_YELLOW}[!] Point DNS to: CNAME → $TUNNEL_ID.cfargotunnel.com${COLOR_RESET}"

  # Preguntar si desea crear el registro DNS automáticamente
  echo -n "Create DNS record automatically? (y/n): "
  read create_dns
  if [[ "$create_dns" == "y" ]]; then
    # Exportar variables para que cfdns.sh pueda usarlas
    export CF_DNS_SUB="$SUB"
    export CF_DNS_TARGET="$TUNNEL_ID.cfargotunnel.com"
    
    # Llamar al script cfdns.sh en modo automático
    SCRIPT_DIR="$(dirname "$0")"
    echo "Calling cfdns.sh from: $SCRIPT_DIR/cfdns.sh"
    ls -l "$SCRIPT_DIR/cfdns.sh"
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
  menu
}


eliminar_tunel() {
  NAME="$1"
  if [[ -z "$NAME" ]]; then
    echo -n "Tunnel name to delete: "
    read NAME
  fi

  CONFIG_PATH="$CONFIG_DIR/$NAME-config.yml"

  if [[ -f "$CONFIG_PATH" ]]; then
    # Extraer ruta del JSON desde el YAML
    CRED_FILE=$(grep 'credentials-file:' "$CONFIG_PATH" | awk '{print $2}' | tr -d '"')
    if [[ -n "$CRED_FILE" && -f "$CRED_FILE" ]]; then
      rm -f "$CRED_FILE"
      echo -e "${COLOR_YELLOW}[-] Credentials file deleted: $CRED_FILE${COLOR_RESET}"
    fi
  fi

  cloudflared tunnel delete "$NAME"
  rm -f "$CONFIG_PATH"
  rm -f "$CONFIG_DIR/pids/$NAME.pid"
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
    if kill -0 "$PID" &>/dev/null; then
      echo -e "${COLOR_YELLOW}[!] $NAME is already running (PID: $PID). Killing it first...${COLOR_RESET}"
      kill "$PID"
      sleep 1
    fi
  fi

  # Start the tunnel fully detached from the terminal
  nohup cloudflared tunnel --config "$CONFIG_PATH" run > /dev/null 2>&1 &
  PID=$!
  echo "$PID" > "$PID_PATH"
  echo -e "${COLOR_GREEN}[+] $NAME started in background (PID: $PID).${COLOR_RESET}"
}

stop_tunel() {
  NAME="$1"
  PID_PATH="$CONFIG_DIR/pids/$NAME.pid"

  if [[ ! -f "$PID_PATH" ]]; then
    echo -e "${COLOR_RED}[!] No PID registered for $NAME.${COLOR_RESET}"
    return
  fi

  PID=$(cat "$PID_PATH")
  if kill -0 "$PID" &>/dev/null; then
    kill "$PID"
    rm -f "$PID_PATH"
    echo -e "${COLOR_YELLOW}[-] $NAME stopped.${COLOR_RESET}"
  else
    echo -e "${COLOR_RED}[!] Process $PID dead, cleaning up.${COLOR_RESET}"
    rm -f "$PID_PATH"
  fi
}

edit_tunel() {
  NAME="$1"
  if [[ -z "$NAME" ]]; then
    echo -n "Tunnel name to edit: "
    read NAME
  fi

  CONFIG_PATH="$CONFIG_DIR/$NAME-config.yml"
  CRED_FILE=$(grep 'credentials-file:' "$CONFIG_PATH" | awk '{print $2}' | tr -d '"')

  if [[ ! -f "$CONFIG_PATH" ]]; then
    echo -e "${COLOR_RED}[!] Tunnel not found: $NAME${COLOR_RESET}"
    return
  fi

  echo -e "${COLOR_CYAN}[*] Editing configuration files for tunnel: $NAME${COLOR_RESET}"
  echo -e "${COLOR_YELLOW}[!] Press Enter to open the files in your default editor.${COLOR_RESET}"

  # Open the YAML config file
  echo -e "${COLOR_BLUE}[*] Editing YAML config: $CONFIG_PATH${COLOR_RESET}"
  read -p "Press Enter to continue..."
  ${EDITOR:-nano} "$CONFIG_PATH"

  # Open the JSON credentials file
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
    printf "%-20s %-10s %-25s\n" "Tunnel" "Status" "Config File"
    print_line

    TUNNELS=($(find "$CONFIG_DIR" -name '*-config.yml'))
    if [[ ${#TUNNELS[@]} -eq 0 ]]; then
      echo -e "${COLOR_RED}No tunnels configured.${COLOR_RESET}"
    else
      for config in "${TUNNELS[@]}"; do
        NAME=$(basename "$config" | sed 's/-config.yml//')
        PID_FILE="$CONFIG_DIR/pids/$NAME.pid"

        if [[ -f "$PID_FILE" ]]; then
          PID=$(cat "$PID_FILE")
          if kill -0 "$PID" &>/dev/null; then
            STATUS="${COLOR_GREEN}RUNNING${COLOR_RESET}"
          else
            STATUS="${COLOR_RED}DEAD${COLOR_RESET}"
            rm -f "$PID_FILE"
          fi
        else
          STATUS="${COLOR_YELLOW}STOPPED${COLOR_RESET}"
        fi

        printf "%-20s %-10b %-25s\n" "$NAME" "$STATUS" "$config"
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
