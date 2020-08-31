package actions

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
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
	//readFile(d.Route, d.Name)
}

func writeMBR(file *os.File, bytes []byte) {
	_, err := file.Write(bytes)
	if err != nil {
		fmt.Println(">> Error writing file. Try again.")
	} else {
		fmt.Println(">> Disk successfully created!")
	}
}

// ReadFile exported
func ReadFile(route string) {
	file, err := os.Open(route)
	defer file.Close()
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
		fmt.Println("Partition ", i)
		fmt.Println("Status: ", string(status))
		fmt.Println("Name:", string(mbr.Partitions[i].Name[:]))
		fmt.Println("Size:", mbr.Partitions[i].Size, "bytes")
		fmt.Println("Start:", mbr.Partitions[i].Start)
		fmt.Println("Fit:", string(mbr.Partitions[i].Fit))
		fmt.Println("Type:", string(mbr.Partitions[i].Type))

		fmt.Println()
	}
	//mbrGraph(mbr)
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
	case "M":
		size = int64(sizeU * 1024 * 1024)
	case "k":
		size = int64(sizeU * 1024)
	case "K":
		size = int64(sizeU * 1024)
	default:
		size = int64(sizeU * 1024 * 1024)
	}

	return size
}

// PARTITIONS

// FDISK exported
type FDISK struct {
	Route  string
	Status byte
	Type   byte
	Fit    byte
	Start  int64
	Size   int64
	Unit   string
	Name   [16]byte
	Delete string
}

// CreatePartition exported
func (f *FDISK) CreatePartition() {

	if f.Fit == 0 {
		f.Fit = byte('W')
	}
	if f.Type == 0 {
		f.Type = byte('P')
	}

	if len(f.Name) == 0 {
		fmt.Println(">> Partition name is missing. Try again.")
	} else if f.Route == "" {
		fmt.Println(">> Partition path is missing. Try again.")
	} else {
		if len(f.Delete) == 0 {
			f.SetPartitionSize()
			f.getDisk(f.Route)
		} else {
			f.deletePartition()
		}
	}

}

// SetPartitionName exported
func (f *FDISK) SetPartitionName(extName string) {
	copy(f.Name[:], extName)
}

// SetPartitionRoute exported
func (f *FDISK) SetPartitionRoute(extRoute string) {
	f.Route = extRoute
}

// SetPSize exported
func (f *FDISK) SetPSize(extSize string) {
	i, _ := strconv.Atoi(extSize)
	f.Size = int64(i)
}

// SetPartitionSize exported
func (f *FDISK) SetPartitionSize() {
	f.Size = f.SetPartitionUnit(f.Size)
}

// SetPartitionFit exported
func (f *FDISK) SetPartitionFit(extFit string) {
	strings.ToUpper(extFit)
	switch extFit {
	case "BF":
		f.Fit = byte('B')
	case "WF":
		f.Fit = byte('W')
	case "FF":
		f.Fit = byte('F')
	default:
		fmt.Println(">> Please, enter a valid option.")
	}
}

// SetPartitionType exported
func (f *FDISK) SetPartitionType(extType string) {
	switch extType {
	case "P":
		f.Type = byte('P')
	case "E":
		f.Type = byte('E')
	case "L":
		f.Type = byte('L')
	default:
		fmt.Println(">> Please, enter a valid option.")
	}
}

// SetFUnit exported
func (f *FDISK) SetFUnit(unit string) {
	f.Unit = unit
}

// SetDeleteOption exported
func (f *FDISK) SetDeleteOption(option string) {
	f.Delete = option
}

// SetPartitionUnit exported
func (f *FDISK) SetPartitionUnit(sizeU int64) int64 {
	var size int64
	switch f.Unit {
	case "m": // Megabytes
		size = int64(sizeU * 1024 * 1024)
	case "M": // Megabytes
		size = int64(sizeU * 1024 * 1024)
	case "k": // Kilobytes
		size = int64(sizeU * 1024)
	case "K": // Kilobytes
		size = int64(sizeU * 1024)
	case "b": // Bytes
		size = int64(sizeU)
	case "B": // Bytes
		size = int64(sizeU)
	default: // Kilobytes
		size = int64(sizeU * 1024)
	}
	return size
}

func (f *FDISK) getDisk(route string) {
	extendedpFound := false
	file, err := os.OpenFile(route, os.O_RDWR, 0644)
	if err != nil {
		fmt.Println(">> Error reading the file. Try again.")
	}
	// Verifying and deploying disk information
	mbr := MBR{}
	size := int(unsafe.Sizeof(mbr))
	data := readBytes(file, size)
	// Reading bytes to mbr
	buff := bytes.NewBuffer(data)
	binary.Read(buff, binary.BigEndian, &mbr)
	// Creating partition
	var buffer bytes.Buffer
	file.Seek(0, 0) // Positioning at the beginning of the file

	// Verifying if there's a extended partition
	for i := 0; i < 4; i++ {
		if mbr.Partitions[i].Type == 'E' {
			extendedpFound = true
		}
	}

	if extendedpFound && (f.Type == 'E' || f.Type == 'b') {
		fmt.Println(">> There's already an extended partition. Please try again.")
		return
	}

	for i := 0; i < 4; i++ {
		if mbr.Partitions[i].Status == 'F' {
			mbr.Partitions[i].Size = f.Size
			mbr.Partitions[i].Status = 'T'
			if f.Fit == ' ' {
				mbr.Partitions[i].Fit = 'W'
			} else {
				mbr.Partitions[i].Fit = f.Fit
			}
			if f.Fit == ' ' {
				mbr.Partitions[i].Type = 'P'
			} else {
				mbr.Partitions[i].Type = f.Type
			}
			mbr.Partitions[i].Name = f.Name
			mbr.Partitions[i].Type = f.Type
			if i == 0 {
				mbr.Partitions[i].Start = int64(binary.Size(mbr)) + 1
				file.Seek(mbr.Partitions[i].Start, 0)
			} else {
				mbr.Partitions[i].Start = int64(mbr.Partitions[i-1].Size + 1)
				file.Seek(mbr.Partitions[i].Start, 0)
			}
			binary.Write(&buffer, binary.BigEndian, &mbr.Partitions[i])
			file.Write(buffer.Bytes())
			mbr.Size = mbr.Size - mbr.Partitions[i].Size // Disk size after adding partition. MBR size always stays de same.
			break
		}
	}
	// Rewriting MBR
	var mbrBuffer bytes.Buffer
	file.Seek(0, 0)
	binary.Write(&mbrBuffer, binary.BigEndian, &mbr)
	_, er := file.Write(mbrBuffer.Bytes())
	if er != nil {
		fmt.Println(">>", er)
	} else {
		fmt.Println(">> Partition created.")
	}

	file.Close()
}

func (f *FDISK) deletePartition() {
	file, err := os.OpenFile(f.Route, os.O_RDWR, 0644)
	if err != nil {
		fmt.Println(">> Error reading the file. Try again.")
	}
	mbr := MBR{}
	size := int(unsafe.Sizeof(mbr))
	data := readBytes(file, size)
	// Reading bytes to mbr
	buff := bytes.NewBuffer(data)
	binary.Read(buff, binary.BigEndian, &mbr)
	file.Seek(0, 0)

	for i := 0; i < 4; i++ {
		if mbr.Partitions[i].Name == f.Name {
			if f.Delete == "full" {
				file.Seek(mbr.Partitions[i].Size, 0)
				size := mbr.Partitions[i].Size
				array := make([]byte, size)
				for j := 0; j < (int(size) - 1); j++ {
					array[i] = 0
				}
				_, err = file.Write(array)
				if err != nil {
					log.Fatal(">> Write failed")
				}

			} else if f.Delete == "fast" {
				mbr.Partitions[i].Status = 'F'
				mbr.Partitions[i].Start = -1
				mbr.Partitions[i].Name = [16]byte{0}
				mbr.Partitions[i].Fit = ' '
				mbr.Partitions[i].Type = ' '
				mbr.Partitions[i].Size = 0
			}
		}
	}
	var mbrBuffer bytes.Buffer
	file.Seek(0, 0)
	binary.Write(&mbrBuffer, binary.BigEndian, &mbr)
	_, er := file.Write(mbrBuffer.Bytes())
	if er != nil {
		fmt.Println(">>", er)
	} else {
		fmt.Println(">> Partition removed.")
	}
}

func mbrGraph(mbr MBR) {
	cont := 4
	f, err := os.Create("mbr.txt")
	defer f.Close()
	if err != nil {
		fmt.Println(">> Error drawing graph!")
	}

	f.WriteString("digraph H { \n node [shape=plaintext];\n")
	f.WriteString(" B [ label=< <TABLE BORDER =\"0\" CELLBORDER=\"1\" CELLSPACING=\"0\">\n")
	f.WriteString("<TR PORT=\"header\">")
	f.WriteString("<TD COLSPAN=\"2\">MBR</TD>")
	f.WriteString("</TR>\n")
	f.WriteString("<TR><TD>Name</TD><TD>Value</TD></TR>\n")
	f.WriteString("<TR><TD PORT=\"1\">MBR_SIZE</TD><TD> " + strconv.Itoa(int(mbr.Size)) + " bytes</TD></TR>\n")
	f.WriteString("<TR><TD PORT=\"2\">MBR_CREATED_AT</TD><TD>  " + string(mbr.Date[:19]) + "</TD></TR>\n")
	f.WriteString("<TR><TD PORT=\"3\">MBR_SIGNATURE</TD><TD> " + strconv.Itoa(int(mbr.Signature)) + "</TD></TR>\n")

	for i := 0; i < 4; i++ {
		status := mbr.Partitions[i].Status
		tpe := mbr.Partitions[i].Type
		fit := mbr.Partitions[i].Fit
		name := (mbr.Partitions[i].Name[:16])
		size := mbr.Partitions[i].Size

		f.WriteString("<TR><TD PORT=\"" + strconv.Itoa(cont) + "\">PARTITION</TD><TD>" + strconv.Itoa(i) + "</TD></TR>\n")
		cont++
		f.WriteString("<TR><TD PORT=\"" + strconv.Itoa(cont) + "\">PARTITION_NAME</TD><TD>" + string(name) + "</TD></TR>\n")
		cont++
		f.WriteString("<TR><TD PORT=\"" + strconv.Itoa(cont) + "\">PARTITION_SIZE</TD><TD>" + strconv.Itoa(int(size)) + "</TD></TR>\n")
		cont++
		f.WriteString("<TR><TD PORT=\"" + strconv.Itoa(cont) + "\">PARTITION_STATUS</TD><TD>" + string(status) + "</TD></TR>\n")
		cont++
		f.WriteString("<TR><TD PORT=\"" + strconv.Itoa(cont) + "\">PARTITION_TYPE</TD><TD>" + string(tpe) + "</TD></TR>\n")
		cont++
		f.WriteString("<TR><TD PORT=\"" + strconv.Itoa(cont) + "\">PARTITION_FIT</TD><TD>" + string(fit) + "</TD></TR>\n")
		cont++
		f.WriteString("<TR><TD PORT=\"" + strconv.Itoa(cont) + "\">PARTITION_START</TD><TD>" + strconv.Itoa(int(mbr.Partitions[i].Start)) + "</TD></TR>\n")
		cont++
	}

	f.WriteString("</TABLE> >];\n")
	f.WriteString("}")

	e := exec.Command("dot", "-Tpng", "mbr.txt", "-o mbr.png")
	if er := e.Run(); er != nil {
		fmt.Println(">> Error", er)
	}
}
