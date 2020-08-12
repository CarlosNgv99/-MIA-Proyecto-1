package parser

import (
	"bufio"
	"log"
	"os"
)

// MKDISK folder (route) creation

func mkdiskAction(route string) {
	//Create a folder/directory at a full qualified path
	err := os.Mkdir(route, 0777)
	if err != nil {
		log.Fatal(err)
	}
}

// Pauses input

func pauseAction() {
	bufio.NewReader(os.Stdin).ReadBytes('\n')
}
