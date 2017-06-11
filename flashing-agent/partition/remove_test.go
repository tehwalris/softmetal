package partition

import (
	"strings"
	"testing"

	pb "git.dolansoft.org/philippe/softmetal/pb"

	"github.com/rekby/gpt"
)

func matchesEnough(a *gpt.Partition, b *gpt.Partition) bool {
	return a.FirstLBA == b.FirstLBA &&
		a.LastLBA == b.LastLBA &&
		a.Id == b.Id
}

func TestRemove(t *testing.T) {
	var cases = []struct {
		table         gpt.Table
		target        pb.FlashingConfig_Partition
		shouldRemove  bool
		remaining     []gpt.Partition
		shouldFail    bool
		shouldContain string
	}{
		{gpt.Table{
			SectorSize: 512,
			Header: gpt.Header{
				PartitionsArrLen: 1,
			},
			Partitions: []gpt.Partition{
				{
					FirstLBA: 50,
					LastLBA:  51,
					Id:       testUuids[1],
					Type:     gpt.PartType(testUuids[2]),
				},
			},
		}, pb.FlashingConfig_Partition{
			PartUuid: testUuidStrings[1],
			GptType:  testUuidStrings[2],
			Size:     1024,
		}, true, []gpt.Partition{}, false, ""},

		{gpt.Table{
			SectorSize: 512,
			Header: gpt.Header{
				PartitionsArrLen: 1,
			},
			Partitions: []gpt.Partition{
				{
					FirstLBA: 50,
					LastLBA:  50,
					Id:       testUuids[1],
					Type:     gpt.PartType(testUuids[2]),
				},
			},
		}, pb.FlashingConfig_Partition{
			PartUuid: testUuidStrings[1],
			GptType:  testUuidStrings[2],
			Size:     1024,
		}, false, []gpt.Partition{
			{
				FirstLBA: 50,
				LastLBA:  50,
				Id:       testUuids[1],
				Type:     gpt.PartType(testUuids[2]),
			},
		}, true, "Refusing to remove"}, // Size mismatch

		{gpt.Table{
			SectorSize: 512,
			Header: gpt.Header{
				PartitionsArrLen: 1,
			},
			Partitions: []gpt.Partition{
				{
					FirstLBA: 50,
					LastLBA:  50,
					Id:       testUuids[1],
					Type:     gpt.PartType(testUuids[2]),
				},
			},
		}, pb.FlashingConfig_Partition{
			PartUuid: testUuidStrings[1],
			GptType:  testUuidStrings[3],
			Size:     512,
		}, false, []gpt.Partition{
			{
				FirstLBA: 50,
				LastLBA:  50,
				Id:       testUuids[1],
				Type:     gpt.PartType(testUuids[2]),
			},
		}, true, "Refusing to remove"}, // Type mismatch

		{gpt.Table{
			SectorSize: 512,
			Header: gpt.Header{
				PartitionsArrLen: 1,
			},
			Partitions: []gpt.Partition{
				{
					FirstLBA: 50,
					LastLBA:  51,
					Id:       testUuids[1],
					Type:     gpt.PartType(testUuids[1]),
				},
			},
		}, pb.FlashingConfig_Partition{
			PartUuid: testUuidStrings[2],
			GptType:  testUuidStrings[1],
			Size:     1024,
		}, false, []gpt.Partition{
			{
				FirstLBA: 50,
				LastLBA:  51,
				Id:       testUuids[1],
				Type:     gpt.PartType(testUuids[1]),
			},
		}, false, ""},

		{gpt.Table{
			SectorSize: 512,
			Header: gpt.Header{
				PartitionsArrLen: 0,
			},
			Partitions: []gpt.Partition{},
		}, pb.FlashingConfig_Partition{
			PartUuid: testUuidStrings[2],
			GptType:  testUuidStrings[1],
			Size:     512,
		}, false, []gpt.Partition{}, false, ""},

		{gpt.Table{
			SectorSize: 512,
			Header: gpt.Header{
				PartitionsArrLen: 2,
			},
			Partitions: []gpt.Partition{
				{
					FirstLBA: 50,
					LastLBA:  50,
					Id:       testUuids[0],
					Type:     gpt.PartType(testUuids[0]),
				},
				{
					FirstLBA: 51,
					LastLBA:  51,
					Id:       testUuids[1],
					Type:     gpt.PartType(testUuids[0]),
				},
			},
		}, pb.FlashingConfig_Partition{
			PartUuid: testUuidStrings[1],
			GptType:  testUuidStrings[0],
			Size:     512,
		}, true, []gpt.Partition{
			{FirstLBA: 50, LastLBA: 50, Id: testUuids[0]},
		}, false, ""},
	}
	for i, c := range cases {
		didRemove, err := Remove(&c.table, &c.target)
		if c.shouldRemove {
			if !didRemove {
				t.Errorf("Test case %v: Did not get excpected removal confirmation.", i)
			}
		} else {
			if didRemove {
				t.Errorf("Test case %v: Got unexpected removal confirmation.", i)
			}
		}
		if len(c.table.Partitions) == len(c.remaining) {
			for j, p := range c.table.Partitions {
				if exp := c.remaining[j]; !matchesEnough(&exp, &p) {
					t.Errorf("Test case %v: Partition %v does not match expected."+
						"Expected/acutual:\n%+v\n%+v", i, j, exp, p)
				}
			}
		} else {
			t.Errorf("Test case %v: Wrong ammount of remaining partitions."+
				"Expected %v, got %v.", i, len(c.remaining), len(c.table.Partitions))
		}
		if hdr := c.table.Header.PartitionsArrLen; uint32(len(c.remaining)) != hdr {
			t.Errorf("Test case %v: Header.PartitionsArrLen does not match remaining partition count. "+
				"Header/actual remaining: %v/%v", i, hdr, len(c.remaining))
		}
		if c.shouldFail {
			if err == nil {
				t.Errorf("Test case %v: Expected error, but got none.", i)
			} else if !strings.Contains(err.Error(), c.shouldContain) {
				t.Errorf("Test case %v: Expected error to contain %v, but it didn't. "+
					"Instead error was: %v", i, c.shouldContain, err)
			}
		} else {
			if err != nil {
				t.Errorf("Test case %v: Excpected no error, but got: %v", i, err)
			}
		}
	}
}
