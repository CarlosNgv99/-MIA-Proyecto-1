package actions

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
)

// MKDISK

// Disk exported
type Disk struct {
	Size  int
	Route string
	Name  string
	Unit  string
}

// CreateDisk exported
func (d *Disk) CreateDisk() {
	if d.Name == "" {
		fmt.Println("Disk name is missing. Try again.")
	} else if d.Route == "" {
		fmt.Println("Disk path is missing. Try again.")
	} else if d.Size == 0 {
		fmt.Println("Size path is missing. Try again.")
	} else {
		err := os.Mkdir(d.Route, 0777)
		if err != nil {
			d.setDisk()
		}
		d.setDisk()
	}
}

// ShowDisk exported
func (d *Disk) ShowDisk() {
	fmt.Println(d)
}

// SetDiskName exported
func (d *Disk) SetDiskName(name string) {
	d.Name = name
	fmt.Println("1")
}

// SetDiskUnit exported
func (d *Disk) SetDiskUnit(unit string) {
	d.Unit = unit
}

// SetDiskRoute exported
func (d *Disk) SetDiskRoute(route string) {
	d.Route = route
}

// SetDiskSize exported
func (d *Disk) SetDiskSize(size string) {
	i, _ := strconv.Atoi(size)
	d.Size = i
}

/*
// MkdiskCreateRoute exported
func MkdiskCreateRoute(sizeDigit string, route string, name string, unit string) {

	i, _ := strconv.Atoi(sizeDigit) // int converted to string
	err := os.Mkdir(route, 0777)
	if err != nil {
		createDisk(i, route, name, unit)
	}
	createDisk(i, route, name, unit)
}*/

func (d *Disk) setDisk() {

	size := SetUnit(d.Unit, d.Size)

	path := d.Route + d.Name

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

// RemoveDisk exported
func RemoveDisk(path string) {
	re := regexp.MustCompile(`[a-zA-Z]([a-zA-Z]|[0-9])*\.dsk`)
	file := re.FindString(path)
	err := os.Remove(path)
	if err != nil {
		fmt.Println("File does not exist.")
	} else {
		fmt.Println("File: " + file + " successfully removed.")
	}
}

// FDISK

// SetUnit exported
func SetUnit(unit string, sizeU int) int64 {

	var size int64

	switch unit {
	case "m":
		size = int64(sizeU * 1024 * 1024)
	case "k":
		size = int64(sizeU * 1024)
	default:
		size = int64(sizeU * 1024 * 1024)
	}

	return size
}
