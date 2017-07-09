package partition

import (
	"strings"
	"testing"

	pb "git.dolansoft.org/philippe/softmetal/pb"

	"github.com/rekby/gpt"
)

func TestAssertGptCompatible(t *testing.T) {
	var cases = []struct {
		disk          gpt.Table
		img           gpt.Table
		shouldFail    bool
		shouldContain string
	}{
		{gpt.Table{
			SectorSize: 512,
			Header:     gpt.Header{},
			Partitions: []gpt.Partition{},
		}, gpt.Table{
			SectorSize: 1024,
			Header:     gpt.Header{},
			Partitions: []gpt.Partition{},
		}, true, "sector"},
		{gpt.Table{
			SectorSize: 1024,
			Header:     gpt.Header{},
			Partitions: []gpt.Partition{},
		}, gpt.Table{
			SectorSize: 1024,
			Header:     gpt.Header{},
			Partitions: []gpt.Partition{},
		}, false, ""},
		{gpt.Table{
			SectorSize: 512,
			Header:     gpt.Header{},
			Partitions: []gpt.Partition{
				gpt.Partition{},
			},
		}, gpt.Table{
			SectorSize: 512,
			Header:     gpt.Header{},
			Partitions: []gpt.Partition{},
		}, false, ""},
	}
	for i, c := range cases {
		e := AssertGptCompatible(&c.disk, &c.img)
		if c.shouldFail {
			if e == nil {
				t.Errorf("Test case %v: Excpected error, but none occured", i)
			} else if !strings.Contains(e.Error(), c.shouldContain) {
				t.Errorf("Test case %v: Excpected error to contain %v, but it didn't."+
					"Instead error was: %v", i, c.shouldContain, e)
			}
		} else {
			if e != nil {
				t.Errorf("Test case %v: Excpected no error, but got: %v", i, e)
			}
		}
	}
}

func TestAssertGptValid(t *testing.T) {
	var typicalHeader = gpt.Header{
		FirstUsableLBA: 5,
		LastUsableLBA:  200,
	}
	var cases = []struct {
		table         gpt.Table
		shouldFail    bool
		shouldContain string
	}{
		{gpt.Table{ // Control
			Header:     typicalHeader,
			SectorSize: 512,
			Partitions: []gpt.Partition{
				{FirstLBA: 50, LastLBA: 50,
					Id: testUuids[1], Type: gpt.PartType(testUuids[1])},
			},
		}, false, ""},
		{gpt.Table{ // Reversed first/last LBA
			Header:     typicalHeader,
			SectorSize: 512,
			Partitions: []gpt.Partition{
				{FirstLBA: 50, LastLBA: 30,
					Id: testUuids[1], Type: gpt.PartType(testUuids[1])},
			},
		}, true, "LBA"},
		{gpt.Table{ // Reversed first/last LBA, but "empty" partition
			Header:     typicalHeader,
			SectorSize: 512,
			Partitions: []gpt.Partition{
				{FirstLBA: 50, LastLBA: 30,
					Id: testUuids[1], Type: gpt.PartType(testUuids[0])},
			},
		}, false, ""},
		{gpt.Table{ // Maximum size
			Header:     typicalHeader,
			SectorSize: 512,
			Partitions: []gpt.Partition{
				{FirstLBA: 5, LastLBA: 200,
					Id: testUuids[1], Type: gpt.PartType(testUuids[1])},
			},
		}, false, ""},
		{gpt.Table{ // Too big (last LBA)
			Header:     typicalHeader,
			SectorSize: 512,
			Partitions: []gpt.Partition{
				{FirstLBA: 5, LastLBA: 201,
					Id: testUuids[1], Type: gpt.PartType(testUuids[1])},
			},
		}, true, "LBA"},
		{gpt.Table{ // Too big (first LBA)
			Header:     typicalHeader,
			SectorSize: 512,
			Partitions: []gpt.Partition{
				{FirstLBA: 4, LastLBA: 200,
					Id: testUuids[1], Type: gpt.PartType(testUuids[1])},
			},
		}, true, "LBA"},
		{gpt.Table{ // Adjancent
			Header:     typicalHeader,
			SectorSize: 512,
			Partitions: []gpt.Partition{
				{FirstLBA: 30, LastLBA: 40,
					Id: testUuids[1], Type: gpt.PartType(testUuids[1])},
				{FirstLBA: 41, LastLBA: 50,
					Id: testUuids[2], Type: gpt.PartType(testUuids[1])},
			},
		}, false, ""},
		{gpt.Table{ // Adjancent, unordered
			Header:     typicalHeader,
			SectorSize: 512,
			Partitions: []gpt.Partition{
				{FirstLBA: 41, LastLBA: 50,
					Id: testUuids[1], Type: gpt.PartType(testUuids[1])},
				{FirstLBA: 30, LastLBA: 40,
					Id: testUuids[2], Type: gpt.PartType(testUuids[1])},
			},
		}, false, ""},
		{gpt.Table{ // Two, with space between
			Header:     typicalHeader,
			SectorSize: 512,
			Partitions: []gpt.Partition{
				{FirstLBA: 30, LastLBA: 40,
					Id: testUuids[1], Type: gpt.PartType(testUuids[1])},
				{FirstLBA: 51, LastLBA: 60,
					Id: testUuids[2], Type: gpt.PartType(testUuids[1])},
			},
		}, false, ""},
		{gpt.Table{ // Just overlapping
			Header:     typicalHeader,
			SectorSize: 512,
			Partitions: []gpt.Partition{
				{FirstLBA: 30, LastLBA: 40,
					Id: testUuids[1], Type: gpt.PartType(testUuids[1])},
				{FirstLBA: 40, LastLBA: 50,
					Id: testUuids[2], Type: gpt.PartType(testUuids[1])},
			},
		}, true, "overlap"},
		{gpt.Table{ // Just overlapping, reversed
			Header:     typicalHeader,
			SectorSize: 512,
			Partitions: []gpt.Partition{
				{FirstLBA: 40, LastLBA: 50,
					Id: testUuids[2], Type: gpt.PartType(testUuids[1])},
				{FirstLBA: 30, LastLBA: 40,
					Id: testUuids[1], Type: gpt.PartType(testUuids[1])},
			},
		}, true, "overlap"},
		{gpt.Table{ // Overlapping, but one "empty" partition
			Header:     typicalHeader,
			SectorSize: 512,
			Partitions: []gpt.Partition{
				{FirstLBA: 30, LastLBA: 40,
					Id: testUuids[1], Type: gpt.PartType(testUuids[1])},
				{FirstLBA: 40, LastLBA: 50,
					Id: testUuids[2], Type: gpt.PartType(testUuids[0])},
			},
		}, false, ""},
		{gpt.Table{ // Overlapping, but one "empty" partition, reversed order
			Header:     typicalHeader,
			SectorSize: 512,
			Partitions: []gpt.Partition{
				{FirstLBA: 40, LastLBA: 50,
					Id: testUuids[2], Type: gpt.PartType(testUuids[0])},
				{FirstLBA: 30, LastLBA: 40,
					Id: testUuids[1], Type: gpt.PartType(testUuids[1])},
			},
		}, false, ""},
		{gpt.Table{ // Overlapping, exact overlap, one "empty" partition
			Header:     typicalHeader,
			SectorSize: 512,
			Partitions: []gpt.Partition{
				{FirstLBA: 35, LastLBA: 35,
					Id: testUuids[2], Type: gpt.PartType(testUuids[0])},
				{FirstLBA: 35, LastLBA: 35,
					Id: testUuids[1], Type: gpt.PartType(testUuids[1])},
			},
		}, false, ""},
		{gpt.Table{ // Overlapping, exact overlap, one "empty" partition, reversed
			Header:     typicalHeader,
			SectorSize: 512,
			Partitions: []gpt.Partition{
				{FirstLBA: 35, LastLBA: 35,
					Id: testUuids[2], Type: gpt.PartType(testUuids[1])},
				{FirstLBA: 35, LastLBA: 35,
					Id: testUuids[1], Type: gpt.PartType(testUuids[0])},
			},
		}, false, ""},
		{gpt.Table{ // Overlapping, unordered
			Header:     typicalHeader,
			SectorSize: 512,
			Partitions: []gpt.Partition{
				{FirstLBA: 40, LastLBA: 50,
					Id: testUuids[1], Type: gpt.PartType(testUuids[1])},
				{FirstLBA: 30, LastLBA: 40,
					Id: testUuids[2], Type: gpt.PartType(testUuids[1])},
			},
		}, true, "overlap"},
		{gpt.Table{ // Overlapping more
			Header:     typicalHeader,
			SectorSize: 512,
			Partitions: []gpt.Partition{
				{FirstLBA: 30, LastLBA: 40,
					Id: testUuids[1], Type: gpt.PartType(testUuids[1])},
				{FirstLBA: 35, LastLBA: 45,
					Id: testUuids[2], Type: gpt.PartType(testUuids[1])},
			},
		}, true, "overlap"},
		{gpt.Table{ // Overlapping, contained
			Header:     typicalHeader,
			SectorSize: 512,
			Partitions: []gpt.Partition{
				{FirstLBA: 30, LastLBA: 40,
					Id: testUuids[1], Type: gpt.PartType(testUuids[1])},
				{FirstLBA: 35, LastLBA: 36,
					Id: testUuids[2], Type: gpt.PartType(testUuids[1])},
			},
		}, true, "overlap"},
		{gpt.Table{ // Zero id
			Header:     typicalHeader,
			SectorSize: 512,
			Partitions: []gpt.Partition{
				{FirstLBA: 30, LastLBA: 40,
					Id: testUuids[0], Type: gpt.PartType(testUuids[1])},
				{FirstLBA: 41, LastLBA: 50,
					Id: testUuids[1], Type: gpt.PartType(testUuids[1])},
			},
		}, false, ""},
		{gpt.Table{ // Zero id
			Header:     typicalHeader,
			SectorSize: 512,
			Partitions: []gpt.Partition{
				{FirstLBA: 30, LastLBA: 40,
					Id: testUuids[0], Type: gpt.PartType(testUuids[1])},
			},
		}, false, ""},
		{gpt.Table{ // Duplicate ids
			Header:     typicalHeader,
			SectorSize: 512,
			Partitions: []gpt.Partition{
				{FirstLBA: 30, LastLBA: 40,
					Id: testUuids[1], Type: gpt.PartType(testUuids[1])},
				{FirstLBA: 41, LastLBA: 50,
					Id: testUuids[1], Type: gpt.PartType(testUuids[1])},
			},
		}, true, "unique"},
		{gpt.Table{ // Duplicate ids, but "empty" partitions
			Header:     typicalHeader,
			SectorSize: 512,
			Partitions: []gpt.Partition{
				{FirstLBA: 30, LastLBA: 40,
					Id: testUuids[1], Type: gpt.PartType(testUuids[0])},
				{FirstLBA: 41, LastLBA: 50,
					Id: testUuids[1], Type: gpt.PartType(testUuids[0])},
			},
		}, false, ""},
		{gpt.Table{ // Duplicate zero ids
			Header:     typicalHeader,
			SectorSize: 512,
			Partitions: []gpt.Partition{
				{FirstLBA: 30, LastLBA: 40,
					Id: testUuids[0], Type: gpt.PartType(testUuids[1])},
				{FirstLBA: 41, LastLBA: 50,
					Id: testUuids[0], Type: gpt.PartType(testUuids[1])},
			},
		}, true, "unique"},
	}

	for i, c := range cases {
		e := AssertGptValid(&c.table)
		if c.shouldFail {
			if e == nil {
				t.Errorf("Test case %v: Excpected error, but none occured", i)
			} else if !strings.Contains(e.Error(), c.shouldContain) {
				t.Errorf("Test case %v: Excpected error to contain %v, but it didn't."+
					"Instead error was: %v", i, c.shouldContain, e)
			}
		} else {
			if e != nil {
				t.Errorf("Test case %v: Excpected no error, but got: %v", i, e)
			}
		}
	}
}

func TestAssertExactMatchIfExists(t *testing.T) {
	var cases = []struct {
		table         gpt.Table
		target        pb.FlashingConfig_Partition
		shouldFail    bool
		shouldContain string
	}{
		{gpt.Table{
			SectorSize: 512,
			Partitions: []gpt.Partition{},
		}, pb.FlashingConfig_Partition{
			PartUuid: testUuidStrings[1],
			GptType:  testUuidStrings[1],
			Size:     10,
		}, false, ""},

		{gpt.Table{
			SectorSize: 512,
			Partitions: []gpt.Partition{
				{FirstLBA: 30, LastLBA: 40,
					Id: testUuids[2], Type: gpt.PartType(testUuids[1])},
			},
		}, pb.FlashingConfig_Partition{
			PartUuid: testUuidStrings[1],
			GptType:  testUuidStrings[1],
			Size:     10,
		}, false, ""},

		{gpt.Table{
			SectorSize: 512,
			Partitions: []gpt.Partition{
				{FirstLBA: 30, LastLBA: 40,
					Id: testUuids[1], Type: gpt.PartType(testUuids[1])},
			},
		}, pb.FlashingConfig_Partition{
			PartUuid: testUuidStrings[1],
			GptType:  testUuidStrings[1],
			Size:     5120,
		}, true, "not as expected"},

		{gpt.Table{
			SectorSize: 512,
			Partitions: []gpt.Partition{
				{FirstLBA: 31, LastLBA: 40,
					Id: testUuids[1], Type: gpt.PartType(testUuids[1])},
			},
		}, pb.FlashingConfig_Partition{
			PartUuid: testUuidStrings[1],
			GptType:  testUuidStrings[1],
			Size:     5120,
		}, false, ""},

		{gpt.Table{
			SectorSize: 512,
			Partitions: []gpt.Partition{
				{FirstLBA: 31, LastLBA: 40,
					Id: testUuids[1], Type: gpt.PartType(testUuids[2])},
			},
		}, pb.FlashingConfig_Partition{
			PartUuid: testUuidStrings[1],
			GptType:  testUuidStrings[1],
			Size:     5120,
		}, true, "not as expected"},

		{gpt.Table{
			SectorSize: 512,
			Partitions: []gpt.Partition{
				{FirstLBA: 0, LastLBA: 0,
					Id: testUuids[0], Type: gpt.PartType(testUuids[0])},
				{FirstLBA: 31, LastLBA: 40,
					Id: testUuids[1], Type: gpt.PartType(testUuids[2])},
				{FirstLBA: 0, LastLBA: 0,
					Id: testUuids[0], Type: gpt.PartType(testUuids[0])},
			},
		}, pb.FlashingConfig_Partition{
			PartUuid: testUuidStrings[1],
			GptType:  testUuidStrings[1],
			Size:     5120,
		}, true, "not as expected"},

		{gpt.Table{
			SectorSize: 512,
			Partitions: []gpt.Partition{
				{FirstLBA: 0, LastLBA: 0,
					Id: testUuids[0], Type: gpt.PartType(testUuids[0])},
				{FirstLBA: 31, LastLBA: 40,
					Id: testUuids[1], Type: gpt.PartType(testUuids[2])},
				{FirstLBA: 0, LastLBA: 0,
					Id: testUuids[0], Type: gpt.PartType(testUuids[0])},
			},
		}, pb.FlashingConfig_Partition{
			PartUuid: testUuidStrings[3],
			GptType:  testUuidStrings[3],
			Size:     5120,
		}, false, ""},
	}

	for i, c := range cases {
		e := AssertExactMatchIfExists(&c.table, &c.target)
		if c.shouldFail {
			if e == nil {
				t.Errorf("Test case %v: Excpected error, but none occured", i)
			} else if !strings.Contains(e.Error(), c.shouldContain) {
				t.Errorf("Test case %v: Excpected error to contain %v, but it didn't."+
					"Instead error was: %v", i, c.shouldContain, e)
			}
		} else {
			if e != nil {
				t.Errorf("Test case %v: Excpected no error, but got: %v", i, e)
			}
		}
	}
}

func TestAssertPersistentValid(t *testing.T) {
	var cases = []struct {
		partitions    []pb.FlashingConfig_Partition
		shouldFail    bool
		shouldContain string
	}{
		{[]pb.FlashingConfig_Partition{}, false, ""},
		{[]pb.FlashingConfig_Partition{
			{PartUuid: testUuidStrings[1], Size: 10, GptType: testUuidStrings[1]},
			{PartUuid: testUuidStrings[2], Size: 10, GptType: testUuidStrings[1]},
		}, false, ""},
		{[]pb.FlashingConfig_Partition{
			{PartUuid: testUuidStrings[1], Size: 10, GptType: testUuidStrings[1]},
			{PartUuid: testUuidStrings[2], Size: 10, GptType: testUuidStrings[0]},
		}, true, "blank"},
		{[]pb.FlashingConfig_Partition{
			{PartUuid: testUuidStrings[1], Size: 10, GptType: testUuidStrings[0]},
			{PartUuid: testUuidStrings[2], Size: 10, GptType: testUuidStrings[0]},
		}, true, "blank"},
		{[]pb.FlashingConfig_Partition{
			{PartUuid: testUuidStrings[2], Size: 10, GptType: testUuidStrings[1]},
			{PartUuid: testUuidStrings[2], Size: 5, GptType: testUuidStrings[2]},
		}, true, "Duplicate"},
		{[]pb.FlashingConfig_Partition{
			{PartUuid: testUuidStrings[2], Size: 10, GptType: testUuidStrings[1]},
			{PartUuid: testUuidStrings[1], Size: 0, GptType: testUuidStrings[2]},
		}, true, "size 0"},
	}

	for i, c := range cases {
		e := AssertPersistentValid(c.partitions)
		if c.shouldFail {
			if e == nil {
				t.Errorf("Test case %v: Excpected error, but none occured", i)
			} else if !strings.Contains(e.Error(), c.shouldContain) {
				t.Errorf("Test case %v: Excpected error to contain %v, but it didn't."+
					"Instead error was: %v", i, c.shouldContain, e)
			}
		} else {
			if e != nil {
				t.Errorf("Test case %v: Excpected no error, but got: %v", i, e)
			}
		}
	}
}
