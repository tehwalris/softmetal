package disk

import (
	"fmt"
	"os"

	"github.com/jaypipes/ghw"
)

func OpenBySerial(combinedSerial string) (file *os.File, diskInfo *ghw.Disk, err error) {
	blockInfo, e := ghw.Block()
	if e != nil {
		return nil, nil, e
	}
	d, found, e := FindDisk(blockInfo, combinedSerial)
	if e != nil {
		return nil, nil, e
	}
	if !found {
		return nil, nil, fmt.Errorf("Disk %v not found.", combinedSerial)
	}
	// TODO
	f, e := os.Open("/home/philippe/temp/test-gpt.img")
	// f, e := os.Open(fmt.Sprintf("/dev/%v", d.Name))
	if e != nil {
		return nil, nil, e
	}
	return f, d, nil
}
