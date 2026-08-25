package web

import (
	"io"
	"net/http"
	"os"

	"amnezia-mikrotik-panel/internal/config"
	"amnezia-mikrotik-panel/internal/database"
)

func handleBackup(w http.ResponseWriter, r *http.Request) {
	cfg := config.Load()

	file, err := os.Open(cfg.DBPath)
	if err != nil {
		http.Error(w, "Failed to open database file", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	w.Header().Set("Content-Disposition", "attachment; filename=panel_backup.db")
	w.Header().Set("Content-Type", "application/octet-stream")
	io.Copy(w, file)
}

func handleRestore(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseMultipartForm(10 << 20) // 10 MB limit
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	file, _, err := r.FormFile("backup")
	if err != nil {
		http.Error(w, "Failed to get file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	cfg := config.Load()

	// Close existing DB connection
	if database.DB != nil {
		database.DB.Close()
	}

	// Overwrite file
	out, err := os.Create(cfg.DBPath)
	if err != nil {
		http.Error(w, "Failed to write database file", http.StatusInternalServerError)
		return
	}
	defer out.Close()

	if _, err := io.Copy(out, file); err != nil {
		http.Error(w, "Failed to save database file", http.StatusInternalServerError)
		return
	}

	// Re-init DB
	if err := database.InitDB(cfg); err != nil {
		http.Error(w, "Failed to reload database", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
