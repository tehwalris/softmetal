package partition

import (
	"fmt"

	pb "git.dolansoft.org/philippe/softmetal/pb"
	"github.com/rekby/gpt"
)

func AssertExistingPartitionsMatchExact(
	table *gpt.Table, expectedPartitions []pb.FlashingConfig_Partition,
) error {
	for i, _ := range expectedPartitions {
		e := &expectedPartitions[i]
		for j, _ := range table.Partitions {
			a := &table.Partitions[j]
			if MatchesId(a, &e.PartUuid) {
				if !Matches(a, e, table.SectorSize) {
					return fmt.Errorf(
						"Partition %v exists, but does not have expected type or size.",
						e.PartUuid,
					)
				}
			}
		}
	}
	return nil
}
