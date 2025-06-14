# 🌐 Cloudflare Tunnel Manager

This project provides a set of scripts to manage Cloudflare Tunnels and DNS records, along with a user-friendly web interface for file management and AI assistance. Originally created for a Moto E5 Play device running Termux and proot-distro Ubuntu, these tools help manage Cloudflare infrastructure from mobile devices.

## 🛠️ Scripts

### 🔧 cfmanager.sh
This script is the main interface for managing Cloudflare Tunnels. It allows you to:
- 🚇 Create new tunnels
- ⏯️ Start, stop, and delete tunnels
- 📊 View the status of all tunnels
- 🖋️ Edit tunnel configuration files (YAML and JSON) for created tunnels
- 🌐 Automatically create DNS records for tunnels through cfdns.sh

**Important Notes:**
- Requires cloudflared to be installed first (use cfinstaller.sh)
- Automatically calls cfdns.sh to create CNAME records for subdomains
- Must be made executable: `chmod +x cfmanager.sh`

**Usage:**
1. Run the script: `./cfmanager.sh`
2. Follow the menu prompts to create, manage, or view tunnels

### 📥 cfinstaller.sh
This script simplifies the installation of cloudflared on Termux (Android) or similar environments.

**Important Notes:**
- Must be made executable: `chmod +x cfinstaller.sh`
- Should be run before using cfmanager.sh

**Usage:**
1. Run the script: `./cfinstaller.sh`
2. It will download and install the latest cloudflared binary

### 📝 cfdns.sh
This script manages DNS records in Cloudflare, specifically creating CNAME records for subdomains. It can be used either:
- 💬 Interactively as a standalone script
- 🤖 Automatically when called by cfmanager.sh

**Important Notes:**
- Must be made executable: `chmod +x cfdns.sh`
- Primarily used by cfmanager.sh for automatic DNS configuration

**Usage:**
1. Run the script: `./cfdns.sh`
2. Follow the prompts to create a DNS record

## 🗂️ Web File Manager & AI Assistant

The project includes a web-based interface that makes file management and system administration easier for non-technical users:

### 🔍 File Search & Management
- Chat-like interface for finding files: Just type what you're looking for
- Intuitive file browser with clickable directory structure
- Built-in editor for common file types (code, text, config files)
- Drag-and-drop file uploads

### 🤖 AI Assistant
- Powered by Google Gemini for instant help
- Ask questions about system administration tasks
- Get step-by-step guidance for complex operations
- AI can help troubleshoot issues and suggest solutions

### 💻 Web Terminal
- Execute commands directly in your browser
- View command output in real-time
- Safe execution environment with command validation

**How to Start the Web Interface:**
1. Make sure you have Go installed
2. Navigate to the project directory
3. Run the following command:
   ```bash
   go run main.go
   ```
4. Open your web browser and go to: http://localhost:12345

## 🔑 Environment Variables (.env)

The following environment variables are required for the scripts to function:

| Variable        | Description                                                                 |
|-----------------|-----------------------------------------------------------------------------|
| 🔐 CF_API_TOKEN    | Your Cloudflare API token with permissions to manage DNS and Tunnels        |
| 🆔 CF_ZONE_ID      | The Zone ID of your domain in Cloudflare                                    |
| 🏢 CF_ACCOUNT_ID   | Your Cloudflare Account ID                                                 |
| 🌍 CF_API_BASE     | The base URL for the Cloudflare API (default: https://api.cloudflare.com/client/v4) |
| 🌐 CF_DOMAIN       | The domain you want to manage (e.g., neptuno.uno)                          |

## ⚙️ How It Works

### 🚇 Tunnel Creation
- When you create a tunnel using cfmanager.sh, it:
  - 📄 Generates a configuration file
  - 🔐 Creates a credentials file
  - 🌐 Automatically calls cfdns.sh to create the necessary CNAME record

### 🌐 DNS Management
- The cfdns.sh script is integrated with cfmanager.sh to handle DNS records
- Creates CNAME records pointing to your tunnel subdomains

### ⚙️ Tunnel Management
- You can start, stop, or delete tunnels using the cfmanager.sh interface
- The status of all tunnels can be viewed in the dashboard

## 📋 Example Workflow
1. 📥 Install cloudflared using cfinstaller.sh
2. 🔧 Configure your .env file with the required Cloudflare credentials
3. 🚇 Run cfmanager.sh to create and manage tunnels (automatically handles DNS)
4. 🌐 Use cfdns.sh separately if you need manual DNS management
5. 🖥️ Access the web interface for file management and AI assistance

## 📱 Moto E5 Play Setup Guide

### Termux Installation
1. Install Termux from F-Droid
2. Update packages:
   ```bash
   pkg update && pkg upgrade
   ```
3. Install essential tools:
   ```bash
   pkg install git wget nano proot-distro
   ```

### Proot Ubuntu Setup
1. Install Ubuntu:
   ```bash
   proot-distro install ubuntu
   ```
2. Login to Ubuntu:
   ```bash
   proot-distro login ubuntu
   ```
3. Update Ubuntu:
   ```bash
   apt update && apt upgrade
   ```

### SSH Access Setup
1. Install OpenSSH in Termux:
   ```bash
   pkg install openssh
   ```
2. Set password:
   ```bash
   passwd
   ```
3. Start SSH server:
   ```bash
   sshd
   ```
4. Find your IP address:
   ```bash
   ifconfig
   ```
5. Connect via SSH:
   ```bash
   ssh <termux-username>@<device-ip> -p 8022
   ```

### File Transfer
Use Termius SFTP for easy file transfers:
- Install Termius on your computer
- Add your Termux SSH connection
- Use the SFTP tab for drag-and-drop file transfers

## 💡 About the Author
👨‍💼 Hi, I'm Jonathan Void! While I'm not a professional developer by trade (I work in Customer Success), I enjoy building tools to simplify complex processes. This project was born out of my own need to make Cloudflare Tunnel management more accessible for non-technical users like myself, especially when working from my Moto E5 Play device. If you find this tool helpful, feel free to use and share it with others!