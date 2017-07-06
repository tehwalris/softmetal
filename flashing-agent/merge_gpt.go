package main

import (
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
	// TODO validate inputs (persistent partitions, ...) (!)
	//	- persistent have unique ids, nozero sizes, ok types
	//  - imageGpt does not contain persistent
	if e := partition.AssertExistingPartitionsMatchExact(diskGpt, persistent); e != nil {
		return e
	}

	partition.RemoveExcept(diskGpt, partitionsToIds(persistent))
	for i, _ := range persistent {
		partition.AddPersistentIfMissing(diskGpt, &persistent[i])
	}
	for _, p := range imageGpt.Partitions {
		if !p.IsEmpty() {
			if e := partition.Add(diskGpt, &p); e != nil {
				return e
			}
		}
	}

	if e := partition.AssertGptValid(diskGpt); e != nil {
		return e
	}
	return nil
}
