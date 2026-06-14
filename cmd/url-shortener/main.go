package main

import (
	"log"

	"ozon-intern/internal/app"
)

func main() {
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
