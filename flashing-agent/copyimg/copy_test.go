package copyimg_test

import (
	"reflect"
	"testing"

	"git.dolansoft.org/philippe/softmetal/flashing-agent/copyimg"
	"github.com/rekby/gpt"
)

func TestPlanFromGPTs(t *testing.T) {
	cases := []struct {
		label      string
		src        gpt.Table
		dst        gpt.Table
		expTasks   []copyimg.Task
		shouldFail bool
	}{
		{
			"empty to empty",
			gpt.Table{},
			gpt.Table{},
			[]copyimg.Task{},
			false,
		},
		{
			"single partition (same size and location)",
			gpt.Table{
				SectorSize: 1024,
				Partitions: []gpt.Partition{
					{Id: testUuids[1], Type: gpt.PartType(testUuids[2]), FirstLBA: 30, LastLBA: 40},
				},
			},
			gpt.Table{
				SectorSize: 1024,
				Partitions: []gpt.Partition{
					{Id: testUuids[1], Type: gpt.PartType(testUuids[2]), FirstLBA: 30, LastLBA: 40},
				},
			},
			[]copyimg.Task{
				{Src: 30 * 1024, Dst: 30 * 1024, Size: 11 * 1024},
			},
			false,
		},
		{
			"single partition (same size, but shifted)",
			gpt.Table{
				SectorSize: 1024,
				Partitions: []gpt.Partition{
					{Id: testUuids[1], Type: gpt.PartType(testUuids[2]), FirstLBA: 30, LastLBA: 40},
				},
			},
			gpt.Table{
				SectorSize: 1024,
				Partitions: []gpt.Partition{
					{Id: testUuids[1], Type: gpt.PartType(testUuids[2]), FirstLBA: 13, LastLBA: 23},
				},
			},
			[]copyimg.Task{
				{Src: 30 * 1024, Dst: 13 * 1024, Size: 11 * 1024},
			},
			false,
		},
		{
			"single partition (different LBA and sector sizes)",
			gpt.Table{
				SectorSize: 512,
				Partitions: []gpt.Partition{
					{Id: testUuids[1], Type: gpt.PartType(testUuids[2]), FirstLBA: 30, LastLBA: 40},
				},
			},
			gpt.Table{
				SectorSize: 666,
				Partitions: []gpt.Partition{
					{Id: testUuids[1], Type: gpt.PartType(testUuids[2]), FirstLBA: 13, LastLBA: 23},
				},
			},
			[]copyimg.Task{
				{Src: 30 * 512, Dst: 13 * 666, Size: 11 * 512},
			},
			false,
		},
		{
			"single partition (ignores different partition types)",
			gpt.Table{
				SectorSize: 1024,
				Partitions: []gpt.Partition{
					{Id: testUuids[1], Type: gpt.PartType(testUuids[1]), FirstLBA: 30, LastLBA: 40},
				},
			},
			gpt.Table{
				SectorSize: 1024,
				Partitions: []gpt.Partition{
					{Id: testUuids[1], Type: gpt.PartType(testUuids[3]), FirstLBA: 30, LastLBA: 40},
				},
			},
			[]copyimg.Task{
				{Src: 30 * 1024, Dst: 30 * 1024, Size: 11 * 1024},
			},
			false,
		},
		{
			"two partitions (reordered in slice)",
			gpt.Table{
				SectorSize: 1024,
				Partitions: []gpt.Partition{
					{Id: testUuids[1], Type: gpt.PartType(testUuids[2]), FirstLBA: 30, LastLBA: 40},
					{Id: testUuids[2], Type: gpt.PartType(testUuids[2]), FirstLBA: 50, LastLBA: 65},
				},
			},
			gpt.Table{
				SectorSize: 1024,
				Partitions: []gpt.Partition{
					{Id: testUuids[2], Type: gpt.PartType(testUuids[2]), FirstLBA: 50, LastLBA: 65},
					{Id: testUuids[1], Type: gpt.PartType(testUuids[2]), FirstLBA: 30, LastLBA: 40},
				},
			},
			[]copyimg.Task{
				{Src: 30 * 1024, Dst: 30 * 1024, Size: 11 * 1024},
				{Src: 50 * 1024, Dst: 50 * 1024, Size: 16 * 1024},
			},
			false,
		},
		{
			"single partition (one LBA in size)",
			gpt.Table{
				SectorSize: 512,
				Partitions: []gpt.Partition{
					{Id: testUuids[1], Type: gpt.PartType(testUuids[2]), FirstLBA: 30, LastLBA: 30},
				},
			},
			gpt.Table{
				SectorSize: 512,
				Partitions: []gpt.Partition{
					{Id: testUuids[1], Type: gpt.PartType(testUuids[2]), FirstLBA: 13, LastLBA: 13},
				},
			},
			[]copyimg.Task{
				{Src: 30 * 512, Dst: 13 * 512, Size: 512},
			},
			false,
		},
		{
			"skips missing partitions (in both source and destination)",
			gpt.Table{
				SectorSize: 1024,
				Partitions: []gpt.Partition{
					{Id: testUuids[1], Type: gpt.PartType(testUuids[2]), FirstLBA: 10, LastLBA: 15},
					{Id: testUuids[3], Type: gpt.PartType(testUuids[2]), FirstLBA: 30, LastLBA: 40},
					{Id: testUuids[2], Type: gpt.PartType(testUuids[0]), FirstLBA: 50, LastLBA: 65},
				},
			},
			gpt.Table{
				SectorSize: 1024,
				Partitions: []gpt.Partition{
					{Id: testUuids[1], Type: gpt.PartType(testUuids[0]), FirstLBA: 10, LastLBA: 15},
					{Id: testUuids[2], Type: gpt.PartType(testUuids[2]), FirstLBA: 50, LastLBA: 65},
					{Id: testUuids[3], Type: gpt.PartType(testUuids[2]), FirstLBA: 30, LastLBA: 40},
				},
			},
			[]copyimg.Task{
				{Src: 30 * 1024, Dst: 30 * 1024, Size: 11 * 1024},
			},
			false,
		},
		{
			"fails when copied source partition has FirstLBA < LastLBA",
			gpt.Table{
				SectorSize: 512,
				Partitions: []gpt.Partition{
					{Id: testUuids[1], Type: gpt.PartType(testUuids[2]), FirstLBA: 30, LastLBA: 29},
				},
			},
			gpt.Table{
				SectorSize: 512,
				Partitions: []gpt.Partition{
					{Id: testUuids[1], Type: gpt.PartType(testUuids[2]), FirstLBA: 13, LastLBA: 13},
				},
			},
			nil,
			true,
		},
		{
			"doesn't fail when destination, uncopied source or empty partitions have FirstLBA < LastLBA",
			gpt.Table{
				SectorSize: 512,
				Partitions: []gpt.Partition{
					{Id: testUuids[1], Type: gpt.PartType(testUuids[2]), FirstLBA: 10, LastLBA: 15},
					{Id: testUuids[2], Type: gpt.PartType(testUuids[2]), FirstLBA: 50, LastLBA: 49}, // uncopied
					{Id: testUuids[3], Type: gpt.PartType(testUuids[0]), FirstLBA: 30, LastLBA: 29}, // empty
				},
			},
			gpt.Table{
				SectorSize: 512,
				Partitions: []gpt.Partition{
					{Id: testUuids[1], Type: gpt.PartType(testUuids[2]), FirstLBA: 10, LastLBA: 9},
					{Id: testUuids[3], Type: gpt.PartType(testUuids[2]), FirstLBA: 30, LastLBA: 29},
				},
			},
			[]copyimg.Task{
				{Src: 10 * 512, Dst: 10 * 512, Size: 6 * 512},
			},
			false,
		},
	}

	for _, c := range cases {
		t.Run(c.label, func(t *testing.T) {
			act, actErr := copyimg.PlanFromGPTs(&c.src, &c.dst)
			if c.shouldFail && actErr == nil {
				t.Errorf("got no error, want some error")
			}
			if !c.shouldFail && actErr != nil {
				t.Errorf("unexpected error: %v", actErr)
			}
			if !(reflect.DeepEqual(act, c.expTasks) || (len(act) == 0 && len(c.expTasks) == 0)) {
				t.Errorf("got %v, want %v", act, c.expTasks)
			}
		})
	}
}
