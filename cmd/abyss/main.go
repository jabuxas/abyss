package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/jabuxas/abyss/internal/routing"
	"github.com/jabuxas/abyss/internal/utils"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
	builtBy = "unknown"
)

func main() {
	versionFlag := flag.Bool("v", false, "print version and build info")
	flag.Parse()

	if *versionFlag {
		fmt.Printf("abyss version: %s\n", version)
		fmt.Printf("git commit: %s\n", commit)
		fmt.Printf("build date: %s\n", date)
		fmt.Printf("built by: %s\n", builtBy)
		os.Exit(0)
	}

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
