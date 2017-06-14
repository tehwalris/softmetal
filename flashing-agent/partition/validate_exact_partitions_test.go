package partition

import (
	"testing"

	pb "git.dolansoft.org/philippe/softmetal/pb"

	"github.com/rekby/gpt"
)

func TestAssertExistingPartitionsMatchExact(t *testing.T) {
	var cases = []struct {
		table              gpt.Table
		expectedPartitions []pb.FlashingConfig_Partition
		shouldFail         bool
	}{
		{gpt.Table{
			SectorSize: 512,
			Partitions: []gpt.Partition{},
		}, []pb.FlashingConfig_Partition{}, false},
		{gpt.Table{
			SectorSize: 512,
			Partitions: []gpt.Partition{},
		}, []pb.FlashingConfig_Partition{
			{PartUuid: testUuidStrings[1], GptType: testUuidStrings[2], Size: 2048},
		}, false},
		{gpt.Table{
			SectorSize: 1024,
			Partitions: []gpt.Partition{
				{Id: testUuids[1], Type: gpt.PartType(testUuids[2]),
					FirstLBA: 5, LastLBA: 6},
			},
		}, []pb.FlashingConfig_Partition{
			{PartUuid: testUuidStrings[1], GptType: testUuidStrings[2], Size: 2048},
		}, false},
		{gpt.Table{
			SectorSize: 512,
			Partitions: []gpt.Partition{
				{Id: testUuids[1], Type: gpt.PartType(testUuids[2]),
					FirstLBA: 5, LastLBA: 8},
			},
		}, []pb.FlashingConfig_Partition{
			{PartUuid: testUuidStrings[1], GptType: testUuidStrings[2], Size: 1024},
		}, true},
		{gpt.Table{
			SectorSize: 512,
			Partitions: []gpt.Partition{
				{Id: testUuids[1], Type: gpt.PartType(testUuids[2]),
					FirstLBA: 5, LastLBA: 6},
				{Id: testUuids[2], Type: gpt.PartType(testUuids[1]),
					FirstLBA: 7, LastLBA: 8},
			},
		}, []pb.FlashingConfig_Partition{
			{PartUuid: testUuidStrings[1], GptType: testUuidStrings[2], Size: 1024},
			{PartUuid: testUuidStrings[2], GptType: testUuidStrings[1], Size: 1024},
		}, false},
		{gpt.Table{
			SectorSize: 512,
			Partitions: []gpt.Partition{
				{Id: testUuids[1], Type: gpt.PartType(testUuids[2]),
					FirstLBA: 5, LastLBA: 6},
				{Id: testUuids[2], Type: gpt.PartType(testUuids[2]),
					FirstLBA: 7, LastLBA: 8},
			},
		}, []pb.FlashingConfig_Partition{
			{PartUuid: testUuidStrings[1], GptType: testUuidStrings[2], Size: 1024},
			{PartUuid: testUuidStrings[2], GptType: testUuidStrings[1], Size: 1024},
		}, true},
	}
	for i, c := range cases {
		e := AssertExistingPartitionsMatchExact(&c.table, c.expectedPartitions)
		if c.shouldFail && e == nil {
			t.Errorf("Test case %v: Excpected error, but none occured", i)
		} else if !c.shouldFail && e != nil {
			t.Errorf("Test case %v: Excpected no error, but got: %v", i, e)
		}
	}
}
