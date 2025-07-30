package main

import (
	conn "zumygo"

	"github.com/subosito/gotenv"
)

func main() {
	gotenv.Load()

	conn.StartClient()
}
