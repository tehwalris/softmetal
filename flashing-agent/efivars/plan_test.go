package efivars_test

import (
	"fmt"
	"math"
	"reflect"
	"testing"

	"github.com/rekby/gpt"

	"git.dolansoft.org/philippe/softmetal/flashing-agent/efivars"
	pb "git.dolansoft.org/philippe/softmetal/pb"
)

func TestPlanUpdate(t *testing.T) {
	typicalIn := efivars.BootEntry{
		DiskGUID:        testUuids[1],
		Path:            `\test\efi\path`,
		PartitionNumber: 3,
		PartitionStart:  20,
		PartitionSize:   123,
	}
	typicalExp := typicalIn
	typicalExp.Description = "Softmetal (boot from disk)"

	fakeEntries := func(n int) map[uint16]efivars.BootEntry {
		out := make(map[uint16]efivars.BootEntry)
		for i := 0; i < n; i++ {
			out[uint16(i)] = efivars.BootEntry{Description: fmt.Sprintf("test entry 0x%x", i)}
		}
		if len(out) != n {
			panic("length mismatch, probably overflow")
		}
		return out
	}

	cases := []struct {
		label      string
		oldOrd     efivars.BootOrder
		oldEntries map[uint16]efivars.BootEntry
		newEntryIn efivars.BootEntry
		expWrite   map[uint16]efivars.BootEntry
		expOrd     efivars.BootOrder
		shouldFail bool
	}{
		{"no entries and empty order",
			efivars.BootOrder{},
			map[uint16]efivars.BootEntry{},
			typicalIn,
			map[uint16]efivars.BootEntry{0x00: typicalExp},
			efivars.BootOrder{0x00},
			false},
		{"non-softmetal entries",
			efivars.BootOrder{0x03, 0x00, 0x07},
			map[uint16]efivars.BootEntry{
				0x03: {Description: "test entry 0x03"},
				0x00: {Description: "test entry 0x00"},
				0x07: {Description: "test entry 0x07"},
			},
			typicalIn,
			map[uint16]efivars.BootEntry{0x01: typicalExp},
			efivars.BootOrder{0x01, 0x03, 0x00, 0x07},
			false},
		{"one softmetal entry",
			efivars.BootOrder{0x03, 0x02},
			map[uint16]efivars.BootEntry{
				0x03: {Description: "test entry 0x03"},
				0x02: {Description: "Softmetal (boot from disk)"},
			},
			typicalIn,
			map[uint16]efivars.BootEntry{0x02: typicalExp},
			efivars.BootOrder{0x02, 0x03},
			false},
		{"multiple softmetal entries",
			efivars.BootOrder{0x07, 0x03, 0x02},
			map[uint16]efivars.BootEntry{
				0x07: {Description: "Softmetal (boot from disk)"},
				0x03: {Description: "test entry 0x03"},
				0x02: {Description: "Softmetal (boot from disk)"},
			},
			typicalIn, nil, nil, true},
		{"ignores duplicates in order",
			efivars.BootOrder{0x03, 0x00, 0x03},
			map[uint16]efivars.BootEntry{
				0x03: {Description: "test entry 0x03"},
				0x00: {Description: "test entry 0x00"},
			},
			typicalIn,
			map[uint16]efivars.BootEntry{0x01: typicalExp},
			efivars.BootOrder{0x01, 0x03, 0x00, 0x03},
			false},
		{"handles duplicates in order if duplicate is softmetal entry",
			efivars.BootOrder{0x03, 0x00, 0x03},
			map[uint16]efivars.BootEntry{
				0x03: {Description: "Softmetal (boot from disk)"},
				0x00: {Description: "test entry 0x00"},
			},
			typicalIn,
			map[uint16]efivars.BootEntry{0x03: typicalExp},
			efivars.BootOrder{0x03, 0x00},
			false},
		{"ignores undefined entries in order",
			efivars.BootOrder{0x03, 0x07},
			map[uint16]efivars.BootEntry{0x03: {Description: "test entry 0x03"}},
			typicalIn,
			map[uint16]efivars.BootEntry{0x00: typicalExp},
			efivars.BootOrder{0x00, 0x03, 0x07},
			false},
		{"ignores entries which are not in order",
			efivars.BootOrder{0x03},
			map[uint16]efivars.BootEntry{
				0x00: {Description: "test entry 0x00"},
				0x03: {Description: "test entry 0x03"},
			},
			typicalIn,
			map[uint16]efivars.BootEntry{0x01: typicalExp},
			efivars.BootOrder{0x01, 0x03},
			false},
		{"prefilled desciption",
			efivars.BootOrder{},
			map[uint16]efivars.BootEntry{},
			efivars.BootEntry{
				Description:     "test description",
				DiskGUID:        testUuids[1],
				Path:            `\test\efi\path`,
				PartitionNumber: 3,
				PartitionStart:  20,
				PartitionSize:   123,
			},
			nil, nil, true},
		{"no free IDs",
			efivars.BootOrder{},
			fakeEntries(65536),
			typicalIn,
			nil, nil, true},
		{"exactly one free ID",
			efivars.BootOrder{},
			fakeEntries(65535),
			typicalIn,
			map[uint16]efivars.BootEntry{math.MaxUint16: typicalExp},
			efivars.BootOrder{math.MaxUint16},
			false},
	}
	for _, c := range cases {
		t.Run(c.label, func(t *testing.T) {
			act, actErr := efivars.PlanUpdate(c.oldOrd, c.oldEntries, c.newEntryIn)
			if c.shouldFail && actErr == nil {
				t.Errorf("got no error, want some error")
			}
			if !c.shouldFail && actErr != nil {
				t.Errorf("unexpected error: %v", actErr)
			}
			exp := &efivars.Update{Write: c.expWrite, Order: c.expOrd}
			if c.shouldFail {
				exp = nil
			}
			if !reflect.DeepEqual(act, exp) {
				t.Errorf("got %+v, want %+v", act, exp)
			}
		})
	}
}

func TestNewBootEntry(t *testing.T) {
	testPartition := gpt.Partition{
		Type:     gpt.PartType(testUuids[2]),
		Id:       testUuids[1],
		FirstLBA: 12,
		LastLBA:  220,
	}
	testDiskGuid := testUuids[5]
	testBootEntry := efivars.BootEntry{
		DiskGUID:        testDiskGuid,
		Path:            `\test\path`,
		PartitionNumber: 0,
		PartitionStart:  12,
		PartitionSize:   209,
	}
	cases := []struct {
		label      string
		target     pb.FlashingConfig_BootEntry
		partitions []gpt.Partition
		exp        *efivars.BootEntry
		shouldFail bool
	}{
		{"single partition (matching)",
			pb.FlashingConfig_BootEntry{Path: `\test\path`, PartUuid: testUuidStrings[1]},
			[]gpt.Partition{testPartition},
			&testBootEntry,
			false},
		{"single partition (not matching)",
			pb.FlashingConfig_BootEntry{Path: `\test\path`, PartUuid: testUuidStrings[2]},
			[]gpt.Partition{testPartition},
			nil, true},
		{"single partition (not matching because empty)",
			pb.FlashingConfig_BootEntry{Path: `\test\path`, PartUuid: testUuidStrings[1]},
			[]gpt.Partition{{
				Type:     gpt.PartType(testUuids[0]),
				Id:       testUuids[1],
				FirstLBA: 12,
				LastLBA:  220,
			}},
			nil, true},
		{"ignores duplicate partitions in table",
			pb.FlashingConfig_BootEntry{Path: `\test\path`, PartUuid: testUuidStrings[1]},
			[]gpt.Partition{testPartition, testPartition},
			&testBootEntry,
			false},
		{"multiple partitions (one matching)",
			pb.FlashingConfig_BootEntry{Path: `\test\path`, PartUuid: testUuidStrings[1]},
			[]gpt.Partition{
				{Type: gpt.PartType(testUuids[2]),
					Id:       testUuids[3],
					FirstLBA: 30,
					LastLBA:  50},
				testPartition,
				{Type: gpt.PartType(testUuids[2]),
					Id:       testUuids[3],
					FirstLBA: 200,
					LastLBA:  400},
			},
			&testBootEntry,
			false},
		// TODO more cases
	}
	for _, c := range cases {
		t.Run(c.label, func(t *testing.T) {
			act, actErr := efivars.NewBootEntry(&c.target, c.partitions, testDiskGuid)
			if c.shouldFail && actErr == nil {
				t.Errorf("got no error, want some error")
			}
			if !c.shouldFail && actErr != nil {
				t.Errorf("unexpected error: %v", actErr)
			}
			if !reflect.DeepEqual(act, c.exp) {
				t.Errorf("got %+v, want %+v", act, c.exp)
			}
		})
	}
}
