package partition

import (
	"testing"

	"github.com/rekby/gpt"
)

func matchesEnough(a *gpt.Partition, b *gpt.Partition) bool {
	return a.FirstLBA == b.FirstLBA &&
		a.LastLBA == b.LastLBA &&
		a.Id == b.Id &&
		a.Type == b.Type
}

func TestRemove(t *testing.T) {
	var cases = []struct {
		table        gpt.Table
		targetUuid   string
		shouldRemove bool
		remaining    []gpt.Partition
	}{
		{gpt.Table{
			Header: gpt.Header{PartitionsArrLen: 1},
			Partitions: []gpt.Partition{
				{FirstLBA: 50, LastLBA: 51,
					Id: testUuids[1], Type: gpt.PartType(testUuids[1])},
			},
		}, testUuidStrings[1], true, []gpt.Partition{
			{FirstLBA: 0, LastLBA: 0,
				Id: testUuids[0], Type: gpt.PartType(testUuids[0])},
		}},

		{gpt.Table{
			Header: gpt.Header{PartitionsArrLen: 1},
			Partitions: []gpt.Partition{
				{FirstLBA: 50, LastLBA: 51,
					Id: testUuids[1], Type: gpt.PartType(testUuids[1])},
			},
		}, testUuidStrings[2], false, []gpt.Partition{
			{FirstLBA: 50, LastLBA: 51,
				Id: testUuids[1], Type: gpt.PartType(testUuids[1])},
		}},

		{gpt.Table{
			Header:     gpt.Header{PartitionsArrLen: 0},
			Partitions: []gpt.Partition{},
		}, testUuidStrings[2], false, []gpt.Partition{}},

		{gpt.Table{
			Header:     gpt.Header{PartitionsArrLen: 0},
			Partitions: []gpt.Partition{},
		}, testUuidStrings[0], false, []gpt.Partition{}},

		{gpt.Table{
			Header: gpt.Header{PartitionsArrLen: 2},
			Partitions: []gpt.Partition{
				{FirstLBA: 50, LastLBA: 50,
					Id: testUuids[1], Type: gpt.PartType(testUuids[1])},
				{FirstLBA: 51, LastLBA: 51,
					Id: testUuids[2], Type: gpt.PartType(testUuids[1])},
			},
		}, testUuidStrings[1], true, []gpt.Partition{
			{FirstLBA: 0, LastLBA: 0,
				Id: testUuids[0], Type: gpt.PartType(testUuids[0])},
			{FirstLBA: 51, LastLBA: 51,
				Id: testUuids[2], Type: gpt.PartType(testUuids[1])},
		}},

		{gpt.Table{
			Header: gpt.Header{PartitionsArrLen: 3},
			Partitions: []gpt.Partition{
				{FirstLBA: 30, LastLBA: 30,
					Id: testUuids[1], Type: gpt.PartType(testUuids[0])},
				{FirstLBA: 50, LastLBA: 50,
					Id: testUuids[1], Type: gpt.PartType(testUuids[1])},
				{FirstLBA: 51, LastLBA: 51,
					Id: testUuids[2], Type: gpt.PartType(testUuids[1])},
			},
		}, testUuidStrings[1], true, []gpt.Partition{
			{FirstLBA: 30, LastLBA: 30,
				Id: testUuids[1], Type: gpt.PartType(testUuids[0])},
			{FirstLBA: 0, LastLBA: 0,
				Id: testUuids[0], Type: gpt.PartType(testUuids[0])},
			{FirstLBA: 51, LastLBA: 51,
				Id: testUuids[2], Type: gpt.PartType(testUuids[1])},
		}},
	}
	for i, c := range cases {
		partitionArrayLenBefore := c.table.Header.PartitionsArrLen
		didRemove := Remove(&c.table, &c.targetUuid)
		partitionArrayLenAfter := c.table.Header.PartitionsArrLen

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

		if partitionArrayLenBefore != partitionArrayLenAfter {
			t.Errorf("Test case %v: Unexpected change of PartitionsArrLen (%v -> %v).",
				i, partitionArrayLenBefore, partitionArrayLenAfter)
		}
	}
}
