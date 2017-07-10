package disk

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/google/uuid"
	"github.com/jaypipes/ghw"
	"github.com/rekby/gpt"
)

func createGpt(d *ghw.Disk) (*gpt.Table, error) {
	sizeBlocks := d.SizeBytes / d.SectorSizeBytes
	if d.SizeBytes%d.SectorSizeBytes != 0 {
		return nil, fmt.Errorf(
			"Expected disk size to be a multiple of sector size. Was %v bytes with %v byte sectors.",
			d.SizeBytes, d.SectorSizeBytes,
		)
	}
	if sizeBlocks < 4096 {
		// Since the GPT library does not check for extremely small disks,
		// we do it here. This means disks have to be at least 2M with typical
		// sector size (512 bytes).
		return nil, fmt.Errorf(
			"Disk too small to use. (%v sectors, %v bytes per sector)",
			sizeBlocks, d.SectorSizeBytes,
		)
	}
	if d.SectorSizeBytes < 128 {
		// Something is probably very wrong if we get here
		return nil, fmt.Errorf("Unexpectedly small sector size (%v bytes).", d.SectorSizeBytes)
	}
	randomUuid, e := uuid.NewRandom()
	if e != nil {
		return nil, e
	}
	table := gpt.Table{
		SectorSize: d.SectorSizeBytes,
		Header: gpt.Header{
			Signature:               [8]byte{0x45, 0x46, 0x49, 0x20, 0x50, 0x41, 0x52, 0x54},
			Revision:                0x00010000, // Version 1.0
			Size:                    0x0000005C, // 92 bytes
			CRC:                     0,          // Set during save
			Reserved:                0,
			HeaderStartLBA:          1,
			HeaderCopyStartLBA:      0, // Set by CreateTableForNewDiskSize
			FirstUsableLBA:          2048,
			LastUsableLBA:           0, // Set by CreateTableForNewDiskSize
			DiskGUID:                gpt.Guid(randomUuid),
			PartitionsTableStartLBA: 2,
			PartitionsArrLen:        128,
			PartitionEntrySize:      128,
			PartitionsCRC:           0,                                  // Set during save
			TrailingBytes:           make([]byte, d.SectorSizeBytes-92), // Rest of sector
		},
		Partitions: make([]gpt.Partition, 128),
	}
	table = table.CreateTableForNewDiskSize(sizeBlocks)
	if table.Header.HeaderCopyStartLBA == 0 || table.Header.LastUsableLBA == 0 {
		return nil, errors.New(
			"Expected GPT table HeaderCopyStartLBA and LastUsableLBA to be set.",
		)
	}
	return &table, nil
}

func GetOrCreateGpt(
	f *os.File, d *ghw.Disk,
) (returnTable *gpt.Table, didCreate bool, err error) {
	if _, e := f.Seek(int64(d.SectorSizeBytes), io.SeekStart); e != nil {
		return nil, false, e
	}
	table, e := gpt.ReadTable(f, d.SectorSizeBytes)
	if e != nil {
		if e.Error() == "Bad GPT signature" {
			// "Bad GPT signature" almost definitely means the disk is empty.
			// This is the only case where a fresh GPT will be created.
			// Other cases fail for safety.
			createdTable, e := createGpt(d)
			if e != nil {
				return nil, false, e
			}
			return createdTable, true, nil
		}
		return nil, false, e
	}
	return &table, false, nil
}
