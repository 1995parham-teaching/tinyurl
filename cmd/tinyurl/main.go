package main

import (
	"github.com/1989michael/tinyurl/internal/cmd"
	"github.com/pterm/pterm"
	"github.com/pterm/pterm/putils"
)

func main() {
	pterm.DefaultCenter.Println("Shorten your URL to easily remember them and share them with your clients")

	s, _ := pterm.DefaultBigText.WithLetters(putils.LettersFromString("TinyURL")).Srender()
	pterm.DefaultCenter.Println(s)

	pterm.DefaultCenter.WithCenterEachLineSeparately().Println("Michael Weiss\nJuly 2023")

	cmd.Execute()
}
