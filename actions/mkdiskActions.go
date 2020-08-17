package actions

import (
	"fmt"
	"log"
	"os"
	"strconv"
)

// MkdiskCreateRoute exported
func MkdiskCreateRoute(sizeDigit string, route string, name string) {
	i, _ := strconv.Atoi(sizeDigit)
	fmt.Println(i)
	err := os.Mkdir(route, 0777)
	if err != nil {
		createDisk(i, route, name)
	}
	createDisk(i, route, name)
}

func createDisk(sizeDigit int, route string, name string) {
	size := int64(sizeDigit * 1024 * 1024)
	fd, err := os.Create(route + name)
	if err != nil {
		log.Fatal("Failed to create output")
	}
	_, err = fd.Seek(size-1, 0)
	if err != nil {
		log.Fatal("Failed to seek")
	}
	_, err = fd.Write([]byte{0})
	if err != nil {
		log.Fatal("Write failed")
	}
	err = fd.Close()
	if err != nil {
		log.Fatal("Failed to close file")
	}
}
