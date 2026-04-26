package main

import (
	"ticketer/internal/app"

	"go.uber.org/fx"
)

func main() {
	fx.New(app.Module).Run()
}
