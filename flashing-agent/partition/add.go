package partition

import (
	"errors"

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
