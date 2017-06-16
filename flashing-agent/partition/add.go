package partition

import (
	"github.com/rekby/gpt"
)

const standardPartitionEntrySize = 128

func Add(table *gpt.Table, p *gpt.Partition) error {
	// TODO This is completely wrong, must look for an
	// empty (type 0 partition) and relpace it
	/*
		p.TrailingBytes = make([]byte,
			table.Header.PartitionEntrySize-standardPartitionEntrySize)
		table.Header.PartitionsArrLen += 1
		minFirstUsableByte := table.Header.PartitionEntrySize * table.Header.PartitionsArrLen
		minFirstUsableLBA := minFirstUsableByte / table.SectorSize
		if minFirstUsableByte%table.SectorSize != 0 {
			minFirstUsableLBA += 1
		}
		if table.Header.FirstUsableLBA < minFirstUsableLBA {
			return errors.New("Can't add partition. No space left in partition table.")
		}
	*/
	return nil
	// TODO actually append a copy of the partition
}
