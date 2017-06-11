package partition

import (
	"testing"

	pb "git.dolansoft.org/philippe/softmetal/pb"

	"github.com/rekby/gpt"
)

var testUuids = []gpt.Guid{
	{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
	{0x1f, 0xe6, 0x90, 0x41, 0xfc, 0xda, 0xb9, 0x4d, 0x83, 0x21, 0xa5, 0xc9, 0x28, 0x47, 0xf7, 0x6b},
	{0x8f, 0x06, 0xf3, 0x1b, 0x1a, 0xff, 0xe5, 0x43, 0xa2, 0xf1, 0x56, 0x39, 0x59, 0x6e, 0xd2, 0xdd},
	{0x60, 0xcc, 0x55, 0x7c, 0xe0, 0xc5, 0xf5, 0x45, 0x97, 0xa1, 0x20, 0x2c, 0xbd, 0x8a, 0x9a, 0xe8},
	{0x4e, 0x63, 0x2f, 0x23, 0xfe, 0x26, 0x88, 0x44, 0x89, 0x3c, 0x31, 0x26, 0xaa, 0x2d, 0xb5, 0x3b},
}

var testUuidStrings = []string{
	"00000000-0000-0000-0000-000000000000",
	"4190e61f-dafc-4db9-8321-a5c92847f76b",
	"1bf3068f-ff1a-43e5-a2f1-5639596ed2dd",
	"7c55cc60-c5e0-45f5-97a1-202cbd8a9ae8",
	"232f634e-26fe-4488-893c-3126aa2db53b",
}

var testCases = []struct {
	real        gpt.Partition
	search      pb.FlashingConfig_Partition
	sectorSize  uint64
	shouldMatch bool
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
	}, 512, true},
	{gpt.Partition{
		Id:       testUuids[0],
		Type:     gpt.PartType(testUuids[1]),
		FirstLBA: 50,
		LastLBA:  149,
	}, pb.FlashingConfig_Partition{
		PartUuid: testUuidStrings[0],
		GptType:  testUuidStrings[1],
		Size:     51200,
	}, 512, true},
	{gpt.Partition{
		Id:       testUuids[0],
		Type:     gpt.PartType(testUuids[1]),
		FirstLBA: 50,
		LastLBA:  150,
	}, pb.FlashingConfig_Partition{
		PartUuid: testUuidStrings[0],
		GptType:  testUuidStrings[1],
		Size:     51200,
	}, 512, false},
	{gpt.Partition{
		Id:       testUuids[1],
		Type:     gpt.PartType(testUuids[1]),
		FirstLBA: 30,
		LastLBA:  30,
	}, pb.FlashingConfig_Partition{
		PartUuid: testUuidStrings[0],
		GptType:  testUuidStrings[1],
		Size:     512,
	}, 512, false},
	{gpt.Partition{
		Id:       testUuids[0],
		Type:     gpt.PartType(testUuids[0]),
		FirstLBA: 30,
		LastLBA:  30,
	}, pb.FlashingConfig_Partition{
		PartUuid: testUuidStrings[0],
		GptType:  testUuidStrings[1],
		Size:     512,
	}, 512, false},
	{gpt.Partition{
		Id:       testUuids[1],
		Type:     gpt.PartType(testUuids[0]),
		FirstLBA: 30,
		LastLBA:  30,
	}, pb.FlashingConfig_Partition{
		PartUuid: testUuidStrings[0],
		GptType:  testUuidStrings[1],
		Size:     512,
	}, 512, false},
	{gpt.Partition{
		Id:       testUuids[0],
		Type:     gpt.PartType(testUuids[1]),
		FirstLBA: 30,
		LastLBA:  30,
	}, pb.FlashingConfig_Partition{
		PartUuid: testUuidStrings[0],
		GptType:  testUuidStrings[1],
		Size:     512,
	}, 1024, false},
}

func TestMatches(t *testing.T) {
	for i, tc := range testCases {
		res := Matches(&tc.real, &tc.search, tc.sectorSize)
		if res != tc.shouldMatch {
			t.Errorf("Test case %v failed. Expected %v, got %v.", i, tc.shouldMatch, res)
		}
	}
}
