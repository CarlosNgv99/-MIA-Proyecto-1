package actions

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"math/rand"
	"os"
	"regexp"
	"strconv"
	"time"
	"unsafe"
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
		fmt.Println(">> Disk name is missing. Try again.")
	} else if d.Route == "" {
		fmt.Println("Disk path is missing. Try again.")
	} else if d.Size == 0 {
		fmt.Println(">> Disk size is missing. Try again.")
	} else {
		err := os.Mkdir(d.Route, 0777)
		if err != nil {
			d.setDisk()
		} else {
			d.setDisk()
		}

	}
}

// SetDiskName exported
func (d *Disk) SetDiskName(name string) {
	d.Name = name
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

func (d *Disk) setDisk() {

	size := SetUnit(d.Unit, d.Size)
	path := d.Route + d.Name

	fd, err := os.Create(path)
	defer fd.Close()
	if err != nil {
		log.Fatal(">> Failed to create output")
	}
	_, err = fd.Seek(size-1, 0)
	if err != nil {
		log.Fatal(">> Failed to seek")
	}
	_, err = fd.Write([]byte{0})
	if err != nil {
		log.Fatal(">> Write failed")
	}

	newMbr := MBR{}

	// Date
	date := time.Now()
	formattedDate := date.Format("2006-01-02 15:04:05")
	copy(newMbr.Date[:], formattedDate)

	// Size
	newMbr.Size = size

	// Signature
	randNum := randomNumber()
	newMbr.Signature = randNum

	// Initializing partitions
	for i := 0; i < 4; i++ {
		newMbr.Partitions[i].Status = 'F'
		newMbr.Partitions[i].Start = -1
	}

	mbr := &newMbr

	// File writing process
	fd.Seek(0, 0)
	var buffer bytes.Buffer

	er := binary.Write(&buffer, binary.BigEndian, mbr)
	if er != nil {
		fmt.Println(">> Error writing file.")
	}
	writeMBR(fd, buffer.Bytes())
	readFile(d.Route, d.Name)
}

func writeMBR(file *os.File, bytes []byte) {
	_, err := file.Write(bytes)
	if err != nil {
		fmt.Println(">> Error writing file. Try again.")
	}
}

func readFile(route string, name string) {
	file, err := os.Open(route + name)
	if err != nil {
		fmt.Println(">> Error reading the file. Try again.")
	}

	mbr := MBR{}

	size := int(unsafe.Sizeof(mbr))
	data := readBytes(file, size)
	buff := bytes.NewBuffer(data)
	_ = binary.Read(buff, binary.BigEndian, &mbr)
	date := string(mbr.Date[:])
	fmt.Println("DISK SIZE:", mbr.Size, "bytes")
	fmt.Println("MBR SIZE:", binary.Size(mbr), "bytes") // Binary.size does not reads structs with slices, use unsafe.sizeof instead
	fmt.Println("CREATED AT:", date)
	fmt.Println("SIGNATURE", (mbr.Signature))
	for i := 0; i < 4; i++ {
		status := mbr.Partitions[i].Status
		fmt.Println("Partition ", i, " Status: ", string(status))
	}

}

func readBytes(file *os.File, size int) []byte {
	bytes := make([]byte, size)
	_, err := file.Read(bytes)
	if err != nil {
		fmt.Println(">> Error reading the file. Try again.")
	}
	return bytes
}

func randomNumber() int64 {
	min := 1
	max := 200
	number := rand.Intn(max-min) + min
	return int64(number)
}

// RemoveDisk exported
func RemoveDisk(path string) {
	re := regexp.MustCompile(`[a-zA-Z]([a-zA-Z]|[0-9])*\.dsk`)
	file := re.FindString(path)
	err := os.Remove(path)
	fmt.Println(file)
	if err != nil {
		fmt.Println(">> File does not exist.")
	} else {
		fmt.Println(">> File: " + file + " successfully removed.")
	}
}

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

// FDISK

// FDisk exported
type FDisk struct {
}

func setPartitionUnit(unit string, sizeU int) int64 {
	var size int64

	switch unit {
	case "m": // Megabytes
		size = int64(sizeU * 1024 * 1024)
	case "k": // Kilobytes
		size = int64(sizeU * 1024)
	case "b": // Bytes
		size = int64(sizeU)
	default: // Kilobytes
		size = int64(sizeU * 1024)
	}

	return size
}

// MBR exported
type MBR struct {
	Size       int64
	Date       [20]byte
	Signature  int64
	Partitions [4]Partition
}

// Partition exported
type Partition struct {
	Status byte
	Type   byte
	Fit    byte
	Start  int64
	Size   int64
	Name   [16]byte
}

// EBR exported
type EBR struct {
	Status byte
	Fit    byte
	Start  int64
	Size   int64
	Next   int64
	name   [16]byte
}

func (m *MBR) setMBR(size int64) MBR {
	m.Size = size
	m.Signature = rand.Int63()
	return *m
}
