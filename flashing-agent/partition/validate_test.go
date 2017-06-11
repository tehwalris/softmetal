package partition

import (
	"strings"
	"testing"

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
	var testUuids = []gpt.Guid{
		{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		{0x1f, 0xe6, 0x90, 0x41, 0xfc, 0xda, 0xb9, 0x4d, 0x83, 0x21, 0xa5, 0xc9, 0x28, 0x47, 0xf7, 0x6b},
		{0x8f, 0x06, 0xf3, 0x1b, 0x1a, 0xff, 0xe5, 0x43, 0xa2, 0xf1, 0x56, 0x39, 0x59, 0x6e, 0xd2, 0xdd},
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
				{FirstLBA: 50, LastLBA: 50, Id: testUuids[1]},
			},
		}, false, ""},
		{gpt.Table{ // Reversed first/last LBA
			Header:     typicalHeader,
			SectorSize: 512,
			Partitions: []gpt.Partition{
				{FirstLBA: 50, LastLBA: 30, Id: testUuids[1]},
			},
		}, true, "LBA"},
		{gpt.Table{ // Maximum size
			Header:     typicalHeader,
			SectorSize: 512,
			Partitions: []gpt.Partition{
				{FirstLBA: 5, LastLBA: 200, Id: testUuids[1]},
			},
		}, false, ""},
		{gpt.Table{ // Too big (last LBA)
			Header:     typicalHeader,
			SectorSize: 512,
			Partitions: []gpt.Partition{
				{FirstLBA: 5, LastLBA: 201, Id: testUuids[1]},
			},
		}, true, "LBA"},
		{gpt.Table{ // Too big (first LBA)
			Header:     typicalHeader,
			SectorSize: 512,
			Partitions: []gpt.Partition{
				{FirstLBA: 4, LastLBA: 200, Id: testUuids[1]},
			},
		}, true, "LBA"},
		{gpt.Table{ // Adjancent
			Header:     typicalHeader,
			SectorSize: 512,
			Partitions: []gpt.Partition{
				{FirstLBA: 30, LastLBA: 40, Id: testUuids[1]},
				{FirstLBA: 41, LastLBA: 50, Id: testUuids[2]},
			},
		}, false, ""},
		{gpt.Table{ // Adjancent, unordered
			Header:     typicalHeader,
			SectorSize: 512,
			Partitions: []gpt.Partition{
				{FirstLBA: 41, LastLBA: 50, Id: testUuids[1]},
				{FirstLBA: 30, LastLBA: 40, Id: testUuids[2]},
			},
		}, false, ""},
		{gpt.Table{ // Two, with space between
			Header:     typicalHeader,
			SectorSize: 512,
			Partitions: []gpt.Partition{
				{FirstLBA: 30, LastLBA: 40, Id: testUuids[1]},
				{FirstLBA: 51, LastLBA: 60, Id: testUuids[2]},
			},
		}, false, ""},
		{gpt.Table{ // Just overlapping
			Header:     typicalHeader,
			SectorSize: 512,
			Partitions: []gpt.Partition{
				{FirstLBA: 30, LastLBA: 40, Id: testUuids[1]},
				{FirstLBA: 40, LastLBA: 50, Id: testUuids[2]},
			},
		}, true, "overlap"},
		{gpt.Table{ // Overlapping, unordered
			Header:     typicalHeader,
			SectorSize: 512,
			Partitions: []gpt.Partition{
				{FirstLBA: 40, LastLBA: 50, Id: testUuids[1]},
				{FirstLBA: 30, LastLBA: 40, Id: testUuids[2]},
			},
		}, true, "overlap"},
		{gpt.Table{ // Overlapping more
			Header:     typicalHeader,
			SectorSize: 512,
			Partitions: []gpt.Partition{
				{FirstLBA: 30, LastLBA: 40, Id: testUuids[1]},
				{FirstLBA: 35, LastLBA: 45, Id: testUuids[2]},
			},
		}, true, "overlap"},
		{gpt.Table{ // Overlapping, contained
			Header:     typicalHeader,
			SectorSize: 512,
			Partitions: []gpt.Partition{
				{FirstLBA: 30, LastLBA: 40, Id: testUuids[1]},
				{FirstLBA: 35, LastLBA: 36, Id: testUuids[2]},
			},
		}, true, "overlap"},
		{gpt.Table{ // Zero id
			Header:     typicalHeader,
			SectorSize: 512,
			Partitions: []gpt.Partition{
				{FirstLBA: 30, LastLBA: 40, Id: testUuids[0]},
				{FirstLBA: 41, LastLBA: 50, Id: testUuids[1]},
			},
		}, false, ""},
		{gpt.Table{ // Zero id
			Header:     typicalHeader,
			SectorSize: 512,
			Partitions: []gpt.Partition{
				{FirstLBA: 30, LastLBA: 40, Id: testUuids[0]},
			},
		}, false, ""},
		{gpt.Table{ // Duplicate ids
			Header:     typicalHeader,
			SectorSize: 512,
			Partitions: []gpt.Partition{
				{FirstLBA: 30, LastLBA: 40, Id: testUuids[1]},
				{FirstLBA: 41, LastLBA: 50, Id: testUuids[1]},
			},
		}, true, "unique"},
		{gpt.Table{ // Duplicate zero ids
			Header:     typicalHeader,
			SectorSize: 512,
			Partitions: []gpt.Partition{
				{FirstLBA: 30, LastLBA: 40, Id: testUuids[0]},
				{FirstLBA: 41, LastLBA: 50, Id: testUuids[0]},
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
