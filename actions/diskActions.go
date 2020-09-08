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

var diskList = []MountedDisk{}
var idChar = 'a'
var idNum = 0

// Disk exported
type Disk struct {
	Size  int
	Route string
	Name  string
	Unit  string
}

// MountedDisk exported
type MountedDisk struct {
	Path          string
	DiskName      string
	PartitionName string
	MountID       string
	Letter        int
	Number        int64
}

// Mount exported
type Mount struct {
	Route string
	Name  string
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
	Name   [16]byte
}

// CreateDisk exported
func (d *Disk) CreateDisk() {
	if d.Name == "" {
		fmt.Println(">> Disk name is missing. Try again.")
	} else if d.Route == "" {
		fmt.Println(">> Disk path is missing. Try again.")
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

// ----------------------------------------- MOUNT --------------------------------------------------------- //

// SetMountRoute exported
func (m *Mount) SetMountRoute(route string) {
	m.Route = route
}

// SetMountName exported
func (m *Mount) SetMountName(name string) {
	m.Name = name
}

// ShowMountedPartitions exported
func ShowMountedPartitions() {
	for _, value := range diskList {
		fmt.Println(">> DISK NAME:", value.DiskName)
		fmt.Println(">> PARTITION NAME:", value.PartitionName)
		fmt.Println(">> PATH:", value.Path)
	}
}

func randChar() string {
	randomChar := 'a' + rune(rand.Intn(26))
	return string(randomChar)
}

// SetMount exported
func (m *Mount) SetMount() {
	file, err := os.OpenFile(m.Route, os.O_RDWR, 0644)
	if err != nil {
		fmt.Println(">> Error reading the file. Try again.")
	}
	defer file.Close()

	// Setting disk name
	re := regexp.MustCompile(`[a-zA-Z]([a-zA-Z]|[0-9])*\.dsk`)
	diskName := re.FindString(m.Route)

	// Verifying and deploying disk information
	mbr := MBR{}
	size := int(unsafe.Sizeof(mbr))
	data := readBytes(file, size)
	mbrBuffer := bytes.NewBuffer(data)
	// Reading bytes to mbr
	binary.Read(mbrBuffer, binary.BigEndian, &mbr)

	for i := 0; i < 4; i++ {
		if mbr.Partitions[i].Type == byte('E') {
			//Starts reading EBRs
			file.Seek(mbr.Partitions[i].Start, 0)
			ebrAux := EBR{}
			sizeEbr := binary.Size(ebrAux)
			ebrData := readBytes(file, sizeEbr)
			bufferAux := bytes.NewBuffer(ebrData)
			_ = binary.Read(bufferAux, binary.BigEndian, &ebrAux)
			if ebrAux.Next != -1 {
				for ebrAux.Next != -1 {
					// Iterates ebrs until found the last one
					file.Seek(ebrAux.Next, 0)
					ebrData := readBytes(file, sizeEbr)
					bufferAux := bytes.NewBuffer(ebrData)
					_ = binary.Read(bufferAux, binary.BigEndian, &ebrAux)
				}
				if string(ebrAux.Name[:len(m.Name)]) == m.Name {
					partitionLetter, partitionNumber, mounted := setMountID(diskName, string(ebrAux.Name[:len(m.Name)]))
					if mounted == false {
						return
					}
					mountID := "vd" + string(partitionLetter) + strconv.Itoa(int(partitionNumber))
					mountedDisk := MountedDisk{
						Path:          m.Route,
						DiskName:      diskName,
						PartitionName: string(ebrAux.Name[:len(m.Name)]),
						Letter:        partitionLetter,
						Number:        partitionNumber,
						MountID:       mountID,
					}
					diskList = append(diskList, mountedDisk)
					fmt.Println(diskList)
					return
				}
			} else {
				if string(ebrAux.Name[:len(m.Name)]) == m.Name {
					partitionLetter, partitionNumber, mounted := setMountID(diskName, string(ebrAux.Name[:len(m.Name)]))
					if mounted == false {
						return
					}
					mountID := "vd" + string(partitionLetter) + strconv.Itoa(int(partitionNumber))

					mountedDisk := MountedDisk{
						Path:          m.Route,
						DiskName:      diskName,
						PartitionName: string(ebrAux.Name[:len(m.Name)]),
						Letter:        partitionLetter,
						Number:        partitionNumber,
						MountID:       mountID,
					}
					diskList = append(diskList, mountedDisk)
					fmt.Println(diskList)
					return
				}
			}
		}
		if string(mbr.Partitions[i].Name[:len(m.Name)]) == m.Name && mbr.Partitions[i].Type == byte('P') {
			partitionLetter, partitionNumber, mounted := setMountID(diskName, string(mbr.Partitions[i].Name[:len(m.Name)]))
			if mounted == false {
				return
			}

			mountID := "vd" + string(partitionLetter) + strconv.Itoa(int(partitionNumber))
			mountedDisk := MountedDisk{
				Path:          m.Route,
				DiskName:      diskName,
				PartitionName: string(mbr.Partitions[i].Name[:len(m.Name)]),
				Letter:        partitionLetter,
				Number:        partitionNumber,
				MountID:       mountID,
			}
			diskList = append(diskList, mountedDisk)
			fmt.Println(diskList)
			return
		}
	}
	fmt.Println(">> Partition not found. Try again.")
}

func setMountID(diskName string, partitionName string) (int, int64, bool) {
	diskFound := false
	var partitionNumber int64
	var partitionLetter int

	for _, value := range diskList {
		if partitionName == value.PartitionName {
			fmt.Println(">> Partition is already mounted")
			return 0, 0, false
		}
	}

	for _, value := range diskList {
		if diskName == value.DiskName {
			partitionNumber = value.Number + 1
			partitionLetter = value.Letter
			diskFound = true
		}

	}
	if diskFound != true {
		partitionLetter = 'a' + idNum
		partitionNumber = 1
		idNum++
		fmt.Println("NEW DISK!")
	}
	fmt.Println(string(partitionLetter), partitionNumber)
	return (partitionLetter), partitionNumber, true
}

// ----------------------------------------------------------------------------------------------------------- //

// SetAddOption exported
func (f *FDISK) SetAddOption(number string) {
	num, _ := strconv.Atoi(number)
	f.AddNumber = int64(num)
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
		return
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
		return
	}
	mbr := MBR{}
	size := int(unsafe.Sizeof(mbr))
	data := readBytes(file, size)
	buff := bytes.NewBuffer(data)
	_ = binary.Read(buff, binary.BigEndian, &mbr)
	date := string(mbr.Date[:])
	fmt.Println(">> ****** MBR INFORMATION ****** ")
	fmt.Println("      DISK SIZE:", mbr.Size, "bytes")
	fmt.Println("      MBR SIZE:", binary.Size(mbr), "bytes") // Binary.size does not reads structs with slices, use unsafe.sizeof instead
	fmt.Println("      CREATED AT:", date)
	fmt.Println("      SIGNATURE", (mbr.Signature))
	fmt.Println()
	for i := 0; i < 4; i++ {
		status := mbr.Partitions[i].Status
		fmt.Println("     *Partition", i)
		fmt.Println("      Status: ", string(status))
		fmt.Println("      Name:", string(mbr.Partitions[i].Name[:]))
		fmt.Println("      Size:", mbr.Partitions[i].Size, "bytes")
		fmt.Println("      Start:", mbr.Partitions[i].Start)
		fmt.Println("      Fit:", string(mbr.Partitions[i].Fit))
		fmt.Println("      Type:", string(mbr.Partitions[i].Type))
		if mbr.Partitions[i].Type == 'E' {
			file.Seek(mbr.Partitions[i].Start, 0)
			ebr := EBR{}
			sizeEbr := binary.Size(ebr)
			dataEbr := readBytes(file, sizeEbr)
			ebrBuff := bytes.NewBuffer(dataEbr)
			_ = binary.Read(ebrBuff, binary.BigEndian, &ebr)
			fmt.Println(" *Logical ", i)
			fmt.Println("  Next:", ebr.Next)
			fmt.Println("  Nombre " + string(ebr.Name[:]))
			fmt.Println("  Size ", ebr.Size)
			fmt.Println("  Start:", ebr.Start)

			i := 1
			if ebr.Next != -1 {
				for ebr.Next != -1 {
					// Iterates ebrs until found the last one

					file.Seek(ebr.Next, 0)

					ebrData := readBytes(file, sizeEbr)
					bufferAux := bytes.NewBuffer(ebrData)
					binary.Read(bufferAux, binary.BigEndian, &ebr)
					fmt.Println(" *Logical ", i)
					fmt.Println("  Next:", ebr.Next)
					fmt.Println("  Nombre " + string(ebr.Name[:]))
					fmt.Println("  Size ", ebr.Size)
					fmt.Println("  Start:", ebr.Start)
					i++
					if i == 24 {
						fmt.Println(">> You have created the maximum logical partitions available.")
						return
					}
				}
			}

			/*fmt.Println("      NAME EBR:", string(ebr.Name[:]))
			fmt.Println("      NEXT EBR:", (ebr.Next))
			fmt.Println("      START EBR:", (ebr.Start))*/

		}
		fmt.Println()
	}
	//mbrGraph(mbr)
}

func readBytes(file *os.File, size int) []byte {
	bytes := make([]byte, size)
	file.Read(bytes)
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
	Route     string
	Status    byte
	Type      byte
	Fit       byte
	Start     int64
	Size      int64
	Unit      string
	Name      [16]byte
	Delete    string
	AddNumber int64
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
		if len(f.Delete) == 0 && f.AddNumber == 0 {
			f.SetPartitionSize()
			f.getDisk(f.Route)
		} else if len(f.Delete) != 0 {
			f.deletePartition()
		} else if f.AddNumber != 0 {
			f.addSize()
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
	case "p":
		f.Type = byte('P')
	case "e":
		f.Type = byte('E')
	case "l":
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
	defer file.Close()
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
	var positionAux int64
	// Verifying if there's a extended partition
	for i := 0; i < 4; i++ {
		if mbr.Partitions[i].Type == byte('E') {
			extendedpFound = true
			positionAux = mbr.Partitions[i].Start
			mbr.Partitions[i].Size = mbr.Partitions[i].Size - f.Size
		}
	}

	if extendedpFound == true && (f.Type == byte('E') || f.Type == byte('e')) {
		fmt.Println(">> There's already an extended partition. Please try again.")
		return
	}

	if f.Type == byte('l') || f.Type == byte('L') {
		for i := 0; i < 4; i++ {
			if mbr.Partitions[i].Type == byte('E') {
				if f.Size > mbr.Partitions[i].Size {
					fmt.Println(">> Logic partition size is bigger than extended partition. Try again.")
					return
				}
				// NEW EBR
				ebr := EBR{}
				ebr.Name = f.Name
				ebr.Size = f.Size
				ebr.Fit = f.Fit
				ebr.Status = 'T'
				ebr.Next = -1
				mbr.Partitions[i].Size = mbr.Partitions[i].Size - ebr.Size
				// Searching for ebr in extended partition
				file.Seek(positionAux, 0)

				if mbr.Partitions[i].Size < 0 {
					fmt.Println(">> There's not enough space within this partition.")
				}

				i := 0
				ebrAux := EBR{}
				sizeEbr := binary.Size(ebrAux)
				ebrData := readBytes(file, sizeEbr)
				bufferAux := bytes.NewBuffer(ebrData)
				_ = binary.Read(bufferAux, binary.BigEndian, &ebrAux)
				if ebrAux.Status == byte('F') {
					// Writes on the first ebr created when extended partition was.
					ebrAux.Name = f.Name
					ebrAux.Size = f.Size
					ebrAux.Fit = f.Fit
					ebrAux.Status = 'T'
					ebrAux.Next = -1
					ebrAux.Start = positionAux + int64(binary.Size(ebr)) // shows where logic partition starts. Starts at the beginning of the extended partition + EBR size.
					/*file.Seek(ebrAux.Start, 0)
					array := make([]byte, ebrAux.Size)
					for j := 0; j < (int(ebrAux.Size) - 1); j++ {
						array[j] = 'P'
					}
					_, err = file.Write(array)*/
					file.Seek(positionAux, 0)
					fmt.Println("EXT START", positionAux)

					var ebrBuffer bytes.Buffer
					file.Seek(ebrAux.Next, 0)
					binary.Write(&ebrBuffer, binary.BigEndian, &ebrAux)
					_, err = file.Write(ebrBuffer.Bytes())
					if err != nil {
						fmt.Println(">> Problem writing logical partition. Try again.")
					} else {
						fmt.Println(">> First logical partition created.")

					}
					break
				}
				// Getting first EBR
				file.Seek(positionAux, 0)
				sizeEbr = binary.Size(ebrAux)
				ebrData = readBytes(file, sizeEbr)
				bufferAux = bytes.NewBuffer(ebrData)
				_ = binary.Read(bufferAux, binary.BigEndian, &ebrAux)
				if ebrAux.Next != -1 {
					for ebrAux.Next != -1 {
						// Iterates ebrs until found the last one
						i++
						file.Seek(ebrAux.Next, 0)
						ebrData := readBytes(file, sizeEbr)
						bufferAux := bytes.NewBuffer(ebrData)
						_ = binary.Read(bufferAux, binary.BigEndian, &ebrAux)
					}
				}

				// Rewrites ebr
				currentSize := ebrAux.Start + ebrAux.Size
				ebr.Start = currentSize + int64(sizeEbr)
				ebrAux.Next = currentSize
				posAux := ebrAux.Start - int64(sizeEbr)
				file.Seek(posAux, 0)
				// Overwriting previous ebr
				var ebrBuffer1 bytes.Buffer

				binary.Write(&ebrBuffer1, binary.BigEndian, &ebrAux)
				_, err = file.Write(ebrBuffer1.Bytes())
				// ------------------------------------------------------

				// Writing new logical partition
				/*file.Seek(ebr.Start, 0)
				array := make([]byte, ebr.Size)
				for j := 0; j < (int(ebr.Size) - 1); j++ {
					array[j] = 'l'
				}
				_, err = file.Write(array)*/
				ebr.Size = f.Size
				ebr.Next = -1
				var ebrBuffer bytes.Buffer
				file.Seek(ebrAux.Next, 0)
				binary.Write(&ebrBuffer, binary.BigEndian, &ebr)
				_, err = file.Write(ebrBuffer.Bytes())
				if err != nil {
					fmt.Println(">> Problem writing logical partition. Try again.")
				} else {
					fmt.Println(">> Logical partition created.")

				}
				break
			}
		}
	} else {
		// Creates extended or primary partition.
		for i := 0; i < 4; i++ {
			if mbr.Partitions[i].Status == byte('F') {
				mbr.Partitions[i].Size = f.Size
				mbr.Partitions[i].Status = byte('T')
				if f.Fit == ' ' {
					mbr.Partitions[i].Fit = byte('W')
				} else {
					mbr.Partitions[i].Fit = f.Fit
				}
				if f.Fit == ' ' {
					mbr.Partitions[i].Type = byte('P')
				} else {
					mbr.Partitions[i].Type = f.Type
				}
				mbr.Partitions[i].Name = f.Name
				mbr.Partitions[i].Type = f.Type
				if i == 0 {
					mbr.Partitions[i].Start = int64(binary.Size(mbr)) + 1
					file.Seek(mbr.Partitions[i].Start, 0)
				} else {
					mbr.Partitions[i].Start = int64(mbr.Partitions[i-1].Size + mbr.Partitions[i-1].Start + 1)
					file.Seek(mbr.Partitions[i].Start, 0)
				}
				if f.Type == byte('E') || f.Type == byte('e') {
					ebr := EBR{}
					ebr.Name = f.Name
					ebr.Size = 0
					ebr.Fit = byte('W')
					ebr.Status = byte('F')
					ebr.Next = -1
					ebr.Start = mbr.Partitions[i].Start
					var ebrBuffer bytes.Buffer
					file.Seek(mbr.Partitions[i].Start, 0)
					binary.Write(&ebrBuffer, binary.BigEndian, &ebr)
					_, err = file.Write(ebrBuffer.Bytes())
				}
				binary.Write(&buffer, binary.BigEndian, &mbr.Partitions[i])
				file.Write(buffer.Bytes())
				mbr.Size = mbr.Size - mbr.Partitions[i].Size // Disk size after adding partition. MBR size always stays de same.
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
				return
			}
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
		if mbr.Partitions[i].Name == f.Name || mbr.Partitions[i].Type == byte('E') {
			if f.Delete == "full" {
				if mbr.Partitions[i].Type == byte('E') {
					if mbr.Partitions[i].Name == f.Name {
						file.Seek(mbr.Partitions[i].Start, 0)
						size := mbr.Partitions[i].Size
						array := make([]byte, size)
						for j := 0; j < (int(size) - 1); j++ {
							array[j] = 0
						}
						_, err = file.Write(array)
						if err != nil {
							log.Fatal(">> Write failed")
						}
						mbr.Partitions[i].Status = 'F'
						mbr.Partitions[i].Start = -1
						mbr.Partitions[i].Name = [16]byte{0}
						mbr.Partitions[i].Fit = ' '
						mbr.Partitions[i].Type = ' '
						mbr.Partitions[i].Size = 0
						break
					}
					sizeAux := mbr.Partitions[i].Size
					fmt.Println("SIZE", sizeAux)

					file.Seek(mbr.Partitions[i].Start, 0)
					// First EBR
					ebrAux := EBR{}
					sizeEbr := binary.Size(ebrAux)
					ebrData := readBytes(file, sizeEbr)
					bufferAux := bytes.NewBuffer(ebrData)
					_ = binary.Read(bufferAux, binary.BigEndian, &ebrAux)
					if f.Name == ebrAux.Name {
						fmt.Println(">> First logical partition cannot be deleted.")
						return
					}
					prevEbr := EBR{}
					if ebrAux.Next != -1 {
						for ebrAux.Next != -1 && f.Name != ebrAux.Name {
							// Iterates ebrs until found the last one
							prevEbr = ebrAux
							i++
							file.Seek(ebrAux.Next, 0)
							ebrData := readBytes(file, sizeEbr)
							bufferAux := bytes.NewBuffer(ebrData)
							_ = binary.Read(bufferAux, binary.BigEndian, &ebrAux)

						}
					}
					if f.Name == ebrAux.Name {

						if ebrAux.Next == -1 {
							// Removes actual partition
							pos := ebrAux.Start - int64(binary.Size(ebrAux))
							file.Seek(pos, 0)
							size := ebrAux.Size + int64(binary.Size(ebrAux))
							array := make([]byte, size)
							for j := 0; j < (int(size) - 1); j++ {
								array[j] = 0
							}
							_, err = file.Write(array)
							if err != nil {
								log.Fatal(">> Write failed")
							}
							prevEbr.Next = -1
							// Overwrites prev ebr
							prevPos := prevEbr.Start - int64(binary.Size(ebrAux))
							file.Seek(prevPos, 0)
						} else {
							prevEbr.Next = ebrAux.Next
							pos := ebrAux.Start - int64(binary.Size(ebrAux))
							file.Seek(pos, 0)
							size := ebrAux.Size + int64(binary.Size(ebrAux))
							array := make([]byte, size)
							for j := 0; j < (int(size) - 1); j++ {
								array[j] = 0
							}
							_, err = file.Write(array)
							if err != nil {
								log.Fatal(">> Write failed")
							}
							prevPos := prevEbr.Start - int64(binary.Size(ebrAux))
							file.Seek(prevPos, 0)
						}
						var ebrBuffer1 bytes.Buffer
						binary.Write(&ebrBuffer1, binary.BigEndian, &prevEbr)
						_, err = file.Write(ebrBuffer1.Bytes())
						if err != nil {
							fmt.Println(">> Error deleting logical partition-")
							break
						}
						fmt.Println(">> Logical partition " + string(ebrAux.Name[:]) + " removed.")
						// Resizing extended partition
						mbr.Partitions[i].Size = sizeAux + ebrAux.Size + int64(binary.Size(ebrAux))
					}

				}
				file.Seek(mbr.Partitions[i].Start, 0)
				size := mbr.Partitions[i].Size
				array := make([]byte, size)
				for j := 0; j < (int(size) - 1); j++ {
					array[j] = 0
				}
				_, err = file.Write(array)
				if err != nil {
					log.Fatal(">> Write failed")
				}
				mbr.Partitions[i].Status = 'F'
				mbr.Partitions[i].Start = -1
				mbr.Partitions[i].Name = [16]byte{0}
				mbr.Partitions[i].Fit = ' '
				mbr.Partitions[i].Type = ' '
				mbr.Partitions[i].Size = 0

			} else if f.Delete == "fast" {

				if mbr.Partitions[i].Type == byte('E') {
					file.Seek(mbr.Partitions[i].Start, 0)
					// First EBR
					ebrAux := EBR{}
					sizeEbr := binary.Size(ebrAux)
					ebrData := readBytes(file, sizeEbr)
					bufferAux := bytes.NewBuffer(ebrData)
					_ = binary.Read(bufferAux, binary.BigEndian, &ebrAux)
					if f.Name == ebrAux.Name {
						fmt.Println(">> First logical partition cannot be deleted.")
						return
					}
					prevEbr := EBR{}
					if ebrAux.Next != -1 {
						for ebrAux.Next != -1 && f.Name != ebrAux.Name {
							// Iterates ebrs until found the last one
							prevEbr = ebrAux
							i++
							file.Seek(ebrAux.Next, 0)
							ebrData := readBytes(file, sizeEbr)
							bufferAux := bytes.NewBuffer(ebrData)
							_ = binary.Read(bufferAux, binary.BigEndian, &ebrAux)

						}
					}
					if f.Name == ebrAux.Name {
						if ebrAux.Next == -1 {
							// Removes actual partition
							prevEbr.Next = -1
							// Overwrites prev ebr
							prevPos := prevEbr.Start - int64(binary.Size(ebrAux))
							file.Seek(prevPos, 0)
						} else {
							prevEbr.Next = ebrAux.Next
							prevPos := prevEbr.Start - int64(binary.Size(ebrAux))
							file.Seek(prevPos, 0)
						}
						var ebrBuffer1 bytes.Buffer
						binary.Write(&ebrBuffer1, binary.BigEndian, &prevEbr)
						_, err = file.Write(ebrBuffer1.Bytes())
						if err != nil {
							fmt.Println(">> Error deleting logical partition-")
							return

						} else {
							fmt.Println(">> Logical partition " + string(ebrAux.Name[:]) + " removed.")
							break
						}
					}

				}

				mbr.Partitions[i].Status = 'F'
				mbr.Partitions[i].Start = -1
				mbr.Partitions[i].Name = [16]byte{0}
				mbr.Partitions[i].Fit = ' '
				mbr.Partitions[i].Type = ' '
				mbr.Partitions[i].Size = 0
			} else {
				fmt.Println(">> Error deleting partition due to incorrect input parameter.")
				return
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
		fmt.Println(">> Partition removed and disk updated.")
	}
}

func (f *FDISK) addSize() {
	size := f.SetPartitionUnit(f.AddNumber)
	file, err := os.OpenFile(f.Route, os.O_RDWR, 0644)
	if err != nil {
		fmt.Println(">> Error reading the file. Try again.")
	}
	defer file.Close()
	// Verifying and deploying disk information
	file.Seek(0, 0)
	mbr := MBR{}
	mbrSize := int(unsafe.Sizeof(mbr))
	data := readBytes(file, mbrSize)
	// Reading bytes to mbr
	mbrBuff := bytes.NewBuffer(data)
	binary.Read(mbrBuff, binary.BigEndian, &mbr)
	for i := 0; i < 4; i++ {
		if mbr.Partitions[i].Type == 'E' {
			file.Seek(mbr.Partitions[i].Start, 0)
			ebrAux := EBR{}
			sizeEbr := binary.Size(ebrAux)
			dataEbr := readBytes(file, sizeEbr)
			ebrBuff := bytes.NewBuffer(dataEbr)
			_ = binary.Read(ebrBuff, binary.BigEndian, &ebrAux)
			i := 1
			if ebrAux.Next != -1 {
				for ebrAux.Next != -1 && ebrAux.Name != f.Name {
					// Iterates ebrs until found the last one
					file.Seek(ebrAux.Next, 0)
					ebrData := readBytes(file, sizeEbr)
					bufferAux := bytes.NewBuffer(ebrData)
					binary.Read(bufferAux, binary.BigEndian, &ebrAux)
					fmt.Println(string(ebrAux.Name[:]))
					i++
					if i == 24 {
						return
					}
				}
			}
			if ebrAux.Name == f.Name {
				if size < 0 {
					ebrAux.Size = ebrAux.Size + (size)
					if ebrAux.Size < 0 {
						fmt.Println(">> You have passed the maximum size to reduce. Try again.")
						return
					}
				} else {
					ebrAux.Size = ebrAux.Size + (size)
				}
				actualPos := ebrAux.Start - int64(binary.Size(ebrAux))
				file.Seek(actualPos, 0)
				var ebrBuffer1 bytes.Buffer
				binary.Write(&ebrBuffer1, binary.BigEndian, &ebrAux)
				_, err = file.Write(ebrBuffer1.Bytes())
				return
			}
		} else {
			if mbr.Partitions[i].Name == f.Name {
				if size < 0 {
					fmt.Println("PREV SIZE:", mbr.Partitions[i].Size)
					mbr.Partitions[i].Size = mbr.Partitions[i].Size + (size)
					fmt.Println("NEW SIZE:", mbr.Partitions[i].Size)
					if mbr.Partitions[i].Size < 0 {
						fmt.Println(">> You have reduced the maximum size allowed. Try again.")
						return
					}

				} else {
					fmt.Println("PREV SIZE:", mbr.Partitions[i].Size)
					mbr.Partitions[i].Size = mbr.Partitions[i].Size + (size)
					fmt.Println("NEW SIZE:", mbr.Partitions[i].Size)
				}
				var mbrBuffer bytes.Buffer
				file.Seek(0, 0)
				binary.Write(&mbrBuffer, binary.BigEndian, &mbr)
				_, er := file.Write(mbrBuffer.Bytes())
				if er != nil {
					fmt.Println(">>", er)
				} else {
					fmt.Println(">> Partition " + string(mbr.Partitions[i].Name[:]) + " resized.")
				}
				return
			}
		}
	}
	fmt.Println(">> Couln't find any coincidence.")
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

	e := exec.Command("dot", "-Tpng", "/home/carlosngv/mbr.txt", "-o", "/home/carlosngv/mbr.png")
	if er := e.Run(); er != nil {
		fmt.Println(">> Error", er)
	}
}
