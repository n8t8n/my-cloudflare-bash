package auth

import (
	"html/template"
	"net/http"
)

var LoginTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Cloudflare Manager Login</title>
<style>
* { box-sizing: border-box; margin: 0; padding: 0; }
body {
  background: #1a1a2e;
  color: #eee;
  font-family: 'Courier New', monospace;
  height: 100vh;
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  gap: 1rem;
}
.logo {
  color: #ff6b35;
  font-size: 2rem;
  font-weight: bold;
  margin-bottom: 1rem;
}
.prompt { color: #ff6b35; font-size: 1.2rem; }
.input {
  background: transparent;
  border: none;
  border-bottom: 2px solid #0f3460;
  color: #16213e;
  font-size: 1.2rem;
  padding: 0.5rem 0;
  outline: none;
  width: 300px;
  text-align: center;
  color: #eee;
}
.input:focus { border-bottom-color: #ff6b35; }
.message { color: #ff4757; min-height: 1.2rem; }
</style>
</head>
<body>
<div class="logo">☁️ CF-MANAGER</div>
<div class="prompt">admin@cloudflare:~$</div>
<input type="password" class="input" id="password" placeholder="enter password" autofocus>
<div class="message" id="message"></div>

<script>
document.getElementById('password').addEventListener('keydown', function(e) {
  if (e.key === 'Enter') {
    const password = this.value.trim();
    if (!password) return;

    document.getElementById('message').textContent = 'Authenticating...';

    fetch('/login', {
      method: 'POST',
      headers: {'Content-Type': 'application/json'},
      body: JSON.stringify({ password })
    })
    .then(res => res.json())
    .then(data => {
      if (data.success) {
        document.getElementById('message').textContent = 'Access granted.';
        window.location.href = '/dashboard';
      } else {
        document.getElementById('message').textContent = 'Access denied.';
        this.value = '';
        this.focus();
      }
    })
    .catch(() => {
      document.getElementById('message').textContent = 'Server error.';
      this.focus();
    });
  }
});
</script>
</body>
</html>`

var loginTemplate *template.Template

func init() {
	var err error
	loginTemplate, err = template.New("login").Parse(LoginTemplate)
	if err != nil {
		panic("Failed to parse login template: " + err.Error())
	}
}

func RenderLogin(w http.ResponseWriter, data interface{}) error {
	w.Header().Set("Content-Type", "text/html")
	return loginTemplate.Execute(w, data)
}
