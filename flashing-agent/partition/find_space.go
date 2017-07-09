package partition

import (
	"sort"

	"github.com/rekby/gpt"
)

type DiskSide int

const (
	Start DiskSide = iota
	End
)

type diskRange struct {
	FirstLBA  uint64
	LastLBA   uint64
	Partition *gpt.Partition
}

func (r *diskRange) enoughBlocksFree(blocks uint64) bool {
	return r.Partition == nil && (r.LastLBA-r.FirstLBA+1) >= blocks
}

func FindSpace(
	table *gpt.Table, blocks uint64, side DiskSide,
) (firstLBA, lastLBA uint64, found bool) {
	if blocks == 0 {
		return 0, 0, false
	}
	diskRanges := calculateDiskRanges(table)
	if side == Start {
		for _, r := range diskRanges {
			if r.enoughBlocksFree(blocks) {
				return r.FirstLBA, r.FirstLBA + blocks - 1, true
			}
		}
	} else {
		for i, _ := range diskRanges {
			if r := diskRanges[len(diskRanges)-1-i]; r.enoughBlocksFree(blocks) {
				return r.LastLBA - blocks + 1, r.LastLBA, true
			}
		}
	}
	return 0, 0, false
}

func calculateDiskRanges(table *gpt.Table) []diskRange {
	diskRanges := make([]diskRange, 0, 2*len(table.Partitions)+1)

	sortedPartitions := make([]*gpt.Partition, 0, len(table.Partitions))
	var j = 0
	for i, _ := range table.Partitions {
		if !table.Partitions[i].IsEmpty() {
			sortedPartitions = sortedPartitions[0 : j+1]
			sortedPartitions[j] = &table.Partitions[i]
			j += 1
		}
	}
	if len(sortedPartitions) == 0 {
		diskRanges = append(diskRanges, diskRange{
			FirstLBA:  table.Header.FirstUsableLBA,
			LastLBA:   table.Header.LastUsableLBA,
			Partition: nil,
		})
		return diskRanges
	}
	sort.Slice(sortedPartitions, func(i, j int) bool {
		return sortedPartitions[i].FirstLBA < sortedPartitions[j].FirstLBA
	})

	for i, p := range sortedPartitions {
		if i == 0 && table.Header.FirstUsableLBA < p.FirstLBA {
			diskRanges = append(diskRanges, diskRange{
				FirstLBA:  table.Header.FirstUsableLBA,
				LastLBA:   p.FirstLBA - 1,
				Partition: nil,
			})
		}
		diskRanges = append(diskRanges, diskRange{
			FirstLBA:  p.FirstLBA,
			LastLBA:   p.LastLBA,
			Partition: p,
		})
		if i+1 == len(sortedPartitions) || table.Header.LastUsableLBA <= p.LastLBA {
			diskRanges = append(diskRanges, diskRange{
				FirstLBA:  p.LastLBA + 1,
				LastLBA:   table.Header.LastUsableLBA,
				Partition: nil,
			})
		} else {
			next := sortedPartitions[i+1]
			if next.FirstLBA > p.LastLBA+1 {
				diskRanges = append(diskRanges, diskRange{
					FirstLBA:  p.LastLBA + 1,
					LastLBA:   next.FirstLBA - 1,
					Partition: nil,
				})
			}
		}
	}
	return diskRanges
}
