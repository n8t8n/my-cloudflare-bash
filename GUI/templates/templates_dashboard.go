package templates

const DashboardTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Cloudflare Manager Dashboard</title>
<style>
* {
  box-sizing: border-box;
  margin: 0;
  padding: 0;
}

body {
  background: #1a1a2e;
  color: #eee;
  font-family: 'Courier New', monospace;
  height: 100vh;
  overflow: hidden;
}

.container {
  padding: 1rem;
  height: 100vh;
  display: flex;
  flex-direction: column;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1rem;
  border-bottom: 1px solid #16213e;
  padding-bottom: 1rem;
}

.title {
  color: #ff6b35;
  font-size: 1.2rem;
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.actions {
  position: relative;
}

.dropdown {
  position: relative;
  display: inline-block;
}

.dropdown-btn {
  background: none;
  border: 1px solid #0f3460;
  color: #0f3460;
  padding: 0.3rem 0.8rem;
  font-family: monospace;
  font-size: 0.8rem;
  cursor: pointer;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.dropdown-btn:hover {
  background: #0f3460;
  color: #eee;
}

.dropdown-content {
  display: none;
  position: absolute;
  right: 0;
  background: #1a1a2e;
  border: 1px solid #16213e;
  min-width: 160px;
  z-index: 1000;
  box-shadow: 0px 8px 16px 0px rgba(0,0,0,0.2);
}

.dropdown-content.show {
  display: block;
}

.dropdown-item {
  color: #eee;
  padding: 0.5rem 1rem;
  text-decoration: none;
  display: block;
  font-size: 0.8rem;
  cursor: pointer;
  border: none;
  background: none;
  width: 100%;
  text-align: left;
  font-family: monospace;
}

.dropdown-item:hover {
  background: #16213e;
  color: #ff6b35;
}

.tabs {
  display: flex;
  border-bottom: 1px solid #16213e;
  margin-bottom: 1rem;
}

.tab {
  padding: 0.5rem 1rem;
  background: none;
  border: none;
  color: #888;
  cursor: pointer;
  font-family: monospace;
  font-size: 0.8rem;
  transition: all 0.2s;
}

.tab.active {
  color: #ff6b35;
  border-bottom: 2px solid #ff6b35;
}

.tab-content {
  display: none;
  flex: 1;
  overflow-y: auto;
}

.tab-content.active {
  display: block;
}

.section {
  border: 1px solid #16213e;
  background: #0f0f23;
  padding: 1rem;
  margin-bottom: 1rem;
  border-radius: 4px;
}

.section-header {
  color: #ff6b35;
  font-size: 0.9rem;
  margin-bottom: 0.5rem;
  text-transform: uppercase;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.btn {
  background: none;
  border: 1px solid;
  color: inherit;
  padding: 0.3rem 0.8rem;
  font-family: monospace;
  font-size: 0.7rem;
  cursor: pointer;
  transition: all 0.2s;
  border-radius: 2px;
}

.btn-primary { border-color: #0f3460; color: #0f3460; }
.btn-secondary { border-color: #ff6b35; color: #ff6b35; }
.btn-danger { border-color: #ff4757; color: #ff4757; }
.btn-success { border-color: #2ed573; color: #2ed573; }

.btn:hover {
  background: currentColor;
  color: #1a1a2e;
}

.btn-small {
  padding: 0.2rem 0.4rem;
  font-size: 0.6rem;
}

.table {
  width: 100%;
  border-collapse: collapse;
}

.table th,
.table td {
  padding: 0.5rem;
  text-align: left;
  border-bottom: 1px solid #16213e;
  font-size: 0.7rem;
}

.table th {
  color: #ff6b35;
  font-weight: bold;
}

.status-running { color: #2ed573; }
.status-stopped { color: #ff4757; }
.status-proxied { color: #ff6b35; }
.status-dns { color: #0f3460; }

.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background: rgba(0, 0, 0, 0.8);
  z-index: 1000;
  display: none;
}

.modal {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  background: #1a1a2e;
  border: 2px solid #ff6b35;
  padding: 1.5rem;
  min-width: 400px;
  max-width: 90vw;
  max-height: 90vh;
  overflow-y: auto;
  border-radius: 4px;
}

.modal-header {
  color: #ff6b35;
  margin-bottom: 1rem;
  font-size: 1rem;
}

.form-group {
  margin-bottom: 1rem;
}

.form-label {
  display: block;
  color: #eee;
  margin-bottom: 0.3rem;
  font-size: 0.8rem;
}

.form-input,
.form-select {
  width: 100%;
  background: transparent;
  border: 1px solid #16213e;
  color: #eee;
  padding: 0.5rem;
  font-family: monospace;
  font-size: 0.8rem;
}

.form-input:focus,
.form-select:focus {
  outline: none;
  border-color: #ff6b35;
}

.form-checkbox {
  margin-right: 0.5rem;
}

.modal-actions {
  display: flex;
  gap: 0.5rem;
  justify-content: flex-end;
  margin-top: 1rem;
}

.toast {
  position: fixed;
  top: 1rem;
  right: 1rem;
  background: #2ed573;
  color: #1a1a2e;
  padding: 0.8rem 1.2rem;
  border-radius: 4px;
  font-size: 0.8rem;
  font-weight: bold;
  z-index: 2000;
  transform: translateX(100%);
  opacity: 0;
  transition: all 0.3s ease;
}

.toast.show {
  transform: translateX(0);
  opacity: 1;
}

.toast.error {
  background: #ff4757;
  color: white;
}

.toast.warning {
  background: #ffa502;
  color: #1a1a2e;
}

.empty-state {
  text-align: center;
  color: #888;
  font-style: italic;
  padding: 2rem;
}

.stats {
  display: flex;
  flex-wrap: wrap;
  gap: 1rem;
  margin-bottom: 1rem;
}

.stat-box {
  border: 1px solid #16213e;
  padding: 1rem;
  background: #0f0f23;
  flex: 1;
  min-width: 150px;
  border-radius: 4px;
}

.stat-label {
  color: #888;
  font-size: 0.7rem;
  text-transform: uppercase;
}

.stat-value {
  color: #ff6b35;
  font-size: 1.2rem;
  font-weight: bold;
  margin: 0.2rem 0;
}

.dropdown-arrow {
  font-size: 0.6rem;
  transition: transform 0.2s;
}

.dropdown.active .dropdown-arrow {
  transform: rotate(180deg);
}
</style>
</head>
<body>
  <div class="container">
    <div class="header">
      <div class="title">
        ☁️ CF-MANAGER
        <span style="font-size: 0.8rem; color: #888;">admin@cloudflare:~/dashboard</span>
      </div>
      <div class="actions">
        <div class="dropdown">
          <button class="dropdown-btn" onclick="toggleDropdown()">
            ACTIONS <span class="dropdown-arrow">▼</span>
          </button>
          <div class="dropdown-content" id="dropdown-menu">
            <button class="dropdown-item" onclick="showCreateDNSModal()">Create DNS Record</button>
            <button class="dropdown-item" onclick="showCreateTunnelModal()">Create Tunnel</button>
            <button class="dropdown-item" onclick="refreshAll()">Refresh All</button>
            <button class="dropdown-item" onclick="logout()">Logout</button>
            <button class="dropdown-item" onclick="showChangePasswordModal()">Change Password</button>
          </div>
        </div>
      </div>
    </div>

    <div class="stats">
      <div class="stat-box">
        <div class="stat-label">DNS Records</div>
        <div class="stat-value" id="dns-count">0</div>
      </div>
      <div class="stat-box">
        <div class="stat-label">Active Tunnels</div>
        <div class="stat-value" id="tunnel-count">0</div>
      </div>
      <div class="stat-box">
        <div class="stat-label">Running Tunnels</div>
        <div class="stat-value" id="running-count">0</div>
      </div>
    </div>

    <div class="tabs">
      <button class="tab active" onclick="switchTab('dns')">DNS RECORDS</button>
      <button class="tab" onclick="switchTab('tunnels')">TUNNELS</button>
    </div>

    <div id="dns-tab" class="tab-content active">
      <div class="section">
        <div class="section-header">
          DNS Records
          <button class="btn btn-primary btn-small" onclick="showCreateDNSModal()">+ ADD RECORD</button>
        </div>
        <div id="dns-records-container">
          <div class="empty-state">Loading DNS records...</div>
        </div>
      </div>
    </div>

    <div id="tunnels-tab" class="tab-content">
      <div class="section">
        <div class="section-header">
          Cloudflared Tunnels
          <button class="btn btn-primary btn-small" onclick="showCreateTunnelModal()">+ CREATE TUNNEL</button>
        </div>
        <div id="tunnels-container">
          <div class="empty-state">Loading tunnels...</div>
        </div>
      </div>
    </div>
  </div>

  <!-- Create DNS Record Modal -->
  <div class="modal-overlay" id="dns-modal">
    <div class="modal">
      <div class="modal-header">Create DNS Record</div>
      <form id="dns-form">
        <div class="form-group">
          <label class="form-label">Subdomain</label>
          <input type="text" class="form-input" id="dns-subdomain" placeholder="app" required>
        </div>
        <div class="form-group">
          <label class="form-label">Record Type</label>
          <select class="form-select" id="dns-type" onchange="updateDNSTargetLabel()">
            <option value="A">A Record</option>
            <option value="CNAME">CNAME Record</option>
          </select>
        </div>
        <div class="form-group">
          <label class="form-label" id="dns-target-label">IP Address</label>
          <input type="text" class="form-input" id="dns-target" placeholder="192.168.1.10" required>
        </div>
        <div class="form-group">
          <label class="form-label">
            <input type="checkbox" class="form-checkbox" id="dns-proxied">
            Enable Cloudflare Proxy (Orange Cloud)
          </label>
        </div>
        <div class="modal-actions">
          <button type="submit" class="btn btn-primary">CREATE</button>
          <button type="button" class="btn btn-secondary" onclick="closeModal('dns-modal')">CANCEL</button>
        </div>
      </form>
    </div>
  </div>

  <!-- Edit DNS Record Modal -->
  <div class="modal-overlay" id="edit-dns-modal">
    <div class="modal">
      <div class="modal-header">Edit DNS Record</div>
      <form id="edit-dns-form">
        <input type="hidden" id="edit-dns-id">
        <div class="form-group">
          <label class="form-label">Name</label>
          <input type="text" class="form-input" id="edit-dns-name" readonly>
        </div>
        <div class="form-group">
          <label class="form-label">Record Type</label>
          <select class="form-select" id="edit-dns-type" onchange="updateEditDNSTargetLabel()">
            <option value="A">A Record</option>
            <option value="CNAME">CNAME Record</option>
            <option value="MX">MX Record</option>
            <option value="TXT">TXT Record</option>
          </select>
        </div>
        <div class="form-group">
          <label class="form-label" id="edit-dns-target-label">Content</label>
          <input type="text" class="form-input" id="edit-dns-content" required>
        </div>
        <div class="form-group">
          <label class="form-label">TTL</label>
          <select class="form-select" id="edit-dns-ttl">
            <option value="1">Auto</option>
            <option value="300">5 minutes</option>
            <option value="600">10 minutes</option>
            <option value="1800">30 minutes</option>
            <option value="3600">1 hour</option>
            <option value="86400">1 day</option>
          </select>
        </div>
        <div class="form-group">
          <label class="form-label">
            <input type="checkbox" class="form-checkbox" id="edit-dns-proxied">
            Enable Cloudflare Proxy (Orange Cloud)
          </label>
        </div>
        <div class="modal-actions">
          <button type="submit" class="btn btn-primary">UPDATE</button>
          <button type="button" class="btn btn-secondary" onclick="closeModal('edit-dns-modal')">CANCEL</button>
        </div>
      </form>
    </div>
  </div>

  <!-- Create Tunnel Modal -->
  <div class="modal-overlay" id="tunnel-modal">
    <div class="modal">
      <div class="modal-header">Create Cloudflared Tunnel</div>
      <form id="tunnel-form">
        <div class="form-group">
          <label class="form-label">Subdomain (leave blank for random name)</label>
          <input type="text" class="form-input" id="tunnel-subdomain" placeholder="app">
        </div>
        <div class="form-group">
          <label class="form-label">Local Port</label>
          <input type="number" class="form-input" id="tunnel-port" placeholder="3000" required>
        </div>
        <div class="modal-actions">
          <button type="submit" class="btn btn-primary">CREATE TUNNEL</button>
          <button type="button" class="btn btn-secondary" onclick="closeModal('tunnel-modal')">CANCEL</button>
        </div>
      </form>
    </div>
  </div>

  <!-- Change Password Modal -->
  <div class="modal-overlay" id="change-password-modal">
    <div class="modal">
      <div class="modal-header">Change Password</div>
      <form id="change-password-form">
        <div class="form-group">
          <label class="form-label">Old Password</label>
          <input type="password" class="form-input" id="old-password" required>
        </div>
        <div class="form-group">
          <label class="form-label">New Password</label>
          <input type="password" class="form-input" id="new-password" required>
        </div>
        <div class="modal-actions">
          <button type="submit" class="btn btn-primary">CHANGE PASSWORD</button>
          <button type="button" class="btn btn-secondary" onclick="closeModal('change-password-modal')">CANCEL</button>
        </div>
      </form>
    </div>
  </div>

  <!-- YAML Editor Modal -->
  <div id="yaml-modal" class="modal">
    <div class="modal-content">
      <div class="modal-header">
        <h2>Edit Tunnel Configuration</h2>
        <span class="close" onclick="closeModal('yaml-modal')">&times;</span>
      </div>
      <div class="modal-body">
        <div class="form-group">
          <label for="yaml-content">YAML Configuration:</label>
          <textarea id="yaml-content" rows="20" placeholder="Enter YAML configuration..."></textarea>
        </div>
        <div class="form-actions">
          <button class="btn btn-secondary" onclick="closeModal('yaml-modal')">Cancel</button>
          <button class="btn btn-success" onclick="saveYamlConfig()">Save Configuration</button>
        </div>
      </div>
    </div>
  </div>

  <div id="toast"></div>

  <script>
    let dnsRecords = [];
    let tunnels = [];

    document.addEventListener('DOMContentLoaded', function() {
      fetchDNSRecords();
      fetchTunnels();
      // Removed automatic polling for Termux optimization
      // setInterval(fetchTunnels, 5000); // Refresh tunnels every 5 seconds
    });

    // Close dropdown when clicking outside
    document.addEventListener('click', function(event) {
      const dropdown = document.querySelector('.dropdown');
      if (!dropdown.contains(event.target)) {
        document.getElementById('dropdown-menu').classList.remove('show');
        dropdown.classList.remove('active');
      }
    });

    function toggleDropdown() {
      const dropdown = document.querySelector('.dropdown');
      const menu = document.getElementById('dropdown-menu');
      menu.classList.toggle('show');
      dropdown.classList.toggle('active');
    }

    function switchTab(tabName) {
      document.querySelectorAll('.tab').forEach(tab => tab.classList.remove('active'));
      document.querySelectorAll('.tab-content').forEach(content => content.classList.remove('active'));
      
      document.querySelector('.tab[onclick="switchTab(\'' + tabName + '\')"]').classList.add('active');
      document.getElementById(tabName + '-tab').classList.add('active');
      
      // Refresh data when switching tabs for better UX
      if (tabName === 'tunnels') {
        fetchTunnels();
      } else if (tabName === 'dns') {
        fetchDNSRecords();
      }
    }

    // Debounced refresh to prevent excessive API calls
    let refreshTimeout;
    function debouncedRefresh() {
      clearTimeout(refreshTimeout);
      refreshTimeout = setTimeout(() => {
        fetchTunnels();
        fetchDNSRecords();
      }, 2000); // Wait 2 seconds after last action
    }

    async function fetchDNSRecords() {
      try {
        const response = await fetch('/dns/records');
        const data = await response.json();
        console.log('DNS Records received:', data); // Debug log
        dnsRecords = data || [];
        renderDNSRecords();
        updateStats();
      } catch (error) {
        console.error('DNS fetch error:', error); // Debug log
        showToast('Failed to fetch DNS records', 'error');
      }
    }

    async function fetchTunnels() {
      try {
        const response = await fetch('/tunnels');
        const data = await response.json();
        console.log('Tunnels received:', data); // Debug log
        
        // Only update if data changed (optimization for Termux)
        if (JSON.stringify(data) !== JSON.stringify(tunnels)) {
          tunnels = data || [];
          renderTunnels();
          updateStats();
        }
      } catch (error) {
        console.error('Tunnels fetch error:', error); // Debug log
        showToast('Failed to fetch tunnels', 'error');
      }
    }

    function renderDNSRecords() {
      const container = document.getElementById('dns-records-container');
      
      if (dnsRecords.length === 0) {
        container.innerHTML = '<div class="empty-state">No DNS records found</div>';
        return;
      }

      const table = '<table class="table">' +
        '<thead><tr><th>Name</th><th>Type</th><th>Content</th><th>TTL</th><th>Proxy</th><th>Actions</th></tr></thead>' +
        '<tbody>' +
        dnsRecords.map(record => 
          '<tr>' +
          '<td>' + (record.name || 'N/A') + '</td>' +
          '<td>' + (record.type || 'N/A') + '</td>' +
          '<td>' + (record.content || 'N/A') + '</td>' +
          '<td>' + (record.ttl === 1 ? 'Auto' : record.ttl) + '</td>' +
          '<td><span class="' + (record.proxied ? 'status-proxied' : 'status-dns') + '">' + 
          (record.proxied ? 'PROXIED' : 'DNS ONLY') + '</span></td>' +
          '<td>' +
          '<button class="btn btn-secondary btn-small" onclick="editDNSRecord(\'' + record.id + '\')">EDIT</button> ' +
          '<button class="btn btn-danger btn-small" onclick="deleteDNSRecord(\'' + record.id + '\')">DELETE</button>' +
          '</td>' +
          '</tr>'
        ).join('') +
        '</tbody></table>';
      
      container.innerHTML = table;
    }

    function renderTunnels() {
      const container = document.getElementById('tunnels-container');
      
      if (tunnels.length === 0) {
        container.innerHTML = '<div class="empty-state">No tunnels configured</div>';
        return;
      }

      const table = '<table class="table">' +
        '<thead><tr><th>Name</th><th>Domain</th><th>Port</th><th>Status</th><th>CPU%</th><th>MEM MB</th><th>Actions</th></tr></thead>' +
        '<tbody>' +
        tunnels.map(tunnel => 
          '<tr>' +
          '<td>' + (tunnel.name || 'N/A') + '</td>' +
          '<td>' + (tunnel.domain || 'N/A') + '</td>' +
          '<td>' + (tunnel.port || 'N/A') + '</td>' +
          '<td><span class="status-' + tunnel.status + '">' + tunnel.status.toUpperCase() + '</span></td>' +
          '<td>' + (tunnel.cpu ? tunnel.cpu.toFixed(1) : 'N/A') + '</td>' +
          '<td>' + (tunnel.memory ? tunnel.memory.toFixed(1) : 'N/A') + '</td>' +
          '<td>' +
          (tunnel.status === 'stopped' ? 
            '<button class="btn btn-success btn-small" onclick="startTunnel(\'' + tunnel.name + '\')">START</button>' :
            '<button class="btn btn-warning btn-small" onclick="stopTunnel(\'' + tunnel.name + '\')">STOP</button>'
          ) +
          ' <button class="btn btn-secondary btn-small" onclick="editTunnelConfig(\'' + tunnel.name + '\')">EDIT YAML</button> ' +
          '<button class="btn btn-danger btn-small" onclick="deleteTunnel(\'' + tunnel.name + '\')">DELETE</button>' +
          '</td>' +
          '</tr>'
        ).join('') +
        '</tbody></table>';
      
      container.innerHTML = table;
    }

    function updateStats() {
      document.getElementById('dns-count').textContent = dnsRecords.length;
      document.getElementById('tunnel-count').textContent = tunnels.length;
      document.getElementById('running-count').textContent = tunnels.filter(t => t.status === 'running').length;
    }

    function showCreateDNSModal() {
      document.getElementById('dns-form').reset();
      updateDNSTargetLabel();
      openModal('dns-modal');
    }

    function showCreateTunnelModal() {
      document.getElementById('tunnel-form').reset();
      openModal('tunnel-modal');
    }

    function updateDNSTargetLabel() {
      const type = document.getElementById('dns-type').value;
      const label = document.getElementById('dns-target-label');
      const input = document.getElementById('dns-target');
      
      if (type === 'A') {
        label.textContent = 'IP Address';
        input.placeholder = '192.168.1.10';
      } else {
        label.textContent = 'Target Host';
        input.placeholder = 'example.com';
      }
    }

    document.getElementById('dns-form').addEventListener('submit', async function(e) {
      e.preventDefault();
      
      const data = {
        subdomain: document.getElementById('dns-subdomain').value,
        type: document.getElementById('dns-type').value,
        target: document.getElementById('dns-target').value,
        proxied: document.getElementById('dns-proxied').checked
      };

      try {
        const response = await fetch('/dns/records', {
          method: 'POST',
          headers: {'Content-Type': 'application/json'},
          body: JSON.stringify(data)
        });

        const result = await response.json();
        
        if (response.ok) {
          showToast('DNS record created successfully', 'success');
          closeModal('dns-modal');
          fetchDNSRecords();
        } else {
          showToast(result.error || 'Failed to create DNS record', 'error');
        }
      } catch (error) {
        showToast('Server error', 'error');
      }
    });

    function editDNSRecord(recordId) {
      const record = dnsRecords.find(r => r.id === recordId);
      if (!record) {
        showToast('Record not found', 'error');
        return;
      }

      document.getElementById('edit-dns-id').value = record.id;
      document.getElementById('edit-dns-name').value = record.name;
      document.getElementById('edit-dns-type').value = record.type;
      document.getElementById('edit-dns-content').value = record.content;
      document.getElementById('edit-dns-ttl').value = record.ttl;
      document.getElementById('edit-dns-proxied').checked = record.proxied;
      
      updateEditDNSTargetLabel();
      openModal('edit-dns-modal');
    }

    function updateEditDNSTargetLabel() {
      const type = document.getElementById('edit-dns-type').value;
      const label = document.getElementById('edit-dns-target-label');
      
      switch(type) {
        case 'A':
          label.textContent = 'IP Address';
          break;
        case 'CNAME':
          label.textContent = 'Target Host';
          break;
        case 'MX':
          label.textContent = 'Mail Server';
          break;
        case 'TXT':
          label.textContent = 'Text Content';
          break;
        default:
          label.textContent = 'Content';
      }
    }

    document.getElementById('edit-dns-form').addEventListener('submit', async function(e) {
      e.preventDefault();
      
      const recordId = document.getElementById('edit-dns-id').value;
      const data = {
        type: document.getElementById('edit-dns-type').value,
        content: document.getElementById('edit-dns-content').value,
        ttl: parseInt(document.getElementById('edit-dns-ttl').value),
        proxied: document.getElementById('edit-dns-proxied').checked
      };

      try {
        const response = await fetch('/dns/records/' + recordId, {
          method: 'PUT',
          headers: {'Content-Type': 'application/json'},
          body: JSON.stringify(data)
        });

        const result = await response.json();
        
        if (response.ok) {
          showToast('DNS record updated successfully', 'success');
          closeModal('edit-dns-modal');
          fetchDNSRecords();
        } else {
          showToast(result.error || 'Failed to update DNS record', 'error');
        }
      } catch (error) {
        showToast('Server error', 'error');
      }
    });

    document.getElementById('tunnel-form').addEventListener('submit', async function(e) {
      e.preventDefault();
      
      const data = {
        subdomain: document.getElementById('tunnel-subdomain').value,
        port: parseInt(document.getElementById('tunnel-port').value)
      };

      try {
        const response = await fetch('/tunnels', {
          method: 'POST',
          headers: {'Content-Type': 'application/json'},
          body: JSON.stringify(data)
        });

        const result = await response.json();
        
        if (response.ok) {
          showToast(result.message || 'Tunnel created successfully', 'success');
          closeModal('tunnel-modal');
          fetchTunnels();
        } else {
          showToast(result.error || 'Failed to create tunnel', 'error');
        }
      } catch (error) {
        showToast('Server error', 'error');
      }
    });

    async function deleteDNSRecord(recordId) {
      if (!confirm('Delete this DNS record?')) return;

      try {
        const response = await fetch('/dns/records/' + recordId, {
          method: 'DELETE'
        });

        if (response.ok) {
          showToast('DNS record deleted', 'success');
          fetchDNSRecords();
        } else {
          showToast('Failed to delete DNS record', 'error');
        }
      } catch (error) {
        showToast('Server error', 'error');
      }
    }

    async function startTunnel(name) {
      try {
        const response = await fetch('/tunnels/' + name + '/start', {
          method: 'POST'
        });

        if (response.ok) {
          showToast('Tunnel started', 'success');
          // Wait 1 second for process to start, then update
          setTimeout(fetchTunnels, 1000);
        } else {
          showToast('Failed to start tunnel', 'error');
        }
      } catch (error) {
        showToast('Server error', 'error');
      }
    }

    async function stopTunnel(name) {
      try {
        const response = await fetch('/tunnels/' + name + '/stop', {
          method: 'POST'
        });

        if (response.ok) {
          showToast('Tunnel stopped', 'success');
          // Wait 1 second for process to stop, then update
          setTimeout(fetchTunnels, 1000);
        } else {
          showToast('Failed to stop tunnel', 'error');
        }
      } catch (error) {
        showToast('Server error', 'error');
      }
    }

    async function deleteTunnel(name) {
      if (!confirm('Delete tunnel "' + name + '"? This will also stop it if running.')) return;

      try {
        const response = await fetch('/tunnels/' + name, {
          method: 'DELETE'
        });

        if (response.ok) {
          showToast('Tunnel deleted', 'success');
          fetchTunnels();
        } else {
          showToast('Failed to delete tunnel', 'error');
        }
      } catch (error) {
        showToast('Server error', 'error');
      }
    }

    function refreshAll() {
      fetchDNSRecords();
      fetchTunnels();
      showToast('Data refreshed', 'success');
    }

    // Add refresh button to header for easy access
    function addRefreshButton() {
      const header = document.querySelector('.header');
      if (!document.getElementById('refresh-btn')) {
        const refreshBtn = document.createElement('button');
        refreshBtn.id = 'refresh-btn';
        refreshBtn.className = 'btn btn-secondary btn-small';
        refreshBtn.innerHTML = '🔄 REFRESH';
        refreshBtn.onclick = refreshAll;
        refreshBtn.style.marginLeft = '1rem';
        header.appendChild(refreshBtn);
      }
    }

    // Initialize refresh button
    document.addEventListener('DOMContentLoaded', function() {
      addRefreshButton();
    });

    function logout() {
      showToast('Logging out...', 'warning');
      setTimeout(() => {
        window.location.href = '/logout';
      }, 1000);
    }

    function openModal(modalId) {
      document.getElementById(modalId).style.display = 'block';
      document.getElementById(modalId).onclick = function(e) {
        if (e.target === this) closeModal(modalId);
      };
    }

    function closeModal(modalId) {
      document.getElementById(modalId).style.display = 'none';
    }

    function showToast(message, type = 'success') {
      const toast = document.getElementById('toast');
      toast.textContent = message;
      toast.className = 'toast ' + type;

      setTimeout(() => toast.classList.add('show'), 100);

      setTimeout(() => {
        toast.classList.remove('show');
      }, 3000);
    }

    function showChangePasswordModal() {
      document.getElementById('change-password-form').reset();
      openModal('change-password-modal');
    }

    document.getElementById('change-password-form').addEventListener('submit', async function(e) {
      e.preventDefault();

      const oldPassword = document.getElementById('old-password').value;
      const newPassword = document.getElementById('new-password').value;

      try {
        const response = await fetch('/change-password', {
          method: 'POST',
          headers: {'Content-Type': 'application/json'},
          body: JSON.stringify({ oldPassword, newPassword })
        });

        const result = await response.json();
        
        if (result.success) {
          showToast('Password changed successfully', 'success');
          closeModal('change-password-modal');
        } else {
          showToast(result.error || 'Failed to change password', 'error');
        }
      } catch (error) {
        showToast('Server error', 'error');
      }
    });

    let currentYamlTunnel = null;

    async function editTunnelConfig(name) {
      try {
        const response = await fetch('/tunnels/' + name + '/config');
        const result = await response.json();
        if (response.ok && result.config) {
          document.getElementById('yaml-content').value = result.config;
          currentYamlTunnel = name;
          openModal('yaml-modal');
        } else {
          showToast(result.error || 'Failed to load config', 'error');
        }
      } catch (error) {
        showToast('Server error', 'error');
      }
    }

    async function saveYamlConfig() {
      if (!currentYamlTunnel) return;
      const config = document.getElementById('yaml-content').value;
      try {
        const response = await fetch('/tunnels/' + currentYamlTunnel + '/config', {
          method: 'PUT',
          headers: {'Content-Type': 'application/json'},
          body: JSON.stringify({ config })
        });
        const result = await response.json();
        if (response.ok && result.success) {
          showToast('Config updated', 'success');
          closeModal('yaml-modal');
          fetchTunnels();
        } else {
          showToast(result.error || 'Failed to update config', 'error');
        }
      } catch (error) {
        showToast('Server error', 'error');
      }
    }

    document.addEventListener('keydown', function(e) {
      if (e.key === 'Escape') {
        closeModal('dns-modal');
        closeModal('edit-dns-modal');
        closeModal('tunnel-modal');
        closeModal('change-password-modal');
      }
    });
  </script>
</body>
</html>`
