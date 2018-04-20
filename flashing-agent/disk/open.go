package disk

import (
	"fmt"
	"os"

	"github.com/jaypipes/ghw"
)

// OpenBySerial finds a a disk by serial and returns its block device file and metadata.
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
		return nil, nil, fmt.Errorf("disk %v not found", combinedSerial)
	}
	f, e := os.OpenFile(fmt.Sprintf("/dev/%v", d.Name), os.O_RDWR|os.O_TRUNC, 0660)
	if e != nil {
		return nil, nil, e
	}
	return f, d, nil
}
