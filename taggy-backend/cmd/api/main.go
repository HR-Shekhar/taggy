package main

import (
	"log"

	"github.com/HR-Shekhar/taggy-backend/internal/app"
)

func main() {
	
	a, err := app.New()
	if err != nil {
		log.Fatal(err)
	}

	a.Run()


}
