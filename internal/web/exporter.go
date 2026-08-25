package web

import (
	"bytes"
	"net/http"
	"strings"
	"text/template"

	"amnezia-mikrotik-panel/internal/config"
	"amnezia-mikrotik-panel/internal/database"
	"amnezia-mikrotik-panel/internal/service"

	"github.com/skip2/go-qrcode"
)

const confTemplate = `[Interface]
PrivateKey = {{ .PrivateKey }}
Address = {{ .Address }}
DNS = 1.1.1.1, 1.0.0.1
Jc = {{ .Jc }}
Jmin = {{ .Jmin }}
Jmax = {{ .Jmax }}
S1 = {{ .S1 }}
S2 = {{ .S2 }}
S3 = {{ .S3 }}
S4 = {{ .S4 }}
H1 = {{ .H1 }}
H2 = {{ .H2 }}
H3 = {{ .H3 }}
H4 = {{ .H4 }}

[Peer]
PublicKey = {{ .ServerPublicKey }}
Endpoint = {{ .Endpoint }}
AllowedIPs = 0.0.0.0/0, ::/0
PersistentKeepalive = 25
`

func buildClientConfig(userID string) (string, error) {
	var privKey, allowedIPs string
	var jc, jmin, jmax, s1, s2, s3, s4 int
	var h1, h2, h3, h4 uint32

	err := database.DB.QueryRow(`
		SELECT preshared_key, allowed_ips, jc, jmin, jmax, s1, s2, s3, s4, h1, h2, h3, h4
		FROM users WHERE id = ?
	`, userID).Scan(
		&privKey, &allowedIPs, &jc, &jmin, &jmax, &s1, &s2, &s3, &s4, &h1, &h2, &h3, &h4,
	)
	if err != nil {
		return "", err
	}

	serverPubKey, err := service.GetServerPublicKey("awg0")
	if err != nil {
		return "", err
	}

	cfg := config.Load()

	tmpl, err := template.New("conf").Parse(confTemplate)
	if err != nil {
		return "", err
	}

	data := map[string]interface{}{
		"PrivateKey":      privKey,
		"Address":         allowedIPs,
		"Jc":              jc,
		"Jmin":            jmin,
		"Jmax":            jmax,
		"S1":              s1,
		"S2":              s2,
		"S3":              s3,
		"S4":              s4,
		"H1":              h1,
		"H2":              h2,
		"H3":              h3,
		"H4":              h4,
		"ServerPublicKey": serverPubKey,
		"Endpoint":        cfg.ServerEndpoint,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func handleExportConf(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/users/export/")
	id = strings.TrimSuffix(id, "/conf")

	conf, err := buildClientConfig(id)
	if err != nil {
		http.Error(w, "Failed to generate config", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Disposition", "attachment; filename=peer_"+id+".conf")
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(conf))
}

func handleExportQR(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/users/export/")
	id = strings.TrimSuffix(id, "/qr")

	conf, err := buildClientConfig(id)
	if err != nil {
		http.Error(w, "Failed to generate config", http.StatusInternalServerError)
		return
	}

	png, err := qrcode.Encode(conf, qrcode.Medium, 256)
	if err != nil {
		http.Error(w, "Failed to generate QR code", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "image/png")
	w.Write(png)
}
