package partition

import (
	"errors"
	"fmt"

	"github.com/rekby/gpt"
)

func Add(table *gpt.Table, p *gpt.Partition) error {
	if p.IsEmpty() {
		return errors.New("Attempted to add an empty partition.")
	}
	for i, _ := range table.Partitions {
		if table.Partitions[i].IsEmpty() {
			table.Partitions[i] = *p
			return nil
		}
	}
	return errors.New("Can't add partition. No space left in partition table.")
}

func AddFindSpace(table *gpt.Table, p *gpt.Partition, side DiskSide) error {
	blocks := p.LastLBA - p.FirstLBA + 1
	firstLBA, lastLBA, found := FindSpace(table, blocks, side)
	if !found {
		return fmt.Errorf("Could not find %v blocks for partition %v",
			blocks, p.Id.String())
	}
	p.FirstLBA = firstLBA
	p.LastLBA = lastLBA
	return Add(table, p)
}
