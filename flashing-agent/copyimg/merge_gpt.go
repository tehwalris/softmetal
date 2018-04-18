package copyimg

import (
	"fmt"

	"git.dolansoft.org/philippe/softmetal/flashing-agent/partition"
	pb "git.dolansoft.org/philippe/softmetal/pb"
	"github.com/rekby/gpt"
)

func partitionsToIds(partitions []pb.FlashingConfig_Partition) []string {
	ids := make([]string, len(partitions))
	for i, _ := range partitions {
		ids[i] = partitions[i].PartUuid
	}
	return ids
}

// WARNING Modifies diskGpt (in memory)
func MergeGpt(
	diskGpt *gpt.Table, imageGpt *gpt.Table, persistent []pb.FlashingConfig_Partition,
) error {
	if e := partition.AssertGptCompatible(diskGpt, imageGpt); e != nil {
		return e
	}
	if e := partition.AssertGptValid(diskGpt); e != nil {
		return e
	}
	if e := partition.AssertGptValid(imageGpt); e != nil {
		return e
	}
	if e := partition.AssertPersistentValid(persistent); e != nil {
		return e
	}
	for i, _ := range persistent {
		id := &persistent[i].PartUuid
		if partition.ContainsId(imageGpt.Partitions, id) {
			return fmt.Errorf(
				"Partition %v in image conflicts with a persistent partition.",
				id,
			)
		}
	}
	if e := partition.AssertExistingPartitionsMatchExact(diskGpt, persistent); e != nil {
		return e
	}

	partition.RemoveExcept(diskGpt, partitionsToIds(persistent))
	for i, _ := range persistent {
		partition.AddPersistentIfMissing(diskGpt, &persistent[i])
	}
	for _, p := range imageGpt.Partitions {
		if !p.IsEmpty() {
			if e := partition.AddFindSpace(diskGpt, &p, partition.Start); e != nil {
				return e
			}
		}
	}

	if e := partition.AssertGptValid(diskGpt); e != nil {
		return e
	}
	return nil
}
