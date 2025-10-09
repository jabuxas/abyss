package main

import (
	"time"

	"github.com/jabuxas/abyss/internal/routing"
	"github.com/jabuxas/abyss/internal/utils"
)

func main() {
	router := routing.GetRouter()

	go func() {
		ticker := time.NewTicker(30 * time.Minute)
		defer ticker.Stop()

		utils.CleanupExpiredFiles(routing.CFG.FilesDir)

		for range ticker.C {
			utils.CleanupExpiredFiles(routing.CFG.FilesDir)
		}
	}()

	router.Run(":3235")
}
