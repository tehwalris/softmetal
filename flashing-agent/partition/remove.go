package partition

import (
	"github.com/rekby/gpt"
)

func Remove(table *gpt.Table, targetUuid *string) (removed bool) {
	for i, p := range table.Partitions {
		if MatchesId(&p, targetUuid) {
			table.Partitions = append(table.Partitions[:i], table.Partitions[i+1:]...)
			table.Header.PartitionsArrLen -= 1
			return true
		}
	}
	return false
}
