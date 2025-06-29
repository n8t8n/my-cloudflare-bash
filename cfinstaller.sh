#!/data/data/com.termux/files/usr/bin/bash

set -e

echo "[+] Downloading cloudflared..."
wget https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-arm -O cloudflared

echo "[+] Making it executable..."
chmod +x cloudflared

echo "[+] Moving to $PREFIX/bin/..."
mv cloudflared $PREFIX/bin/

echo "[✔] cloudflared installed successfully."
