package web

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"amnezia-mikrotik-panel/internal/database"
	"amnezia-mikrotik-panel/internal/models"
	"amnezia-mikrotik-panel/internal/service"

	"github.com/google/uuid"
)

type StatsResponse struct {
	TotalUsers   int   `json:"total_users"`
	ActiveUsers  int   `json:"active_users"`
	ExpiredUsers int   `json:"expired_users"`
	LimitedUsers int   `json:"limited_users"`
	TotalBytes   int64 `json:"total_bytes"`
}

func handleGetStats(w http.ResponseWriter, r *http.Request) {
	var stats StatsResponse

	err := database.DB.QueryRow(`
		SELECT 
			COUNT(*), 
			SUM(CASE WHEN status = 'active' THEN 1 ELSE 0 END),
			SUM(CASE WHEN status = 'expired' THEN 1 ELSE 0 END),
			SUM(CASE WHEN status = 'limited' THEN 1 ELSE 0 END),
			SUM(total_bytes)
		FROM users
	`).Scan(&stats.TotalUsers, &stats.ActiveUsers, &stats.ExpiredUsers, &stats.LimitedUsers, &stats.TotalBytes)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func handleGetUsers(w http.ResponseWriter, r *http.Request) {
	rows, err := database.DB.Query("SELECT id, name, public_key, allowed_ips, status, data_limit_bytes, total_bytes, expires_at, created_at, updated_at FROM users")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		var expiresAt sql.NullTime
		if err := rows.Scan(&u.ID, &u.Name, &u.PublicKey, &u.AllowedIPs, &u.Status, &u.DataLimitBytes, &u.TotalBytes, &expiresAt, &u.CreatedAt, &u.UpdatedAt); err != nil {
			log.Printf("Scan error: %v", err)
			continue
		}
		if expiresAt.Valid {
			u.ExpiresAt = &expiresAt.Time
		}
		users = append(users, u)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

type CreateUserRequest struct {
	Name           string `json:"name"`
	AllowedIPs     string `json:"allowed_ips"`
	DataLimitBytes int64  `json:"data_limit_bytes"`
	ExpiresAt      string `json:"expires_at"` // RFC3339

	// Obfuscation params
	Jc   int    `json:"jc"`
	Jmin int    `json:"jmin"`
	Jmax int    `json:"jmax"`
	S1   int    `json:"s1"`
	S2   int    `json:"s2"`
	S3   int    `json:"s3"`
	S4   int    `json:"s4"`
	H1   uint32 `json:"h1"`
	H2   uint32 `json:"h2"`
	H3   uint32 `json:"h3"`
	H4   uint32 `json:"h4"`
}

func handleCreateUser(w http.ResponseWriter, r *http.Request) {
	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	privKey, err := service.GeneratePrivateKey()
	if err != nil {
		http.Error(w, "Failed to generate private key", http.StatusInternalServerError)
		return
	}

	pubKey, err := service.GeneratePublicKey(privKey)
	if err != nil {
		http.Error(w, "Failed to generate public key", http.StatusInternalServerError)
		return
	}

	var expiresAt *time.Time
	if req.ExpiresAt != "" {
		t, err := time.Parse(time.RFC3339, req.ExpiresAt)
		if err == nil {
			expiresAt = &t
		}
	}

	id := uuid.New().String()

	_, err = database.DB.Exec(`
		INSERT INTO users (
			id, name, public_key, preshared_key, allowed_ips, status, data_limit_bytes, 
			expires_at, jc, jmin, jmax, s1, s2, s3, s4, h1, h2, h3, h4
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, id, req.Name, pubKey, privKey, req.AllowedIPs, models.StatusActive, req.DataLimitBytes,
		expiresAt, req.Jc, req.Jmin, req.Jmax, req.S1, req.S2, req.S3, req.S4, req.H1, req.H2, req.H3, req.H4)

	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Add peer to interface
	if err := service.AddPeer("awg0", pubKey, req.AllowedIPs); err != nil {
		log.Printf("Failed to add peer to interface, but it's in DB: %v", err)
	}

	w.WriteHeader(http.StatusCreated)
}

func handleDeleteUser(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/users/")
	
	var pubKey string
	err := database.DB.QueryRow("SELECT public_key FROM users WHERE id = ?", id).Scan(&pubKey)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Remove from interface
	service.RemovePeer("awg0", pubKey)

	// Remove from DB
	_, err = database.DB.Exec("DELETE FROM users WHERE id = ?", id)
	if err != nil {
		http.Error(w, "Failed to delete user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func handleToggleUser(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/users/toggle/")
	
	var pubKey, status, allowedIPs string
	err := database.DB.QueryRow("SELECT public_key, status, allowed_ips FROM users WHERE id = ?", id).Scan(&pubKey, &status, &allowedIPs)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	newStatus := models.StatusDisabled
	if status == string(models.StatusDisabled) || status == string(models.StatusExpired) || status == string(models.StatusLimited) {
		newStatus = models.StatusActive
		// Re-add peer to interface
		service.AddPeer("awg0", pubKey, allowedIPs)
	} else {
		// Remove peer
		service.RemovePeer("awg0", pubKey)
	}

	_, err = database.DB.Exec("UPDATE users SET status = ? WHERE id = ?", newStatus, id)
	if err != nil {
		http.Error(w, "Failed to update user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
