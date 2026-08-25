package worker

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"amnezia-mikrotik-panel/internal/database"
	"amnezia-mikrotik-panel/internal/models"
	"amnezia-mikrotik-panel/internal/service"
)

// StartCron job that runs every 30 seconds
func StartCron() {
	ticker := time.NewTicker(30 * time.Second)
	go func() {
		for {
			<-ticker.C
			err := processBandwidthAndLimits()
			if err != nil {
				log.Printf("Cron error: %v", err)
			}
		}
	}()
	log.Println("Background worker (cron) started.")
}

func processBandwidthAndLimits() error {
	dump, err := service.ShowDump("awg0")
	if err != nil {
		return fmt.Errorf("failed to get awg dump: %v", err)
	}

	lines := strings.Split(dump, "\n")
	
	// Map of pubkey -> (rx + tx)
	currentUsage := make(map[string]int64)

	for _, line := range lines {
		fields := strings.Split(line, "\t")
		// peer lines have 8 fields
		if len(fields) >= 8 && len(fields[0]) == 44 { // 44 is base64 key length roughly
			pubKey := fields[0]
			rx, _ := strconv.ParseInt(fields[5], 10, 64)
			tx, _ := strconv.ParseInt(fields[6], 10, 64)
			currentUsage[pubKey] = rx + tx
		}
	}

	rows, err := database.DB.Query("SELECT id, public_key, status, data_limit_bytes, total_bytes, session_start_rxtx, expires_at FROM users WHERE status = 'active'")
	if err != nil {
		return fmt.Errorf("failed to query active users: %v", err)
	}
	defer rows.Close()

	var toUpdate []models.User

	for rows.Next() {
		var user models.User
		var expiresAt sql.NullTime
		if err := rows.Scan(&user.ID, &user.PublicKey, &user.Status, &user.DataLimitBytes, &user.TotalBytes, &user.SessionStartRxTx, &expiresAt); err != nil {
			log.Printf("Failed to scan user: %v", err)
			continue
		}
		if expiresAt.Valid {
			user.ExpiresAt = &expiresAt.Time
		}

		// Calculate new total
		currRxTx, ok := currentUsage[user.PublicKey]
		if !ok {
			// Peer not in dump (might not have connected yet), skip calculation
			continue
		}

		if currRxTx < user.SessionStartRxTx {
			// Reboot/interface reset happened
			user.SessionStartRxTx = 0
		}

		delta := currRxTx - user.SessionStartRxTx
		user.TotalBytes += delta
		user.SessionStartRxTx = currRxTx

		// Check Limits & Expiration
		statusChanged := false
		if user.DataLimitBytes > 0 && user.TotalBytes >= user.DataLimitBytes {
			user.Status = models.StatusLimited
			statusChanged = true
		} else if user.ExpiresAt != nil && time.Now().After(*user.ExpiresAt) {
			user.Status = models.StatusExpired
			statusChanged = true
		}

		toUpdate = append(toUpdate, user)

		if statusChanged {
			log.Printf("User %s (%s) exceeded limit. Disconnecting...", user.ID, user.PublicKey)
			if err := service.RemovePeer("awg0", user.PublicKey); err != nil {
				log.Printf("Failed to remove peer %s from interface: %v", user.PublicKey, err)
			}
		}
	}

	// Persist updates to DB
	tx, err := database.DB.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare("UPDATE users SET status = ?, total_bytes = ?, session_start_rxtx = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?")
	if err != nil {
		tx.Rollback()
		return err
	}
	defer stmt.Close()

	for _, u := range toUpdate {
		_, err := stmt.Exec(u.Status, u.TotalBytes, u.SessionStartRxTx, u.ID)
		if err != nil {
			log.Printf("Failed to update user %s in DB: %v", u.ID, err)
		}
	}

	return tx.Commit()
}
