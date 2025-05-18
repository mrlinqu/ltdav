package main

import (
	"fmt"
	"os"

	"github.com/mrlinqu/ltdav/internal/app/htpasswd"
	"github.com/mrlinqu/ltdav/internal/app/htpasswd/config"
)

func main() {
	cfg, err := config.New(os.Args[1:])
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	app := htpasswd.New()

	if err := app.Run(cfg); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

}
