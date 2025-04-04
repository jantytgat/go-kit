package app

import "fmt"

func New() *App {
	return &App{}
}

type App struct{}

func (a *App) Execute() {
	fmt.Println("Hello World")
}
