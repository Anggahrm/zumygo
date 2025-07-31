package main

import (
	"fmt"
	"github.com/subosito/gotenv"
)

func main() {
	gotenv.Load()

	fmt.Println("Starting WhatsApp bot...")
	StartClient()
}
