package partition

import (
	"fmt"

	pb "git.dolansoft.org/philippe/softmetal/pb"
	"github.com/rekby/gpt"
)

func Remove(
	table *gpt.Table, target *pb.FlashingConfig_Partition,
) (removed bool, err error) {
	for i, p := range table.Partitions {
		if MatchesId(&p, &target.PartUuid) {
			if !Matches(&p, target, table.SectorSize) {
				return false, fmt.Errorf("Partition with ID %v does not match expectations. "+
					"Refusing to remove. Expected/actual:\n%+v\n%+v",
					target.PartUuid, target, table)
			}
			table.Partitions = append(table.Partitions[:i], table.Partitions[i+1:]...)
			table.Header.PartitionsArrLen -= 1
			return true, nil
		}
	}
	return false, nil
}
