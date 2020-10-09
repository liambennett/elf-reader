package main

import (
	"fmt"
	"os"
	"errors"
	"bytes"
	"encoding/binary"
)

var (
	MAGICELF = []byte { 0x7F, 0x45, 0x4c, 0x46 , 0x02}
)

func main() {
	var err error
	defer func() {
		if err != nil{
			fmt.Println(err)
			os.Exit(1)
		}
	}()

	f, err := os.Open("testcode")
	defer f.Close()
	if err != nil{
		return
	}

	val, err := readBytesFromFile(f, 0, 5)
	if err != nil{
		return
	}

	if !bytes.Equal(val, MAGICELF) {
		err = errors.New("This is not an 64 bit ELF binary")
		return
	}

	header, err := readBytesFromFile(f, 0, 64)
	if err != nil{
		return
	}

	sectionOffset := int64(binary.LittleEndian.Uint64(header[40:48]))
	sectionHeaderSize := int64(binary.LittleEndian.Uint16(header[58:60]))
	sectionCount := int(binary.LittleEndian.Uint16(header[60:62]))
	sectionNameIndex := int64(binary.LittleEndian.Uint16(header[62:64]))

	sectionHeaders := make([][]byte,sectionCount)
	
	for h := 0; h < sectionCount; h++ { 
		val, err = readBytesFromFile(f, int64(sectionOffset + (int64(h) * sectionHeaderSize)), int(sectionHeaderSize))
		if err != nil{
			return
		}
		sectionHeaders[h] = val
		fmt.Printf("sh %x\n", val)
	}

	val, err = readBytesFromFile(f, int64(sectionOffset + (sectionNameIndex * sectionHeaderSize)+ 24), 8)
	if err != nil{
		return
	}
	sectionNameOffset := int64(binary.LittleEndian.Uint64(val))

	sectionNames := make([]string,0,32)
	for _, item := range sectionHeaders {
		start := sectionNameOffset + int64(binary.LittleEndian.Uint16(item))
		nameBytes, err := readBytesUntilNull(f, start)
		if err != nil{
			return
		}
		sectionNames = append(sectionNames,string(nameBytes))
	}

	symbolIndex := 0

	for i, item := range sectionNames {
		if item == ".symtab" {
			symbolIndex = i
			break
		}
	}

	symbolTableOffset := int64(binary.LittleEndian.Uint64(sectionHeaders[symbolIndex][24:32]))
	symbolTableSize := int64(binary.LittleEndian.Uint64(sectionHeaders[symbolIndex][32:40]))
	symbolTableEntrySize := int64(binary.LittleEndian.Uint64(sectionHeaders[symbolIndex][56:64]))
	symbolTableEnd := symbolTableOffset + symbolTableSize
	//fmt.Printf("%x", sectionHeaders[symbolIndex][24:32])
	for{
		if symbolTableOffset >= symbolTableEnd {
			break
		}
		val, err = readBytesFromFile(f, symbolTableOffset, int(symbolTableEntrySize))
		if err != nil{
			return
		}
		fmt.Printf("%x\n", val)
		//fmt.Printf("%v", symbolTableOffset)
		symbolTableOffset += symbolTableEntrySize
	}
}

func readBytesUntilNull(f *os.File, start int64) ([]byte, error){
	offset := int64(0)
	m := make([]byte,0,16)
	for {
		val, err := readBytesFromFile(f, start+offset, 1)
		if err != nil{
			return nil, err
		}
		if bytes.Equal(val,[]byte{0x00}){
			break
		} 
		m = append(m,val...)
		offset++
	}
	return m,nil
}

func readBytesFromFile(f *os.File, start int64, amount int) ([]byte, error) {
	if _, err := f.Seek(start, 0); err != nil {
		return nil, err
	}

	bytes := make([]byte, amount)
	_, err := f.Read(bytes)
	return bytes, err
} 