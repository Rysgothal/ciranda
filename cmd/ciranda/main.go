package main

import (
	"ciranda/internal/cli"
	"ciranda/internal/http"
	"fmt"
)

func main() {
	cli.Menu()
	if err := http.Init(); err != nil {
		fmt.Println("An error occurred: " + err.Error())
	}
}
