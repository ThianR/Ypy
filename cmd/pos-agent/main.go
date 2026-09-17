package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"time"

	databaseplatform "github.com/ypy-erp/ypy/internal/platform/database"
	"github.com/ypy-erp/ypy/internal/poslocal"
)

func main() {
	path := os.Getenv("YPY_POS_DB")
	if path == "" {
		path = "ypy-pos.db"
	}
	db, err := poslocal.OpenStore(path)
	if err != nil {
		log.Fatal(err)
	}
	defer databaseplatform.Close(db)
	if destination := os.Getenv("YPY_POS_BACKUP"); destination != "" {
		if err := poslocal.BackupDatabase(db, destination); err != nil {
			log.Fatalf("POS backup failed: %v", err)
		}
		fmt.Printf("Ypy POS backup created: %s\n", destination)
		return
	}
	fmt.Printf("Ypy POS agent ready (SQLite: %s)\n", path)
	if central := os.Getenv("YPY_POS_CENTRAL_URL"); central != "" {
		company, terminal := os.Getenv("YPY_POS_EMPRESA_ID"), os.Getenv("YPY_POS_TERMINAL_ID")
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
		defer stop()
		interval := 30 * time.Second
		if raw := os.Getenv("YPY_POS_SYNC_INTERVAL_SECONDS"); raw != "" {
			if seconds, parseErr := strconv.Atoi(raw); parseErr == nil && seconds > 0 && seconds <= 3600 {
				interval = time.Duration(seconds) * time.Second
			}
		}
		sync := func() {
			attemptCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
			defer cancel()
			if sent, err := poslocal.SyncPending(attemptCtx, db, central, company, terminal, nil); err != nil {
				log.Printf("POS sync pending: %v", err)
			} else if sent > 0 {
				log.Printf("POS sync applied: %d", sent)
			}
		}
		sync()
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				log.Printf("Ypy POS agent stopping")
				return
			case <-ticker.C:
				sync()
			}
		}
	}
}
