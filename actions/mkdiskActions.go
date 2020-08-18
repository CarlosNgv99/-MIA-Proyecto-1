package actions

import (
	"log"
	"os"
	"strconv"
)

// MkdiskCreateRoute exported
func MkdiskCreateRoute(sizeDigit string, route string, name string, unit string) {

	i, _ := strconv.Atoi(sizeDigit) // int converted to string
	err := os.Mkdir(route, 0777)
	if err != nil {
		createDisk(i, route, name, unit)
	}
	createDisk(i, route, name, unit)
}

func createDisk(sizeDigit int, route string, name string, unit string) {

	var size int64

	switch unit {
	case "m":
		size = int64(sizeDigit * 1024 * 1024)
	case "k":
		size = int64(sizeDigit * 1024)
	default:
		size = int64(sizeDigit * 1024 * 1024)
	}

	path := route + name

	fd, err := os.Create(path)
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
