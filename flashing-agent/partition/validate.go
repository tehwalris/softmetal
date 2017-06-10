package partition

import (
	"errors"
	"sort"

	"github.com/rekby/gpt"
)

func AssertGptCompatible(disk, img gpt.Table) error {
	if disk.SectorSize != img.SectorSize {
		return errors.New("Mismated sector sizes")
	}
	return nil
}

func isInsideTable(partition *gpt.Partition, table *gpt.Table) bool {
	return partition.FirstLBA >= table.Header.FirstUsableLBA &&
		partition.FirstLBA <= table.Header.LastUsableLBA &&
		partition.LastLBA >= table.Header.FirstUsableLBA &&
		partition.LastLBA <= table.Header.LastUsableLBA
}

func AssertValidLayout(table *gpt.Table) error {
	partitions := make([]gpt.Partition, len(table.Partitions))
	copy(table.Partitions, partitions)
	sort.Slice(partitions, func(i, j int) bool {
		return partitions[i].FirstLBA < partitions[j].FirstLBA
	})
	for i, p := range partitions {
		if !isInsideTable(&p, table) || p.FirstLBA > p.LastLBA {
			return errors.New("Partition has invalid first/last LBA.")
		}
		if i != 0 && partitions[i-1].LastLBA >= partitions[i].FirstLBA {
			return errors.New("Partitions overlap.")
		}
	}
	return nil
}
