package partition

import (
	"testing"

	pb "git.dolansoft.org/philippe/softmetal/pb"

	"github.com/rekby/gpt"
)

func TestMatches(t *testing.T) {
	var testCases = []struct {
		real          gpt.Partition
		search        pb.FlashingConfig_Partition
		sectorSize    uint64
		shouldMatch   bool
		shouldMatchId bool
	}{
		{gpt.Partition{
			Id:       testUuids[0],
			Type:     gpt.PartType(testUuids[1]),
			FirstLBA: 30,
			LastLBA:  30,
		}, pb.FlashingConfig_Partition{
			PartUuid: testUuidStrings[0],
			GptType:  testUuidStrings[1],
			Size:     512,
		}, 512, true, true},
		{gpt.Partition{
			Id:       testUuids[0],
			Type:     gpt.PartType(testUuids[1]),
			FirstLBA: 50,
			LastLBA:  149,
		}, pb.FlashingConfig_Partition{
			PartUuid: testUuidStrings[0],
			GptType:  testUuidStrings[1],
			Size:     51200,
		}, 512, true, true},
		{gpt.Partition{
			Id:       testUuids[0],
			Type:     gpt.PartType(testUuids[1]),
			FirstLBA: 50,
			LastLBA:  150,
		}, pb.FlashingConfig_Partition{
			PartUuid: testUuidStrings[0],
			GptType:  testUuidStrings[1],
			Size:     51200,
		}, 512, false, true},
		{gpt.Partition{
			Id:       testUuids[1],
			Type:     gpt.PartType(testUuids[1]),
			FirstLBA: 30,
			LastLBA:  30,
		}, pb.FlashingConfig_Partition{
			PartUuid: testUuidStrings[0],
			GptType:  testUuidStrings[1],
			Size:     512,
		}, 512, false, false},
		{gpt.Partition{
			Id:       testUuids[0],
			Type:     gpt.PartType(testUuids[0]),
			FirstLBA: 30,
			LastLBA:  30,
		}, pb.FlashingConfig_Partition{
			PartUuid: testUuidStrings[0],
			GptType:  testUuidStrings[1],
			Size:     512,
		}, 512, false, true},
		{gpt.Partition{
			Id:       testUuids[1],
			Type:     gpt.PartType(testUuids[0]),
			FirstLBA: 30,
			LastLBA:  30,
		}, pb.FlashingConfig_Partition{
			PartUuid: testUuidStrings[0],
			GptType:  testUuidStrings[1],
			Size:     512,
		}, 512, false, false},
		{gpt.Partition{
			Id:       testUuids[0],
			Type:     gpt.PartType(testUuids[1]),
			FirstLBA: 30,
			LastLBA:  30,
		}, pb.FlashingConfig_Partition{
			PartUuid: testUuidStrings[0],
			GptType:  testUuidStrings[1],
			Size:     512,
		}, 1024, false, true},
	}

	for i, tc := range testCases {
		if res := Matches(&tc.real, &tc.search, tc.sectorSize); res != tc.shouldMatch {
			t.Errorf("Test case %v failed. Expected Matches to return %v, got %v.",
				i, tc.shouldMatch, res)
		}
		if res := MatchesId(&tc.real, &tc.search.PartUuid); res != tc.shouldMatchId {
			t.Errorf("Test case %v failed. Expected MatchesId to return %v, got %v.",
				i, tc.shouldMatchId, res)
		}
	}
}

func TestContainsId(t *testing.T) {
	var testCases = []struct {
		partitions []gpt.Partition
		targetId   string
		shouldFind bool
	}{
		{[]gpt.Partition{}, testUuidStrings[1], false},
		{[]gpt.Partition{
			{Id: testUuids[1], Type: gpt.PartType(testUuids[2])},
		}, testUuidStrings[1], true},
		{[]gpt.Partition{
			{Id: testUuids[1], Type: gpt.PartType(testUuids[0])},
			{Id: testUuids[2], Type: gpt.PartType(testUuids[2])},
		}, testUuidStrings[1], false},
	}

	for i, c := range testCases {
		found := ContainsId(c.partitions, &c.targetId)
		if found && !c.shouldFind {
			t.Errorf("Test case %v: Found unexpected matching partition.", i)
		}
		if !found && c.shouldFind {
			t.Errorf("Test case %v: Expected to find a matching partition, but didn't.", i)
		}
	}
}
