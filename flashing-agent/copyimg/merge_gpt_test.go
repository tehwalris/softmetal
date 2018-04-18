package copyimg_test

import (
	"strings"
	"testing"

	"git.dolansoft.org/philippe/softmetal/flashing-agent/copyimg"
	pb "git.dolansoft.org/philippe/softmetal/pb"

	"github.com/rekby/gpt"
)

func TestMergeGpt(t *testing.T) {
	var cases = []struct {
		diskGpt       gpt.Table
		imageGpt      gpt.Table
		persistent    []pb.FlashingConfig_Partition
		shouldFail    bool
		shouldContain string
		remaining     []gpt.Partition
	}{
		{
			gpt.Table{
				SectorSize: 1024,
				Header:     gpt.Header{FirstUsableLBA: 5, LastUsableLBA: 200},
				Partitions: []gpt.Partition{},
			},
			gpt.Table{
				SectorSize: 1024,
				Header:     gpt.Header{FirstUsableLBA: 8, LastUsableLBA: 150},
				Partitions: []gpt.Partition{},
			},
			[]pb.FlashingConfig_Partition{},
			false, "",
			[]gpt.Partition{},
		},
		{
			gpt.Table{
				SectorSize: 1024,
				Header:     gpt.Header{FirstUsableLBA: 5, LastUsableLBA: 200},
				Partitions: []gpt.Partition{},
			},
			gpt.Table{
				SectorSize: 512,
				Header:     gpt.Header{FirstUsableLBA: 8, LastUsableLBA: 150},
				Partitions: []gpt.Partition{},
			},
			[]pb.FlashingConfig_Partition{},
			true, "sector size",
			[]gpt.Partition{},
		},
		{
			gpt.Table{
				SectorSize: 1024,
				Header:     gpt.Header{FirstUsableLBA: 5, LastUsableLBA: 200},
				Partitions: []gpt.Partition{
					{FirstLBA: 30, LastLBA: 40,
						Id: testUuids[1], Type: gpt.PartType(testUuids[1])},
					{FirstLBA: 35, LastLBA: 45,
						Id: testUuids[2], Type: gpt.PartType(testUuids[1])},
				},
			},
			gpt.Table{
				SectorSize: 1024,
				Header:     gpt.Header{FirstUsableLBA: 8, LastUsableLBA: 150},
				Partitions: []gpt.Partition{},
			},
			[]pb.FlashingConfig_Partition{},
			true, "overlap",
			[]gpt.Partition{},
		},
		{
			gpt.Table{
				SectorSize: 1024,
				Header:     gpt.Header{FirstUsableLBA: 5, LastUsableLBA: 200},
				Partitions: []gpt.Partition{},
			},
			gpt.Table{
				SectorSize: 1024,
				Header:     gpt.Header{FirstUsableLBA: 8, LastUsableLBA: 150},
				Partitions: []gpt.Partition{
					{FirstLBA: 30, LastLBA: 40,
						Id: testUuids[1], Type: gpt.PartType(testUuids[1])},
					{FirstLBA: 35, LastLBA: 45,
						Id: testUuids[2], Type: gpt.PartType(testUuids[1])},
				},
			},
			[]pb.FlashingConfig_Partition{},
			true, "overlap",
			[]gpt.Partition{},
		},
		{
			gpt.Table{
				SectorSize: 1024,
				Header:     gpt.Header{FirstUsableLBA: 5, LastUsableLBA: 200},
				Partitions: []gpt.Partition{},
			},
			gpt.Table{
				SectorSize: 1024,
				Header:     gpt.Header{FirstUsableLBA: 8, LastUsableLBA: 150},
				Partitions: []gpt.Partition{},
			},
			[]pb.FlashingConfig_Partition{
				{PartUuid: testUuidStrings[2], Size: 10, GptType: testUuidStrings[1]},
				{PartUuid: testUuidStrings[2], Size: 10, GptType: testUuidStrings[1]},
			},
			true, "Duplicate",
			[]gpt.Partition{},
		},
		{
			gpt.Table{
				SectorSize: 1024,
				Header:     gpt.Header{FirstUsableLBA: 5, LastUsableLBA: 200},
				Partitions: []gpt.Partition{},
			},
			gpt.Table{
				SectorSize: 1024,
				Header:     gpt.Header{FirstUsableLBA: 8, LastUsableLBA: 150},
				Partitions: []gpt.Partition{
					{FirstLBA: 35, LastLBA: 45,
						Id: testUuids[2], Type: gpt.PartType(testUuids[2])},
				},
			},
			[]pb.FlashingConfig_Partition{
				{PartUuid: testUuidStrings[2], Size: 10, GptType: testUuidStrings[1]},
			},
			true, "conflicts",
			[]gpt.Partition{},
		},
		{
			gpt.Table{
				SectorSize: 1024,
				Header:     gpt.Header{FirstUsableLBA: 5, LastUsableLBA: 200},
				Partitions: []gpt.Partition{
					{FirstLBA: 35, LastLBA: 35,
						Id: testUuids[2], Type: gpt.PartType(testUuids[2])},
				},
			},
			gpt.Table{
				SectorSize: 1024,
				Header:     gpt.Header{FirstUsableLBA: 8, LastUsableLBA: 150},
				Partitions: []gpt.Partition{},
			},
			[]pb.FlashingConfig_Partition{
				{PartUuid: testUuidStrings[2], Size: 1024, GptType: testUuidStrings[1]},
			},
			true, "type",
			[]gpt.Partition{},
		},
		{
			gpt.Table{
				SectorSize: 1024,
				Header:     gpt.Header{FirstUsableLBA: 5, LastUsableLBA: 200},
				Partitions: []gpt.Partition{
					{FirstLBA: 35, LastLBA: 45,
						Id: testUuids[2], Type: gpt.PartType(testUuids[1])},
				},
			},
			gpt.Table{
				SectorSize: 1024,
				Header:     gpt.Header{FirstUsableLBA: 8, LastUsableLBA: 150},
				Partitions: []gpt.Partition{},
			},
			[]pb.FlashingConfig_Partition{
				{PartUuid: testUuidStrings[2], Size: 1024, GptType: testUuidStrings[1]},
			},
			true, "size",
			[]gpt.Partition{},
		},
		{
			gpt.Table{
				SectorSize: 1024,
				Header:     gpt.Header{FirstUsableLBA: 5, LastUsableLBA: 200},
				Partitions: []gpt.Partition{
					{FirstLBA: 35, LastLBA: 35,
						Id: testUuids[2], Type: gpt.PartType(testUuids[1])},
				},
			},
			gpt.Table{
				SectorSize: 1024,
				Header:     gpt.Header{FirstUsableLBA: 8, LastUsableLBA: 150},
				Partitions: []gpt.Partition{},
			},
			[]pb.FlashingConfig_Partition{
				{PartUuid: testUuidStrings[2], Size: 1024, GptType: testUuidStrings[1]},
			},
			false, "",
			[]gpt.Partition{
				{FirstLBA: 35, LastLBA: 35,
					Id: testUuids[2], Type: gpt.PartType(testUuids[1])},
			},
		},
		{
			gpt.Table{
				SectorSize: 1024,
				Header:     gpt.Header{FirstUsableLBA: 5, LastUsableLBA: 200},
				Partitions: []gpt.Partition{
					{FirstLBA: 35, LastLBA: 35,
						Id: testUuids[2], Type: gpt.PartType(testUuids[0])},
					{FirstLBA: 35, LastLBA: 35,
						Id: testUuids[2], Type: gpt.PartType(testUuids[1])},
					{FirstLBA: 0, LastLBA: 0,
						Id: testUuids[0], Type: gpt.PartType(testUuids[0])},
					{FirstLBA: 40, LastLBA: 42,
						Id: testUuids[3], Type: gpt.PartType(testUuids[0])},
					{FirstLBA: 40, LastLBA: 42,
						Id: testUuids[3], Type: gpt.PartType(testUuids[2])},
					{FirstLBA: 0, LastLBA: 0,
						Id: testUuids[0], Type: gpt.PartType(testUuids[0])},
				},
			},
			gpt.Table{
				SectorSize: 1024,
				Header:     gpt.Header{FirstUsableLBA: 8, LastUsableLBA: 150},
				Partitions: []gpt.Partition{
					{FirstLBA: 100, LastLBA: 102,
						Id: testUuids[4], Type: gpt.PartType(testUuids[2])},
					{FirstLBA: 10, LastLBA: 20,
						Id: testUuids[3], Type: gpt.PartType(testUuids[1])},
				},
			},
			[]pb.FlashingConfig_Partition{
				{PartUuid: testUuidStrings[2], Size: 1024, GptType: testUuidStrings[1]},
				{PartUuid: testUuidStrings[1], Size: 2048, GptType: testUuidStrings[2]},
			},
			false, "",
			[]gpt.Partition{
				{FirstLBA: 199, LastLBA: 200,
					Id: testUuids[1], Type: gpt.PartType(testUuids[2])},
				{FirstLBA: 35, LastLBA: 35,
					Id: testUuids[2], Type: gpt.PartType(testUuids[1])},
				{FirstLBA: 5, LastLBA: 7,
					Id: testUuids[4], Type: gpt.PartType(testUuids[2])},
				{FirstLBA: 8, LastLBA: 18,
					Id: testUuids[3], Type: gpt.PartType(testUuids[1])},
				{FirstLBA: 0, LastLBA: 0,
					Id: testUuids[0], Type: gpt.PartType(testUuids[0])},
				{FirstLBA: 0, LastLBA: 0,
					Id: testUuids[0], Type: gpt.PartType(testUuids[0])},
			},
		},
	}
	for i, c := range cases {
		e := copyimg.MergeGpt(&c.diskGpt, &c.imageGpt, c.persistent)
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
			} else {
				if len(c.diskGpt.Partitions) == len(c.remaining) {
					for j, p := range c.diskGpt.Partitions {
						if exp := c.remaining[j]; !matchesEnough(&exp, &p) {
							t.Errorf("Test case %v: Partition %v does not match expected. "+
								"Expected/acutual:\n%+v\n%+v", i, j, exp, p)
						}
					}
				} else {
					t.Errorf("Test case %v: Wrong ammount of remaining partitions. "+
						"Expected %v, got %v.", i, len(c.remaining), len(c.diskGpt.Partitions))
				}
			}
		}
	}
}
