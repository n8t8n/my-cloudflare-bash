// Dashboard JavaScript
let currentTab = 'dns';

// Initialize dashboard
document.addEventListener('DOMContentLoaded', function() {
  fetchDNSRecords();
  fetchTunnels();
  
  // Set up form event listeners
  setupFormListeners();
  
  // Set up keyboard shortcuts
  setupKeyboardShortcuts();
});

function setupFormListeners() {
  // DNS form
  document.getElementById('dns-form').addEventListener('submit', async function(e) {
    e.preventDefault();
    
    const data = {
      subdomain: document.getElementById('dns-subdomain').value,
      type: document.getElementById('dns-type').value,
      target: document.getElementById('dns-target').value,
      proxied: document.getElementById('dns-proxied').checked
    };

    try {
      const response = await fetch('/api/dns/records', {
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

  // Edit DNS form
  document.getElementById('edit-dns-form').addEventListener('submit', async function(e) {
    e.preventDefault();
    
    const recordId = document.getElementById('edit-dns-id').value;
    const data = {
      type: document.getElementById('edit-dns-type').value,
      content: document.getElementById('edit-dns-target').value,
      proxied: document.getElementById('edit-dns-proxied').checked
    };

    try {
      const response = await fetch('/api/dns/records/' + recordId, {
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

  // Tunnel form
  document.getElementById('tunnel-form').addEventListener('submit', async function(e) {
    e.preventDefault();
    
    const data = {
      subdomain: document.getElementById('tunnel-subdomain').value,
      port: parseInt(document.getElementById('tunnel-port').value)
    };

    try {
      const response = await fetch('/api/tunnels', {
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

  // Change password form
  document.getElementById('change-password-form').addEventListener('submit', async function(e) {
    e.preventDefault();

    const oldPassword = document.getElementById('old-password').value;
    const newPassword = document.getElementById('new-password').value;

    try {
      const response = await fetch('/api/change-password', {
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
}

function setupKeyboardShortcuts() {
  document.addEventListener('keydown', function(e) {
    if (e.key === 'Escape') {
      closeModal('dns-modal');
      closeModal('edit-dns-modal');
      closeModal('tunnel-modal');
      closeModal('change-password-modal');
    }
  });
}

// Tab switching
function switchTab(tabName) {
  // Update tab buttons
  document.querySelectorAll('.tab').forEach(tab => tab.classList.remove('active'));
  event.target.classList.add('active');
  
  // Update tab content
  document.querySelectorAll('.tab-content').forEach(content => content.classList.remove('active'));
  document.getElementById(tabName + '-content').classList.add('active');
  
  currentTab = tabName;
}

// Dropdown functionality
function toggleDropdown() {
  const dropdown = document.querySelector('.dropdown');
  const menu = document.getElementById('dropdown-menu');
  
  dropdown.classList.toggle('active');
  menu.classList.toggle('show');
}

// Close dropdown when clicking outside
document.addEventListener('click', function(e) {
  if (!e.target.closest('.dropdown')) {
    document.querySelector('.dropdown').classList.remove('active');
    document.getElementById('dropdown-menu').classList.remove('show');
  }
});

// DNS Functions
async function fetchDNSRecords() {
  try {
    const response = await fetch('/api/dns/records');
    const records = await response.json();
    
    if (response.ok) {
      renderDNSRecords(records);
      document.getElementById('dns-count').textContent = records.length;
    } else {
      showToast('Failed to fetch DNS records', 'error');
    }
  } catch (error) {
    showToast('Server error', 'error');
  }
}

function renderDNSRecords(records) {
  const container = document.getElementById('dns-table-container');
  
  if (records.length === 0) {
    container.innerHTML = '<div class="empty-state">No DNS records found</div>';
    return;
  }
  
  const table = `
    <table class="table">
      <thead>
        <tr>
          <th>Name</th>
          <th>Type</th>
          <th>Content</th>
          <th>Proxy</th>
          <th>Actions</th>
        </tr>
      </thead>
      <tbody>
        ${records.map(record => `
          <tr>
            <td>${record.name}</td>
            <td>${record.type}</td>
            <td>${record.content}</td>
            <td><span class="status-${record.proxied ? 'proxied' : 'dns'}">${record.proxied ? 'ON' : 'OFF'}</span></td>
            <td>
              <button class="btn btn-secondary btn-small" onclick="editDNSRecord('${record.id}', '${record.type}', '${record.content}', ${record.proxied})">Edit</button>
              <button class="btn btn-danger btn-small" onclick="deleteDNSRecord('${record.id}')">Delete</button>
            </td>
          </tr>
        `).join('')}
      </tbody>
    </table>
  `;
  
  container.innerHTML = table;
}

async function deleteDNSRecord(recordId) {
  if (!confirm('Delete this DNS record?')) return;

  try {
    const response = await fetch('/api/dns/records/' + recordId, {
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

function editDNSRecord(recordId, type, content, proxied) {
  document.getElementById('edit-dns-id').value = recordId;
  document.getElementById('edit-dns-type').value = type;
  document.getElementById('edit-dns-target').value = content;
  document.getElementById('edit-dns-proxied').checked = proxied;
  openModal('edit-dns-modal');
}

// Tunnel Functions
async function fetchTunnels() {
  try {
    const response = await fetch('/api/tunnels');
    const tunnels = await response.json();
    
    if (response.ok) {
      renderTunnels(tunnels);
      document.getElementById('tunnel-count').textContent = tunnels.length;
      document.getElementById('running-count').textContent = tunnels.filter(t => t.status === 'running').length;
    } else {
      showToast('Failed to fetch tunnels', 'error');
    }
  } catch (error) {
    showToast('Server error', 'error');
  }
}

function renderTunnels(tunnels) {
  const container = document.getElementById('tunnels-table-container');
  
  if (tunnels.length === 0) {
    container.innerHTML = '<div class="empty-state">No tunnels found</div>';
    return;
  }
  
  const table = `
    <table class="table">
      <thead>
        <tr>
          <th>Name</th>
          <th>Domain</th>
          <th>Port</th>
          <th>Status</th>
          <th>CPU</th>
          <th>Memory</th>
          <th>Actions</th>
        </tr>
      </thead>
      <tbody>
        ${tunnels.map(tunnel => `
          <tr>
            <td>${tunnel.name}</td>
            <td>${tunnel.domain}</td>
            <td>${tunnel.port}</td>
            <td><span class="status-${tunnel.status}">${tunnel.status.toUpperCase()}</span></td>
            <td>${tunnel.cpu ? tunnel.cpu.toFixed(1) + '%' : 'N/A'}</td>
            <td>${tunnel.memory ? tunnel.memory.toFixed(1) + 'MB' : 'N/A'}</td>
            <td>
              ${tunnel.status === 'running' 
                ? `<button class="btn btn-danger btn-small" onclick="stopTunnel('${tunnel.name}')">Stop</button>`
                : `<button class="btn btn-success btn-small" onclick="startTunnel('${tunnel.name}')">Start</button>`
              }
              <button class="btn btn-danger btn-small" onclick="deleteTunnel('${tunnel.name}')">Delete</button>
            </td>
          </tr>
        `).join('')}
      </tbody>
    </table>
  `;
  
  container.innerHTML = table;
}

async function startTunnel(name) {
  try {
    const response = await fetch('/api/tunnels/' + name + '/start', {
      method: 'POST'
    });

    if (response.ok) {
      showToast('Tunnel started', 'success');
      fetchTunnels();
    } else {
      showToast('Failed to start tunnel', 'error');
    }
  } catch (error) {
    showToast('Server error', 'error');
  }
}

async function stopTunnel(name) {
  try {
    const response = await fetch('/api/tunnels/' + name + '/stop', {
      method: 'POST'
    });

    if (response.ok) {
      showToast('Tunnel stopped', 'success');
      fetchTunnels();
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
    const response = await fetch('/api/tunnels/' + name, {
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

// Utility Functions
function refreshAll() {
  fetchDNSRecords();
  fetchTunnels();
  showToast('Data refreshed', 'success');
}

function logout() {
  showToast('Logging out...', 'warning');
  setTimeout(() => {
    window.location.href = '/api/logout';
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

function showCreateDNSModal() {
  document.getElementById('dns-form').reset();
  openModal('dns-modal');
}

function showCreateTunnelModal() {
  document.getElementById('tunnel-form').reset();
  openModal('tunnel-modal');
}

function showChangePasswordModal() {
  document.getElementById('change-password-form').reset();
  openModal('change-password-modal');
} 