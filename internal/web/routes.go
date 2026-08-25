package web

import (
	"html/template"
	"net/http"
	"strings"
)

// indexTmpl is parsed once at startup to avoid repeated disk I/O on each request.
var indexTmpl *template.Template

func init() {
	var err error
	indexTmpl, err = template.ParseFiles("templates/index.html")
	if err != nil {
		// Panic early so the error is visible at startup, not silently on first request.
		panic("failed to parse templates/index.html: " + err.Error())
	}
}

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
		indexTmpl.Execute(w, nil)
	})
}

