package partition

import (
	"github.com/rekby/gpt"
)

func clearPartition(p *gpt.Partition) {
	p.Type = gpt.PartType([16]byte{})
	// The rest of these do not have to be zeroed, but it's cleaner
	p.Id = gpt.Guid([16]byte{})
	p.FirstLBA = 0
	p.LastLBA = 0
	p.PartNameUTF16 = [72]byte{}
	p.Flags = gpt.Flags([8]byte{})
}

func Remove(table *gpt.Table, targetUuid *string) (removed bool) {
	for i, _ := range table.Partitions {
		p := &table.Partitions[i]
		if !p.IsEmpty() && MatchesId(p, targetUuid) {
			clearPartition(p)
			return true
		}
	}
	return false
}

func RemoveExcept(table *gpt.Table, targetUuids []string) {
OUTER:
	for i, _ := range table.Partitions {
		p := &table.Partitions[i]
		if p.IsEmpty() {
			continue
		}
		for j, _ := range targetUuids {
			if MatchesId(p, &targetUuids[j]) {
				continue OUTER
			}
		}
		clearPartition(p)
	}
}
