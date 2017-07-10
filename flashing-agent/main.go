package main

import (
	"git.dolansoft.org/philippe/softmetal/flashing-agent/disk"
	"git.dolansoft.org/philippe/softmetal/flashing-agent/partition"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	f, diskInfo, e := disk.OpenBySerial("TOSHIBA_THNSFJ256GCSU_46KS117IT8LW")
	check(e)
	table, e := disk.GetOrCreateGpt(f, diskInfo)
	check(e)

	partition.PrintTable(table)
	// merge gpt
	// write gpt
}
