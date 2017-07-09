package main

import (
	"fmt"
	"io"
	"os"

	"git.dolansoft.org/philippe/softmetal/flashing-agent/disk"
	"git.dolansoft.org/philippe/softmetal/flashing-agent/partition"

	"github.com/jaypipes/ghw"
	"github.com/rekby/gpt"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	targetDiskCombinedSerial := "TOSHIBA_THNSFJ256GCSU_46KS117IT8LW"

	blockInfo, e := ghw.Block()
	check(e)
	disk, found, e := disk.FindDisk(blockInfo, targetDiskCombinedSerial)
	check(e)
	if !found {
		panic("Disk not found")
	}

	f, e := os.Open(fmt.Sprintf("/dev/%v", disk.Name))
	check(e)
	_, e = f.Seek(int64(disk.SectorSizeBytes), io.SeekStart)
	check(e)
	table, e := gpt.ReadTable(f, disk.SectorSizeBytes)
	check(e)

	partition.PrintTable(&table)

	// merge gpt
	// write gpt
}
