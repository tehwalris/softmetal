package disk

import (
	"fmt"

	"github.com/jaypipes/ghw"
)

// FindDisk locates a disk with a specific serial number in ghw.BlockInfo.
func FindDisk(
	blockInfo *ghw.BlockInfo, targetDiskSerial string,
) (disk *ghw.Disk, found bool, err error) {
	if targetDiskSerial == "" || targetDiskSerial == "unknown" {
		return nil, false, fmt.Errorf(
			"bad serial number \"%v\" as search target",
			targetDiskSerial,
		)
	}
	var matching *ghw.Disk
	for _, d := range blockInfo.Disks {
		if d.SerialNumber == targetDiskSerial {
			if matching != nil {
				return nil, false, fmt.Errorf(
					"encountered duplicate serial number (%v) while searching for disk",
					targetDiskSerial,
				)
			}
			matching = d
		}
	}
	return matching, (matching != nil), nil
}
