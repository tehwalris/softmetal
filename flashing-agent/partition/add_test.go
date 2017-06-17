package partition

import (
	"strings"
	"testing"

	"github.com/rekby/gpt"
)

func TestAdd(t *testing.T) {
	var cases = []struct {
		table         gpt.Table
		toAdd         gpt.Partition
		remaining     []gpt.Partition
		shouldFail    bool
		shouldContain string
	}{
		{gpt.Table{
			Header: gpt.Header{PartitionsArrLen: 2},
			Partitions: []gpt.Partition{
				{FirstLBA: 50, LastLBA: 51,
					Id: testUuids[1], Type: gpt.PartType(testUuids[0])},
				{FirstLBA: 50, LastLBA: 51,
					Id: testUuids[1], Type: gpt.PartType(testUuids[1])},
			},
		}, gpt.Partition{
			FirstLBA: 20, LastLBA: 30,
			Id: testUuids[2], Type: gpt.PartType(testUuids[3]),
		}, []gpt.Partition{
			{FirstLBA: 20, LastLBA: 30,
				Id: testUuids[2], Type: gpt.PartType(testUuids[3])},
			{FirstLBA: 50, LastLBA: 51,
				Id: testUuids[1], Type: gpt.PartType(testUuids[1])},
		}, false, ""},

		{gpt.Table{
			Header: gpt.Header{PartitionsArrLen: 2},
			Partitions: []gpt.Partition{
				{FirstLBA: 50, LastLBA: 51,
					Id: testUuids[1], Type: gpt.PartType(testUuids[1])},
				{FirstLBA: 50, LastLBA: 51,
					Id: testUuids[1], Type: gpt.PartType(testUuids[0])},
			},
		}, gpt.Partition{
			FirstLBA: 20, LastLBA: 30,
			Id: testUuids[2], Type: gpt.PartType(testUuids[3]),
		}, []gpt.Partition{
			{FirstLBA: 50, LastLBA: 51,
				Id: testUuids[1], Type: gpt.PartType(testUuids[1])},
			{FirstLBA: 20, LastLBA: 30,
				Id: testUuids[2], Type: gpt.PartType(testUuids[3])},
		}, false, ""},

		{gpt.Table{
			Header: gpt.Header{PartitionsArrLen: 2},
			Partitions: []gpt.Partition{
				{FirstLBA: 50, LastLBA: 51,
					Id: testUuids[1], Type: gpt.PartType(testUuids[1])},
				{FirstLBA: 60, LastLBA: 61,
					Id: testUuids[2], Type: gpt.PartType(testUuids[1])},
			},
		}, gpt.Partition{
			FirstLBA: 20, LastLBA: 30,
			Id: testUuids[2], Type: gpt.PartType(testUuids[3]),
		}, []gpt.Partition{
			{FirstLBA: 50, LastLBA: 51,
				Id: testUuids[1], Type: gpt.PartType(testUuids[1])},
			{FirstLBA: 60, LastLBA: 61,
				Id: testUuids[2], Type: gpt.PartType(testUuids[1])},
		}, true, "No space left"},

		{gpt.Table{
			Header: gpt.Header{PartitionsArrLen: 2},
			Partitions: []gpt.Partition{
				{FirstLBA: 50, LastLBA: 51,
					Id: testUuids[1], Type: gpt.PartType(testUuids[1])},
				{FirstLBA: 60, LastLBA: 61,
					Id: testUuids[1], Type: gpt.PartType(testUuids[0])},
			},
		}, gpt.Partition{
			FirstLBA: 20, LastLBA: 30,
			Id: testUuids[2], Type: gpt.PartType(testUuids[0]),
		}, []gpt.Partition{
			{FirstLBA: 50, LastLBA: 51,
				Id: testUuids[1], Type: gpt.PartType(testUuids[1])},
			{FirstLBA: 60, LastLBA: 61,
				Id: testUuids[1], Type: gpt.PartType(testUuids[0])},
		}, true, "empty partition"},
	}
	for i, c := range cases {
		partitionArrayLenBefore := c.table.Header.PartitionsArrLen
		e := Add(&c.table, &c.toAdd)
		partitionArrayLenAfter := c.table.Header.PartitionsArrLen

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
