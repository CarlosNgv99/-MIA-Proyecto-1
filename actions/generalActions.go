package actions

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
)

// PauseAction exported
func PauseAction() {
	fmt.Println("Press any key to continue.")
	bufio.NewReader(os.Stdin).ReadBytes('\n')
}

// GetFile exported
func GetFile(route string) { // Gets file from route
	re := regexp.MustCompile(`[a-zA-Z]([a-zA-Z]|[0-9])*\.mia`)
	file := re.FindString(route)
	fmt.Println(file)
}

// PrintParameter exported
func PrintParameter(parameter string) {
	fmt.Println("Parameter:", parameter)
}
