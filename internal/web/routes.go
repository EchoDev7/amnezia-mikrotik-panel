package web

import (
	"html/template"
	"net/http"
	"strings"
)

func RegisterRoutes() {
	// API Routes
	http.HandleFunc("/api/stats", handleGetStats)
	http.HandleFunc("/api/users", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			handleGetUsers(w, r)
		} else if r.Method == http.MethodPost {
			handleCreateUser(w, r)
		}
	})
	http.HandleFunc("/api/users/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/users/export/") {
			if strings.HasSuffix(r.URL.Path, "/conf") {
				handleExportConf(w, r)
			} else if strings.HasSuffix(r.URL.Path, "/qr") {
				handleExportQR(w, r)
			}
			return
		}
		if strings.HasPrefix(r.URL.Path, "/api/users/toggle/") {
			if r.Method == http.MethodPost {
				handleToggleUser(w, r)
			}
			return
		}
		if r.Method == http.MethodDelete {
			handleDeleteUser(w, r)
		}
	})

	// Backup / Restore
	http.HandleFunc("/api/backup", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			handleBackup(w, r)
		}
	})
	http.HandleFunc("/api/restore", handleRestore)

	// Frontend Dashboard
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		tmpl, err := template.ParseFiles("templates/index.html")
		if err != nil {
			http.Error(w, "Failed to load template", http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, nil)
	})
}
