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

crear_tunel() {
  echo -n "Subdominio (ej: home): "
  read SUB
  echo -n "Puerto local (ej: 3000): "
  read PORT

  echo -e "${COLOR_CYAN}[*] Creando túnel: $SUB.neptuno.uno → localhost:$PORT ...${COLOR_RESET}"
  
  TUNNEL_OUTPUT=$(cloudflared tunnel create "$SUB-tunnel" 2>&1)
  TUNNEL_ID=$(echo "$TUNNEL_OUTPUT" | grep -oE '[0-9a-f\-]{36}' | head -n 1)

  if [[ -z "$TUNNEL_ID" ]]; then
    echo -e "${COLOR_RED}[!] No se pudo crear el túnel. Revisión: ${TUNNEL_OUTPUT}${COLOR_RESET}"
    return
  fi

  echo -e "${COLOR_GREEN}[+] Túnel creado con ID: $TUNNEL_ID${COLOR_RESET}"

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

  echo -e "${COLOR_GREEN}[+] Config generado: $CONFIG_PATH${COLOR_RESET}"
  echo -e "${COLOR_YELLOW}[!] Apuntar DNS a: CNAME → $TUNNEL_ID.cfargotunnel.com${COLOR_RESET}"

  # Preguntar si desea crear el registro DNS automáticamente
  echo -n "¿Crear registro DNS automáticamente? (s/n): "
  read create_dns
  if [[ "$create_dns" == "s" ]]; then
    # Exportar variables para que cfdns.sh pueda usarlas
    export CF_DNS_SUB="$SUB"
    export CF_DNS_TARGET="$TUNNEL_ID.cfargotunnel.com"
    
    # Llamar al script cfdns.sh en modo automático
    SCRIPT_DIR="$(dirname "$0")"
    echo "Calling cfdns.sh from: $SCRIPT_DIR/cfdns.sh"
    ls -l "$SCRIPT_DIR/cfdns.sh"
    "$SCRIPT_DIR/cfdns.sh" --auto-tunnel
    
    echo -e "${COLOR_GREEN}[+] Registro DNS creado automáticamente.${COLOR_RESET}"
  fi

  echo -n "¿Querés iniciar el túnel ahora? (s/n): "
  read start
  if [[ "$start" == "s" ]]; then
    cloudflared tunnel --config "$CONFIG_PATH" run &
    echo -e "${COLOR_GREEN}[+] Túnel iniciado en background.${COLOR_RESET}"
  fi

  sleep 1
  menu
}


eliminar_tunel() {
  NAME="$1"
  if [[ -z "$NAME" ]]; then
    echo -n "Nombre del túnel a eliminar: "
    read NAME
  fi

  CONFIG_PATH="$CONFIG_DIR/$NAME-config.yml"

  if [[ -f "$CONFIG_PATH" ]]; then
    # Extraer ruta del JSON desde el YAML
    CRED_FILE=$(grep 'credentials-file:' "$CONFIG_PATH" | awk '{print $2}' | tr -d '"')
    if [[ -n "$CRED_FILE" && -f "$CRED_FILE" ]]; then
      rm -f "$CRED_FILE"
      echo -e "${COLOR_YELLOW}[-] Archivo de credenciales eliminado: $CRED_FILE${COLOR_RESET}"
    fi
  fi

  cloudflared tunnel delete "$NAME"
  rm -f "$CONFIG_PATH"
  rm -f "$CONFIG_DIR/pids/$NAME.pid"
  echo -e "${COLOR_RED}[-] Túnel $NAME eliminado.${COLOR_RESET}"
}


start_tunel() {
  NAME="$1"
  CONFIG_PATH="$CONFIG_DIR/$NAME-config.yml"
  PID_PATH="$CONFIG_DIR/pids/$NAME.pid"

  if [[ ! -f "$CONFIG_PATH" ]]; then
    echo -e "${COLOR_RED}[!] Túnel no encontrado: $NAME${COLOR_RESET}"
    return
  fi

  if [[ -f "$PID_PATH" ]]; then
    PID=$(cat "$PID_PATH")
    if kill -0 "$PID" &>/dev/null; then
      echo -e "${COLOR_YELLOW}[!] $NAME ya está en ejecución (PID: $PID).${COLOR_RESET}"
      return
    fi
  fi

  cloudflared tunnel --config "$CONFIG_PATH" run &> /dev/null &
  PID=$!
  echo "$PID" > "$PID_PATH"
  echo -e "${COLOR_GREEN}[+] $NAME iniciado (PID: $PID).${COLOR_RESET}"
}

stop_tunel() {
  NAME="$1"
  PID_PATH="$CONFIG_DIR/pids/$NAME.pid"

  if [[ ! -f "$PID_PATH" ]]; then
    echo -e "${COLOR_RED}[!] No hay PID registrado para $NAME.${COLOR_RESET}"
    return
  fi

  PID=$(cat "$PID_PATH")
  if kill -0 "$PID" &>/dev/null; then
    kill "$PID"
    rm -f "$PID_PATH"
    echo -e "${COLOR_YELLOW}[-] $NAME detenido.${COLOR_RESET}"
  else
    echo -e "${COLOR_RED}[!] Proceso $PID muerto, limpiando.${COLOR_RESET}"
    rm -f "$PID_PATH"
  fi
}

status_tuneles() {
  while true; do
    clear
    echo -e "${COLOR_CYAN}==== Cloudflared Tunnel Dashboard ====${COLOR_RESET}"
    print_line
    printf "%-20s %-10s %-25s\n" "Túnel" "Estado" "Archivo Config"
    print_line

    TUNNELS=($(find "$CONFIG_DIR" -name '*-config.yml'))
    if [[ ${#TUNNELS[@]} -eq 0 ]]; then
      echo -e "${COLOR_RED}No hay túneles configurados.${COLOR_RESET}"
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
    echo -e "${COLOR_BLUE}Comandos:${COLOR_RESET} start <nombre> | stop <nombre> | delete <nombre>"
    echo -e "          start-all | stop-all | delete-all | refresh | back"
    print_line
    echo -n ">> "
    read action param

    case $action in
      start) start_tunel "$param" ;;
      stop) stop_tunel "$param" ;;
      delete) eliminar_tunel "$param" ;;
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
      *) echo -e "${COLOR_RED}[!] Comando inválido${COLOR_RESET}" ; sleep 1 ;;
    esac
    sleep 1
  done
}

menu() {
  while true; do
    clear
    echo -e "${COLOR_BLUE}Cloudflared Tunnel Manager${COLOR_RESET}"
    print_line
    echo "1. Crear nuevo túnel"
    echo "2. Ver estado de túneles"
    echo "3. Salir"
    print_line
    echo -n "Elige opción: "
    read opt

    case $opt in
      1) crear_tunel ;;
      2) status_tuneles ;;
      3) exit ;;
      *) echo -e "${COLOR_RED}[!] Opción inválida${COLOR_RESET}" ; sleep 1 ;;
    esac
  done
}

menu
