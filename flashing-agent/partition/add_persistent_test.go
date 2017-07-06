package partition

import (
	"strings"
	"testing"

	pb "git.dolansoft.org/philippe/softmetal/pb"

	"github.com/rekby/gpt"
)

func TestAddPersistent(t *testing.T) {
	var cases = []struct {
		table         gpt.Table
		toAdd         pb.FlashingConfig_Partition
		remaining     []gpt.Partition
		shouldFail    bool
		shouldContain string
	}{
		{gpt.Table{
			SectorSize: 512,
			Header: gpt.Header{
				PartitionsArrLen: 2,
				FirstUsableLBA:   5,
				LastUsableLBA:    70,
			},
			Partitions: []gpt.Partition{
				{FirstLBA: 50, LastLBA: 51,
					Id: testUuids[1], Type: gpt.PartType(testUuids[3])},
				{FirstLBA: 0, LastLBA: 0,
					Id: testUuids[0], Type: gpt.PartType(testUuids[0])},
			},
		}, pb.FlashingConfig_Partition{
			PartUuid: testUuidStrings[3],
			GptType:  testUuidStrings[1],
			Size:     5120,
		}, []gpt.Partition{
			{FirstLBA: 50, LastLBA: 51,
				Id: testUuids[1], Type: gpt.PartType(testUuids[3])},
			{FirstLBA: 61, LastLBA: 70,
				Id: testUuids[3], Type: gpt.PartType(testUuids[1])},
		}, false, ""},

		{gpt.Table{
			SectorSize: 512,
			Header: gpt.Header{
				PartitionsArrLen: 2,
				FirstUsableLBA:   5,
				LastUsableLBA:    70,
			},
			Partitions: []gpt.Partition{
				{FirstLBA: 50, LastLBA: 51,
					Id: testUuids[1], Type: gpt.PartType(testUuids[3])},
				{FirstLBA: 0, LastLBA: 0,
					Id: testUuids[0], Type: gpt.PartType(testUuids[0])},
			},
		}, pb.FlashingConfig_Partition{
			PartUuid: testUuidStrings[3],
			GptType:  testUuidStrings[1],
			Size:     51200,
		}, []gpt.Partition{
			{FirstLBA: 50, LastLBA: 51,
				Id: testUuids[1], Type: gpt.PartType(testUuids[3])},
			{FirstLBA: 0, LastLBA: 0,
				Id: testUuids[0], Type: gpt.PartType(testUuids[0])},
		}, true, "Could not find 100 blocks"},

		{gpt.Table{
			SectorSize: 512,
			Header: gpt.Header{
				PartitionsArrLen: 2,
				FirstUsableLBA:   5,
				LastUsableLBA:    70,
			},
			Partitions: []gpt.Partition{
				{FirstLBA: 50, LastLBA: 51,
					Id: testUuids[1], Type: gpt.PartType(testUuids[3])},
				{FirstLBA: 0, LastLBA: 0,
					Id: testUuids[0], Type: gpt.PartType(testUuids[0])},
			},
		}, pb.FlashingConfig_Partition{
			PartUuid: testUuidStrings[3],
			GptType:  testUuidStrings[1],
			Size:     10240,
		}, []gpt.Partition{
			{FirstLBA: 50, LastLBA: 51,
				Id: testUuids[1], Type: gpt.PartType(testUuids[3])},
			{FirstLBA: 30, LastLBA: 49,
				Id: testUuids[3], Type: gpt.PartType(testUuids[1])},
		}, false, ""},

		{gpt.Table{
			SectorSize: 1024,
			Header: gpt.Header{
				PartitionsArrLen: 2,
				FirstUsableLBA:   5,
				LastUsableLBA:    70,
			},
			Partitions: []gpt.Partition{
				{FirstLBA: 50, LastLBA: 51,
					Id: testUuids[1], Type: gpt.PartType(testUuids[3])},
				{FirstLBA: 0, LastLBA: 0,
					Id: testUuids[0], Type: gpt.PartType(testUuids[0])},
			},
		}, pb.FlashingConfig_Partition{
			PartUuid: testUuidStrings[3],
			GptType:  testUuidStrings[1],
			Size:     10240,
		}, []gpt.Partition{
			{FirstLBA: 50, LastLBA: 51,
				Id: testUuids[1], Type: gpt.PartType(testUuids[3])},
			{FirstLBA: 61, LastLBA: 70,
				Id: testUuids[3], Type: gpt.PartType(testUuids[1])},
		}, false, ""},

		{gpt.Table{
			SectorSize: 1024,
			Header: gpt.Header{
				PartitionsArrLen: 2,
				FirstUsableLBA:   5,
				LastUsableLBA:    70,
			},
			Partitions: []gpt.Partition{
				{FirstLBA: 50, LastLBA: 51,
					Id: testUuids[1], Type: gpt.PartType(testUuids[3])},
				{FirstLBA: 0, LastLBA: 0,
					Id: testUuids[0], Type: gpt.PartType(testUuids[0])},
			},
		}, pb.FlashingConfig_Partition{
			PartUuid: "walrus",
			GptType:  testUuidStrings[1],
			Size:     10240,
		}, []gpt.Partition{
			{FirstLBA: 50, LastLBA: 51,
				Id: testUuids[1], Type: gpt.PartType(testUuids[3])},
			{FirstLBA: 0, LastLBA: 0,
				Id: testUuids[0], Type: gpt.PartType(testUuids[0])},
		}, true, "guid"},

		{gpt.Table{
			SectorSize: 1024,
			Header: gpt.Header{
				PartitionsArrLen: 2,
				FirstUsableLBA:   5,
				LastUsableLBA:    70,
			},
			Partitions: []gpt.Partition{
				{FirstLBA: 50, LastLBA: 51,
					Id: testUuids[1], Type: gpt.PartType(testUuids[3])},
				{FirstLBA: 0, LastLBA: 0,
					Id: testUuids[0], Type: gpt.PartType(testUuids[0])},
			},
		}, pb.FlashingConfig_Partition{
			PartUuid: testUuidStrings[3],
			GptType:  "walrus",
			Size:     10240,
		}, []gpt.Partition{
			{FirstLBA: 50, LastLBA: 51,
				Id: testUuids[1], Type: gpt.PartType(testUuids[3])},
			{FirstLBA: 0, LastLBA: 0,
				Id: testUuids[0], Type: gpt.PartType(testUuids[0])},
		}, true, "guid"},

		{gpt.Table{
			SectorSize: 512,
			Header: gpt.Header{
				PartitionsArrLen: 2,
				FirstUsableLBA:   5,
				LastUsableLBA:    70,
			},
			Partitions: []gpt.Partition{
				{FirstLBA: 50, LastLBA: 51,
					Id: testUuids[1], Type: gpt.PartType(testUuids[3])},
				{FirstLBA: 0, LastLBA: 0,
					Id: testUuids[0], Type: gpt.PartType(testUuids[0])},
			},
		}, pb.FlashingConfig_Partition{
			PartUuid: testUuidStrings[1],
			GptType:  testUuidStrings[3],
			Size:     1024,
		}, []gpt.Partition{
			{FirstLBA: 50, LastLBA: 51,
				Id: testUuids[1], Type: gpt.PartType(testUuids[3])},
			{FirstLBA: 0, LastLBA: 0,
				Id: testUuids[0], Type: gpt.PartType(testUuids[0])},
		}, false, ""},
	}
	for i, c := range cases {
		e := AddPersistentIfMissing(&c.table, &c.toAdd)

		if c.shouldFail {
			if e == nil {
				t.Errorf("Test case %v: Excpected error, but none occured", i)
			} else if !strings.Contains(e.Error(), c.shouldContain) {
				t.Errorf("Test case %v: Excpected error to contain %v, but it didn't. "+
					"Instead error was: %v", i, c.shouldContain, e)
			}
		} else {
			if e != nil {
				t.Errorf("Test case %v: Excpected no error, but got: %v", i, e)
			}
		}

		if len(c.table.Partitions) == len(c.remaining) {
			for j, p := range c.table.Partitions {
				if exp := c.remaining[j]; !matchesEnough(&exp, &p) {
					t.Errorf("Test case %v: Partition %v does not match expected. "+
						"Expected/acutual:\n%+v\n%+v", i, j, exp, p)
				}
			}
		} else {
			t.Errorf("Test case %v: Wrong ammount of remaining partitions. "+
				"Expected %v, got %v.", i, len(c.remaining), len(c.table.Partitions))
		}
	}
}
