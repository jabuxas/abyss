package main

import (
	"github.com/jabuxas/abyss/internal/routing"
)

func main() {
	router := routing.GetRouter()
	router.Run(":3235")
}
