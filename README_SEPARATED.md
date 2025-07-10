# Cloudflare Manager - Separated Architecture

This is the **separated version** of the Cloudflare Manager, where the frontend and backend are cleanly separated for better maintainability and development experience.

## 🏗️ Architecture Overview

### **Separated Structure**
```
├── backend/                 # Go API server
│   ├── main.go             # API server + static file server
│   ├── handlers/           # API handlers
│   ├── auth/              # Authentication logic
│   ├── dns/               # DNS management
│   ├── tunnels/           # Tunnel management
│   ├── middleware/        # HTTP middleware
│   ├── start.sh           # Startup script
│   └── go.mod
├── static/                 # Static frontend files
│   ├── index.html         # Main dashboard
│   ├── login.html         # Login page
│   ├── css/
│   │   └── style.css      # All styles
│   └── js/
│       └── dashboard.js   # Dashboard functionality
└── GUI/                   # Original monolithic version (kept for reference)
```

## 🚀 Benefits of Separation

### **1. Development Benefits**
- **Easier Frontend Development**: HTML/CSS/JS files can be edited directly
- **Better IDE Support**: Syntax highlighting, autocomplete for frontend files
- **Faster Iteration**: No need to rebuild Go binary for frontend changes
- **Version Control**: Better diff tracking for frontend changes

### **2. Maintenance Benefits**
- **Clear Separation**: Frontend and backend concerns are separated
- **Easier Debugging**: Can debug frontend and backend independently
- **Better Organization**: Clear file structure and responsibilities

### **3. Deployment Benefits**
- **Single Binary**: Still deployed as one binary (no CORS issues)
- **Static File Serving**: Efficient static file serving by Go
- **No Build Process**: Frontend doesn't need compilation

## 📊 Comparison with Monolithic Version

| Aspect | Monolithic | Separated |
|--------|------------|-----------|
| **Frontend Location** | Embedded in Go templates | Static HTML/CSS/JS files |
| **Frontend Editing** | Requires Go rebuild | Direct file editing |
| **File Size** | 990 lines in one file | Split into logical files |
| **Development Speed** | Slower (rebuild needed) | Faster (direct editing) |
| **Deployment** | Single binary | Single binary |
| **CORS** | None (same origin) | None (same origin) |

## 🔧 API Endpoints

All API endpoints are prefixed with `/api`:

### **Authentication**
- `POST /api/login` - Login
- `GET /api/logout` - Logout
- `POST /api/change-password` - Change password

### **DNS Management**
- `GET /api/dns/records` - List DNS records
- `POST /api/dns/records` - Create DNS record
- `PUT /api/dns/records/{id}` - Update DNS record
- `DELETE /api/dns/records/{id}` - Delete DNS record

### **Tunnel Management**
- `GET /api/tunnels` - List tunnels
- `POST /api/tunnels` - Create tunnel
- `DELETE /api/tunnels/{name}` - Delete tunnel
- `POST /api/tunnels/{name}/start` - Start tunnel
- `POST /api/tunnels/{name}/stop` - Stop tunnel
- `GET /api/tunnels/{name}/status` - Get tunnel status

### **System**
- `GET /api/system/status` - System status

## 🚀 Quick Start

### **1. Setup Environment**
```bash
# Copy environment variables
cp .env backend/

# Set required variables in backend/.env
CF_API_TOKEN=your_api_token
CF_ZONE_ID=your_zone_id
CF_DOMAIN=your_domain.com
```

### **2. Start the Application**
```bash
# Navigate to backend directory
cd backend

# Start the application
./start.sh
```

### **3. Access the Application**
- **Frontend**: http://localhost:3000
- **Login**: http://localhost:3000/login.html
- **Default Password**: `admin`

## 🛠️ Development

### **Frontend Development**
```bash
# Edit frontend files directly
vim static/css/style.css
vim static/js/dashboard.js
vim static/index.html
```

### **Backend Development**
```bash
# Edit backend files
vim backend/handlers/handlers.go
vim backend/tunnels/tunnels.go

# Rebuild and restart
cd backend && ./start.sh
```

### **Adding New Features**

#### **Frontend Changes**
1. Edit the appropriate file in `static/`
2. Refresh browser to see changes
3. No rebuild needed

#### **Backend Changes**
1. Edit Go files in `backend/`
2. Run `cd backend && ./start.sh` to rebuild and restart
3. Frontend will automatically use new API endpoints

## 📁 File Structure Details

### **Backend (`backend/`)**
- **`main.go`**: HTTP server, routes, static file serving
- **`handlers/`**: API endpoint handlers
- **`auth/`**: Authentication and session management
- **`dns/`**: DNS record management
- **`tunnels/`**: Tunnel management
- **`middleware/`**: HTTP middleware (CORS, auth, rate limiting)

### **Frontend (`static/`)**
- **`index.html`**: Main dashboard page
- **`login.html`**: Login page
- **`css/style.css`**: All styles (extracted from templates)
- **`js/dashboard.js`**: All JavaScript functionality

## 🔄 Migration from Monolithic

The separated version maintains **100% compatibility** with the monolithic version:

1. **Same Features**: All functionality preserved
2. **Same API**: All endpoints work identically
3. **Same Authentication**: Session management unchanged
4. **Same Deployment**: Single binary deployment

## 🎯 Advantages Over Bash Scripts

### **Resource Usage Comparison**
| Component | Bash Scripts | Separated GUI |
|-----------|--------------|---------------|
| **Storage** | 28KB | 120KB |
| **Binary Size** | N/A | 8MB (optimized) |
| **Dependencies** | Shell + cloudflared | Go + 8 dependencies |
| **Features** | Basic CLI | Full web interface |
| **Usability** | Terminal only | Web browser |

### **Feature Comparison**
| Feature | Bash Scripts | Separated GUI |
|---------|--------------|---------------|
| **DNS Management** | ✅ Basic | ✅ Full CRUD |
| **Tunnel Management** | ✅ Basic | ✅ Full CRUD |
| **Real-time Status** | ❌ Manual refresh | ✅ Auto-refresh |
| **User Interface** | ❌ CLI only | ✅ Modern web UI |
| **Authentication** | ❌ None | ✅ Secure login |
| **Multi-user** | ❌ No | ✅ Session-based |

## 🚀 Future Enhancements

With the separated architecture, it's easier to add:

1. **Frontend Framework**: Could migrate to React/Vue/Angular
2. **Build Process**: Could add webpack/vite for optimization
3. **PWA Features**: Service workers, offline support
4. **Real-time Updates**: WebSocket integration
5. **Mobile App**: Could create mobile frontend

## 📝 Notes

- **Backward Compatibility**: The separated version is a drop-in replacement
- **No Breaking Changes**: All existing functionality preserved
- **Easy Rollback**: Can always go back to monolithic version
- **Performance**: Same performance as monolithic version
- **Security**: Same security model as monolithic version 