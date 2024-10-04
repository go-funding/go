package printer

import (
	"fmt"
	"github.com/fatih/color"
	"golang.org/x/term"
	"os"
	"strings"
)

var logo = `
  ▄████  ▒█████  
 ██▒ ▀█▒▒██▒  ██▒
▒██░▄▄▄░▒██░  ██▒
░▓█  ██▓▒██   ██░
░▒▓███▀▒░ ████▓▒░
 ░▒   ▒ ░ ▒░▒░▒░ 
  ░   ░   ░ ▒ ▒░ 
░ ░   ░ ░ ░ ░ ▒  
      ░     ░ ░`

var title = `..-. ..- -. -.. .. -. --. ... / -.-. ..- .-. ... . -..`
var padding = 30

func PrintLogo() {
	if !term.IsTerminal(int(os.Stdout.Fd())) {
		return
	}

	var maxLogoLineLenght int

	// Center logo
	fmt.Print(strings.Repeat("\n", 1))
	lines := strings.Split(logo, "\n")
	for _, line := range lines {
		fmt.Print(strings.Repeat(" ", padding))
		fmt.Print(color.HiRedString(line))
		fmt.Print("\n")

		if len(line) > maxLogoLineLenght {
			maxLogoLineLenght = len(line)
		}
	}
	fmt.Print(strings.Repeat("\n", 1))

	// Center title
	var center = (padding + maxLogoLineLenght) / 2
	fmt.Print(strings.Repeat(" ", center-len(title)/2))
	fmt.Print(color.HiBlueString(title), "\n\n")
}

func PadStringsToLength(length int, strs ...string) string {
	var result string
	for _, str := range strs {
		result += str + strings.Repeat(" ", length-len(str))
	}
	return result
}
