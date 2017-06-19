package partition

import (
	"errors"
	"fmt"
	"sort"

	pb "git.dolansoft.org/philippe/softmetal/pb"

	"github.com/rekby/gpt"
)

var zeroUuid = gpt.Guid{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
var zeroPartType = gpt.PartType(zeroUuid)

func AssertGptCompatible(disk, img *gpt.Table) error {
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

func AssertGptValid(table *gpt.Table) error {
	if e := assertValidLayout(table); e != nil {
		return e
	}
	if e := assertUniqueIds(table.Partitions); e != nil {
		return e
	}
	return nil
}

func assertValidLayout(table *gpt.Table) error {
	partitions := make([]gpt.Partition, len(table.Partitions))
	copy(partitions, table.Partitions)
	sort.Slice(partitions, func(i, j int) bool {
		return partitions[i].FirstLBA < partitions[j].FirstLBA
	})
	for i, p := range partitions {
		if p.IsEmpty() {
			continue
		}
		if !isInsideTable(&p, table) || p.FirstLBA > p.LastLBA {
			return errors.New("Partition has invalid first/last LBA.")
		}
		if i != 0 && partitions[i-1].LastLBA >= partitions[i].FirstLBA {
			return errors.New("Partitions overlap.")
		}
	}
	return nil
}

func assertUniqueIds(partitions []gpt.Partition) error {
	ids := make(map[gpt.Guid]bool)
	for _, p := range partitions {
		if p.IsEmpty() {
			continue
		}
		if _, exists := ids[p.Id]; exists {
			return fmt.Errorf("Partition id %v not unique.", p.Id.String())
		}
		ids[p.Id] = true
	}
	return nil
}

func AssertExactMatchIfExists(table *gpt.Table, target *pb.FlashingConfig_Partition) error {
	for i, _ := range table.Partitions {
		p := &table.Partitions[i]
		if !p.IsEmpty() &&
			MatchesId(p, &target.PartUuid) &&
			!Matches(p, target, table.SectorSize) {
			return fmt.Errorf(
				"Partition %v found, but it's type and/or size are not as expected.",
				target.PartUuid,
			)
		}
	}
	return nil
}
