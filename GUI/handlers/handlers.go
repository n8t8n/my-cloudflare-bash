package handlers

import (
	"encoding/json"
	"net/http"

	"cf-manager/auth"
	"cf-manager/dns"
	"cf-manager/templates"
	"cf-manager/tunnels"

	"github.com/gorilla/mux"
)

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var req auth.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	success := auth.ValidatePassword(req.Password)
	if success {
		// Pass the request to SetSession so it can determine the appropriate cookie settings
		auth.SetSession(w, r)
	}

	resp := auth.LoginResponse{Success: success}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	auth.ClearSession(w)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	auth.RenderLogin(w, nil)
}

func DashboardHandler(w http.ResponseWriter, r *http.Request) {
	data := &templates.TemplateData{
		Title: "Cloudflare Manager",
	}
	templates.RenderDashboard(w, data)
}

// DNS Handlers

func ListDNSRecordsHandler(w http.ResponseWriter, r *http.Request) {
	records, err := dns.ListDNSRecords()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(records)
}

func CreateDNSRecordHandler(w http.ResponseWriter, r *http.Request) {
	var req dns.CreateDNSRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	record, err := dns.CreateDNSRecord(req)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(record)
}

func DeleteDNSRecordHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	recordID := vars["id"]

	if err := dns.DeleteDNSRecord(recordID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

func UpdateDNSRecordHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	recordID := vars["id"]

	var req dns.UpdateDNSRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	record, err := dns.UpdateDNSRecord(recordID, req)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(record)
}

// Tunnel Handlers

func ListTunnelsHandler(w http.ResponseWriter, r *http.Request) {
	tunnelList, err := tunnels.ListTunnels()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tunnelList)
}

func CreateTunnelHandler(w http.ResponseWriter, r *http.Request) {
	var req tunnels.CreateTunnelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	tunnel, err := tunnels.CreateTunnel(req)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tunnel)
}

func DeleteTunnelHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	if err := tunnels.DeleteTunnel(name); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

func StartTunnelHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	if err := tunnels.StartTunnel(name); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

func StopTunnelHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	if err := tunnels.StopTunnel(name); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

func GetTunnelStatusHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	tunnel, err := tunnels.GetTunnelStatus(name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tunnel)
}

func EditTunnelConfigHandler(w http.ResponseWriter, r *http.Request) {
	// This would be implemented for editing tunnel configurations
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Config editing not implemented yet"})
}

func SystemStatusHandler(w http.ResponseWriter, r *http.Request) {
	// Basic system status
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "operational",
		"uptime": "running",
	})
}

func ChangePasswordHandler(w http.ResponseWriter, r *http.Request) {
	var req auth.ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	success := auth.ChangePassword(req.OldPassword, req.NewPassword)
	response := auth.ChangePasswordResponse{Success: success}

	w.Header().Set("Content-Type", "application/json")
	if success {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusUnauthorized)
	}
	json.NewEncoder(w).Encode(response)
}
