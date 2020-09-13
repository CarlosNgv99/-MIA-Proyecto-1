package parser

import "fmt"

type Disk struct {
	size int
	path string
	name string
	unit string
}

// CreateDisk exported
func CreateDisk() Disk {
	newDisk := Disk{}

	return newDisk
}

func (d *Disk) setName(name string) {
	d.name = name
	fmt.Println(d.name)
}
