package main

import (
	"fmt"

	"github.com/jaypipes/ghw"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	block, e := ghw.Block()
	check(e)

	fmt.Println(block.String())
	for _, disk := range block.Disks {
		fmt.Println(disk.String())
		fmt.Println(disk.SerialNumber)
		for _, part := range disk.Partitions {
			fmt.Println(part.String())
		}
	}

	/*
		f, e := os.Open("/dev/sda")
		check(e)

		_, e = f.Seek(512, io.SeekStart) // TODO other block sizes
		check(e)
		table, e := gpt.ReadTable(f, 512)
		check(e)

		fmt.Printf("%+v", &table)
	*/
}
