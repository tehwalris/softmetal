package partition

import (
	"testing"

	"github.com/rekby/gpt"
)

func TestFindSpace(t *testing.T) {
	var cases = []struct {
		table            gpt.Table
		blocks           uint64
		side             DiskSide
		expectedFirstLBA uint64
		expectedLastLBA  uint64
		expectedFound    bool
	}{
		{gpt.Table{
			Header:     gpt.Header{FirstUsableLBA: 5, LastUsableLBA: 10},
			Partitions: []gpt.Partition{},
		}, 0, Start, 0, 0, false},
		{gpt.Table{
			Header:     gpt.Header{FirstUsableLBA: 5, LastUsableLBA: 10},
			Partitions: []gpt.Partition{},
		}, 0, End, 0, 0, false},
		{gpt.Table{
			Header:     gpt.Header{FirstUsableLBA: 5, LastUsableLBA: 10},
			Partitions: []gpt.Partition{},
		}, 1, Start, 5, 5, true},
		{gpt.Table{
			Header:     gpt.Header{FirstUsableLBA: 5, LastUsableLBA: 10},
			Partitions: []gpt.Partition{},
		}, 1, End, 10, 10, true},
		{gpt.Table{
			Header:     gpt.Header{FirstUsableLBA: 5, LastUsableLBA: 10},
			Partitions: []gpt.Partition{},
		}, 3, Start, 5, 7, true},
		{gpt.Table{
			Header:     gpt.Header{FirstUsableLBA: 5, LastUsableLBA: 10},
			Partitions: []gpt.Partition{},
		}, 3, End, 8, 10, true},
		{gpt.Table{
			Header:     gpt.Header{FirstUsableLBA: 5, LastUsableLBA: 10},
			Partitions: []gpt.Partition{},
		}, 6, Start, 5, 10, true},
		{gpt.Table{
			Header:     gpt.Header{FirstUsableLBA: 5, LastUsableLBA: 10},
			Partitions: []gpt.Partition{},
		}, 6, End, 5, 10, true},
		{gpt.Table{
			Header:     gpt.Header{FirstUsableLBA: 5, LastUsableLBA: 10},
			Partitions: []gpt.Partition{},
		}, 7, Start, 0, 0, false},
		{gpt.Table{
			Header:     gpt.Header{FirstUsableLBA: 5, LastUsableLBA: 10},
			Partitions: []gpt.Partition{},
		}, 7, End, 0, 0, false},
		{gpt.Table{
			Header: gpt.Header{FirstUsableLBA: 5, LastUsableLBA: 10},
			Partitions: []gpt.Partition{
				{FirstLBA: 7, LastLBA: 8, Type: gpt.PartType(testUuids[1])},
			},
		}, 1, Start, 5, 5, true},
		{gpt.Table{
			Header: gpt.Header{FirstUsableLBA: 5, LastUsableLBA: 10},
			Partitions: []gpt.Partition{
				{FirstLBA: 7, LastLBA: 8, Type: gpt.PartType(testUuids[1])},
			},
		}, 1, End, 10, 10, true},
		{gpt.Table{
			Header: gpt.Header{FirstUsableLBA: 5, LastUsableLBA: 10},
			Partitions: []gpt.Partition{
				{FirstLBA: 7, LastLBA: 8, Type: gpt.PartType(testUuids[1])},
			},
		}, 2, Start, 5, 6, true},
		{gpt.Table{
			Header: gpt.Header{FirstUsableLBA: 5, LastUsableLBA: 10},
			Partitions: []gpt.Partition{
				{FirstLBA: 7, LastLBA: 8, Type: gpt.PartType(testUuids[1])},
			},
		}, 2, End, 9, 10, true},
		{gpt.Table{
			Header: gpt.Header{FirstUsableLBA: 5, LastUsableLBA: 10},
			Partitions: []gpt.Partition{
				{FirstLBA: 7, LastLBA: 8, Type: gpt.PartType(testUuids[1])},
			},
		}, 3, Start, 0, 0, false},
		{gpt.Table{
			Header: gpt.Header{FirstUsableLBA: 5, LastUsableLBA: 10},
			Partitions: []gpt.Partition{
				{FirstLBA: 7, LastLBA: 8, Type: gpt.PartType(testUuids[1])},
			},
		}, 3, End, 0, 0, false},
		{gpt.Table{
			Header: gpt.Header{FirstUsableLBA: 5, LastUsableLBA: 20},
			Partitions: []gpt.Partition{
				{FirstLBA: 7, LastLBA: 8, Type: gpt.PartType(testUuids[1])},
				{FirstLBA: 15, LastLBA: 18, Type: gpt.PartType(testUuids[1])},
			},
		}, 2, Start, 5, 6, true},
		{gpt.Table{
			Header: gpt.Header{FirstUsableLBA: 5, LastUsableLBA: 20},
			Partitions: []gpt.Partition{
				{FirstLBA: 7, LastLBA: 8, Type: gpt.PartType(testUuids[1])},
				{FirstLBA: 15, LastLBA: 18, Type: gpt.PartType(testUuids[1])},
			},
		}, 2, End, 19, 20, true},
		{gpt.Table{
			Header: gpt.Header{FirstUsableLBA: 5, LastUsableLBA: 20},
			Partitions: []gpt.Partition{
				{FirstLBA: 7, LastLBA: 8, Type: gpt.PartType(testUuids[1])},
				{FirstLBA: 15, LastLBA: 18, Type: gpt.PartType(testUuids[1])},
			},
		}, 3, Start, 9, 11, true},
		{gpt.Table{
			Header: gpt.Header{FirstUsableLBA: 5, LastUsableLBA: 20},
			Partitions: []gpt.Partition{
				{FirstLBA: 7, LastLBA: 8, Type: gpt.PartType(testUuids[1])},
				{FirstLBA: 15, LastLBA: 18, Type: gpt.PartType(testUuids[1])},
			},
		}, 3, End, 12, 14, true},
		{gpt.Table{
			Header: gpt.Header{FirstUsableLBA: 5, LastUsableLBA: 20},
			Partitions: []gpt.Partition{
				{FirstLBA: 7, LastLBA: 8, Type: gpt.PartType(testUuids[1])},
				{FirstLBA: 15, LastLBA: 18, Type: gpt.PartType(testUuids[1])},
			},
		}, 6, Start, 9, 14, true},
		{gpt.Table{
			Header: gpt.Header{FirstUsableLBA: 5, LastUsableLBA: 20},
			Partitions: []gpt.Partition{
				{FirstLBA: 7, LastLBA: 8, Type: gpt.PartType(testUuids[1])},
				{FirstLBA: 15, LastLBA: 18, Type: gpt.PartType(testUuids[1])},
			},
		}, 6, End, 9, 14, true},
		{gpt.Table{
			Header: gpt.Header{FirstUsableLBA: 5, LastUsableLBA: 20},
			Partitions: []gpt.Partition{
				{FirstLBA: 7, LastLBA: 8, Type: gpt.PartType(testUuids[1])},
				{FirstLBA: 15, LastLBA: 18, Type: gpt.PartType(testUuids[1])},
			},
		}, 7, Start, 0, 0, false},
		{gpt.Table{
			Header: gpt.Header{FirstUsableLBA: 5, LastUsableLBA: 20},
			Partitions: []gpt.Partition{
				{FirstLBA: 7, LastLBA: 8, Type: gpt.PartType(testUuids[1])},
				{FirstLBA: 15, LastLBA: 18, Type: gpt.PartType(testUuids[1])},
			},
		}, 7, End, 0, 0, false},
		// Empty partitions don't count
		{gpt.Table{
			Header: gpt.Header{FirstUsableLBA: 5, LastUsableLBA: 10},
			Partitions: []gpt.Partition{
				{FirstLBA: 5, LastLBA: 6, Type: gpt.PartType(testUuids[0])},
				{FirstLBA: 7, LastLBA: 9, Type: gpt.PartType(testUuids[0])},
			},
		}, 3, Start, 5, 7, true},
		{gpt.Table{
			Header: gpt.Header{FirstUsableLBA: 5, LastUsableLBA: 5},
			Partitions: []gpt.Partition{
				{FirstLBA: 5, LastLBA: 5, Type: gpt.PartType(testUuids[0])},
				{FirstLBA: 5, LastLBA: 5, Type: gpt.PartType(testUuids[0])},
				{FirstLBA: 5, LastLBA: 5, Type: gpt.PartType(testUuids[0])},
			},
		}, 1, End, 5, 5, true},
	}
	for i, c := range cases {
		firstLBA, lastLBA, found := FindSpace(&c.table, c.blocks, c.side)
		if firstLBA != c.expectedFirstLBA {
			t.Errorf("Test case %v: Wrong FirstLBA - expected/actual: %v/%v",
				i, c.expectedFirstLBA, firstLBA)
		}
		if lastLBA != c.expectedLastLBA {
			t.Errorf("Test case %v: Wrong LastLBA - expected/actual: %v/%v",
				i, c.expectedLastLBA, lastLBA)
		}
		if found != c.expectedFound {
			t.Errorf("Test case %v: Wrong found flag - expected/actual: %v/%v",
				i, c.expectedFound, found)
		}
	}
}
