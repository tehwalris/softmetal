package disk

import (
	"fmt"

	"github.com/jaypipes/ghw"
)

func FindDisk(
	blockInfo *ghw.BlockInfo, targetDiskSerial string,
) (disk *ghw.Disk, found bool, err error) {
	if targetDiskSerial == "" || targetDiskSerial == "unknown" {
		return nil, false, fmt.Errorf(
			"Bad serial number \"%v\" as search target.",
			targetDiskSerial,
		)
	}
	var matching *ghw.Disk = nil
	for _, d := range blockInfo.Disks {
		if d.SerialNumber == targetDiskSerial {
			if matching != nil {
				return nil, false, fmt.Errorf(
					"Encountered duplicate serial number (%v) while searching for disk.",
					targetDiskSerial,
				)
			}
			matching = d
		}
	}
	return matching, (matching != nil), nil
}
