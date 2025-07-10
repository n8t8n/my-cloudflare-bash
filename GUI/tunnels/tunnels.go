package tunnels

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/shirou/gopsutil/v3/process"
	"gopkg.in/yaml.v3"
)

type Tunnel struct {
	Name      string    `json:"name"`
	ID        string    `json:"id"`
	Port      int       `json:"port"`
	Domain    string    `json:"domain"`
	Status    string    `json:"status"`
	PID       int       `json:"pid"`
	CPU       float64   `json:"cpu"`
	Memory    float32   `json:"memory"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateTunnelRequest struct {
	Subdomain string `json:"subdomain"`
	Port      int    `json:"port"`
}

var configDir = filepath.Join(os.Getenv("HOME"), ".cloudflared")

func init() {
	os.MkdirAll(filepath.Join(configDir, "pids"), 0755)
}

func generatePetName() string {
	adjectives := []string{
		"morbid", "sarcastic", "chaotic", "deranged", "manic",
		"nihilistic", "twisted", "grotesque", "unhinged", "macabre",
		"sinister", "eerie", "hysterical", "toxic", "bleak",
		"lunatic", "cryptic", "damned", "grim", "volatile",
	}

	nouns := []string{
		"ghost", "zombie", "vampire", "skeleton", "demon",
		"witch", "phantom", "wraith", "specter", "goblin",
		"banshee", "ghoul", "mutant", "shade", "reaper",
	}

	adj := adjectives[time.Now().UnixNano()%int64(len(adjectives))]
	noun := nouns[time.Now().UnixNano()%int64(len(nouns))]

	return fmt.Sprintf("%s-%s", adj, noun)
}

func CreateTunnel(req CreateTunnelRequest) (*Tunnel, error) {
	domain := os.Getenv("CF_DOMAIN")
	if domain == "" {
		return nil, fmt.Errorf("CF_DOMAIN environment variable not set")
	}

	subdomain := req.Subdomain
	if subdomain == "" {
		subdomain = generatePetName()
	}

	tunnelName := fmt.Sprintf("%s-tunnel", subdomain)

	// Create tunnel using cloudflared
	cmd := exec.Command("cloudflared", "tunnel", "create", tunnelName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to create tunnel: %v, output: %s", err, string(output))
	}

	// Extract tunnel ID from output
	re := regexp.MustCompile(`[0-9a-f\-]{36}`)
	tunnelID := re.FindString(string(output))
	if tunnelID == "" {
		return nil, fmt.Errorf("could not extract tunnel ID from output: %s", string(output))
	}

	// Create config file
	configPath := filepath.Join(configDir, fmt.Sprintf("%s-config.yml", subdomain))
	credPath := filepath.Join(configDir, fmt.Sprintf("%s.json", tunnelID))

	configContent := fmt.Sprintf(`tunnel: "%s"
credentials-file: "%s"
ingress:
  - hostname: "%s.%s"
    icmp: false
    service: "http://0.0.0.0:%d"
  - service: http_status:404
`, tunnelID, credPath, subdomain, domain, req.Port)

	if err := ioutil.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		return nil, fmt.Errorf("failed to write config file: %v", err)
	}

	// Automatically create DNS record for the tunnel
	dnsTarget := fmt.Sprintf("%s.cfargotunnel.com", tunnelID)
	if err := createTunnelDNSRecord(subdomain, dnsTarget); err != nil {
		// Log the error but don't fail the tunnel creation
		fmt.Printf("Warning: Failed to create DNS record: %v\n", err)
	}

	tunnel := &Tunnel{
		Name:      subdomain,
		ID:        tunnelID,
		Port:      req.Port,
		Domain:    fmt.Sprintf("%s.%s", subdomain, domain),
		Status:    "stopped",
		CreatedAt: time.Now(),
	}

	return tunnel, nil
}

// createTunnelDNSRecord creates a CNAME record for the tunnel
func createTunnelDNSRecord(subdomain, target string) error {
	apiToken := os.Getenv("CF_API_TOKEN")
	zoneID := os.Getenv("CF_ZONE_ID")
	domain := os.Getenv("CF_DOMAIN")

	if apiToken == "" || zoneID == "" || domain == "" {
		return fmt.Errorf("missing required environment variables for DNS creation")
	}

	fullName := fmt.Sprintf("%s.%s", subdomain, domain)

	// Check if record already exists
	checkURL := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/dns_records?name=%s", zoneID, fullName)
	req, err := http.NewRequest("GET", checkURL, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+apiToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// If record already exists, return success
	if strings.Contains(string(body), fmt.Sprintf(`"name":"%s"`, fullName)) {
		return nil
	}

	// Create the DNS record
	createURL := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/dns_records", zoneID)
	payload := map[string]interface{}{
		"type":    "CNAME",
		"name":    fullName,
		"content": target,
		"ttl":     1,
		"proxied": true, // Enable Cloudflare proxy by default
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err = http.NewRequest("POST", createURL, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+apiToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err = client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// Check if creation was successful
	if !strings.Contains(string(body), `"success":true`) {
		return fmt.Errorf("failed to create DNS record: %s", string(body))
	}

	return nil
}

func ListTunnels() ([]*Tunnel, error) {
	var tunnels []*Tunnel

	files, err := ioutil.ReadDir(configDir)
	if err != nil {
		return tunnels, nil
	}

	for _, file := range files {
		if strings.HasSuffix(file.Name(), "-config.yml") {
			name := strings.TrimSuffix(file.Name(), "-config.yml")
			tunnel, err := getTunnelFromConfig(name)
			if err != nil {
				continue
			}
			tunnels = append(tunnels, tunnel)
		}
	}

	return tunnels, nil
}

func getTunnelFromConfig(name string) (*Tunnel, error) {
	configPath := filepath.Join(configDir, fmt.Sprintf("%s-config.yml", name))

	content, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	// Parse config to extract tunnel info
	lines := strings.Split(string(content), "\n")
	var tunnelID, hostname string
	var port int

	for i, line := range lines {
		line = strings.TrimSpace(line)

		// Extract tunnel ID
		if strings.HasPrefix(line, "tunnel:") {
			parts := strings.Split(line, ":")
			if len(parts) > 1 {
				tunnelID = strings.Trim(parts[1], ` "`)
			}
		}

		// Extract hostname from ingress section
		if strings.HasPrefix(line, "- hostname:") {
			parts := strings.Split(line, ":")
			if len(parts) > 1 {
				hostname = strings.Trim(parts[1], ` "`)
			}

			// Look for the service line that follows hostname
			for j := i + 1; j < len(lines) && j < i+5; j++ {
				serviceLine := strings.TrimSpace(lines[j])
				if strings.HasPrefix(serviceLine, "service:") && strings.Contains(serviceLine, "http://") {
					// Extract port from service line like: service: "http://0.0.0.0:30111"
					// Use regex to extract port number
					re := regexp.MustCompile(`:(\d+)`)
					matches := re.FindStringSubmatch(serviceLine)
					if len(matches) > 1 {
						if p, err := strconv.Atoi(matches[1]); err == nil {
							port = p
							break
						}
					}
				}
			}
		}
	}

	tunnel := &Tunnel{
		Name:   name,
		ID:     tunnelID,
		Port:   port,
		Domain: hostname,
		Status: "stopped",
	}

	// Check if tunnel is running
	pidPath := filepath.Join(configDir, "pids", fmt.Sprintf("%s.pid", name))
	if pidBytes, err := ioutil.ReadFile(pidPath); err == nil {
		if pid, err := strconv.Atoi(strings.TrimSpace(string(pidBytes))); err == nil {
			if proc, err := process.NewProcess(int32(pid)); err == nil {
				if running, _ := proc.IsRunning(); running {
					tunnel.Status = "running"
					tunnel.PID = pid

					// Get CPU and memory usage
					if cpuPercent, err := proc.CPUPercent(); err == nil {
						tunnel.CPU = cpuPercent
					}
					if memInfo, err := proc.MemoryInfo(); err == nil {
						tunnel.Memory = float32(memInfo.RSS) / 1024 / 1024 // MB
					}
				} else {
					os.Remove(pidPath)
				}
			}
		}
	}

	return tunnel, nil
}

func StartTunnel(name string) error {
	configPath := filepath.Join(configDir, fmt.Sprintf("%s-config.yml", name))
	pidPath := filepath.Join(configDir, "pids", fmt.Sprintf("%s.pid", name))

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return fmt.Errorf("tunnel config not found: %s", name)
	}

	// Stop existing process if running
	if pidBytes, err := ioutil.ReadFile(pidPath); err == nil {
		if pid, err := strconv.Atoi(strings.TrimSpace(string(pidBytes))); err == nil {
			if proc, err := os.FindProcess(pid); err == nil {
				proc.Signal(syscall.SIGTERM)
				time.Sleep(time.Second)
			}
		}
	}

	// Start new process
	cmd := exec.Command("cloudflared", "tunnel", "--config", configPath, "run")
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start tunnel: %v", err)
	}

	// Save PID
	if err := ioutil.WriteFile(pidPath, []byte(strconv.Itoa(cmd.Process.Pid)), 0644); err != nil {
		return fmt.Errorf("failed to save PID: %v", err)
	}

	return nil
}

func StopTunnel(name string) error {
	pidPath := filepath.Join(configDir, "pids", fmt.Sprintf("%s.pid", name))

	pidBytes, err := ioutil.ReadFile(pidPath)
	if err != nil {
		return fmt.Errorf("tunnel not running or PID file not found")
	}

	pid, err := strconv.Atoi(strings.TrimSpace(string(pidBytes)))
	if err != nil {
		return fmt.Errorf("invalid PID in file")
	}

	proc, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("process not found")
	}

	if err := proc.Signal(syscall.SIGTERM); err != nil {
		return fmt.Errorf("failed to stop process: %v", err)
	}

	os.Remove(pidPath)
	return nil
}

func DeleteTunnel(name string) error {
	// Stop tunnel first
	StopTunnel(name)

	// Delete tunnel using cloudflared
	tunnelName := fmt.Sprintf("%s-tunnel", name)
	cmd := exec.Command("cloudflared", "tunnel", "delete", tunnelName)
	if err := cmd.Run(); err != nil {
		// Continue even if delete fails
	}

	// Remove config files
	configPath := filepath.Join(configDir, fmt.Sprintf("%s-config.yml", name))
	os.Remove(configPath)

	// Remove PID file
	pidPath := filepath.Join(configDir, "pids", fmt.Sprintf("%s.pid", name))
	os.Remove(pidPath)

	return nil
}

func GetTunnelStatus(name string) (*Tunnel, error) {
	return getTunnelFromConfig(name)
}

func GetTunnelConfig(name string) (string, error) {
	configPath := filepath.Join(configDir, name+"-config.yml")

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return "", fmt.Errorf("tunnel config not found: %s", name)
	}

	content, err := ioutil.ReadFile(configPath)
	if err != nil {
		return "", fmt.Errorf("failed to read config file: %v", err)
	}

	return string(content), nil
}

func UpdateTunnelConfig(name, config string) error {
	configPath := filepath.Join(configDir, name+"-config.yml")

	// Validate YAML syntax before writing
	var testConfig map[string]interface{}
	if err := yaml.Unmarshal([]byte(config), &testConfig); err != nil {
		return fmt.Errorf("invalid YAML syntax: %v", err)
	}

	// Write the updated config
	if err := ioutil.WriteFile(configPath, []byte(config), 0644); err != nil {
		return fmt.Errorf("failed to write config file: %v", err)
	}

	return nil
}
