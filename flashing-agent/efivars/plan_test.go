package efivars_test

import (
	"fmt"
	"math"
	"reflect"
	"testing"

	"github.com/rekby/gpt"

	"git.dolansoft.org/philippe/softmetal/flashing-agent/efivars"
)

func TestPlanUpdate(t *testing.T) {
	typicalIn := efivars.BootEntry{
		Path:            `\test\efi\path`,
		PartitionGUID:   testUuids[1],
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
				Path:            `\test\efi\path`,
				PartitionGUID:   testUuids[1],
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
	espType := gpt.PartType{0x28, 0x73, 0x2A, 0xC1, 0x1F, 0xF8, 0xD2, 0x11, 0xBA, 0x4B, 0x00, 0xA0, 0xC9, 0x3E, 0xC9, 0x3B}
	bootEntry := func(idIdx int, num uint32, start uint64, size uint64) *efivars.BootEntry {
		return &efivars.BootEntry{
			Path:            `\test\path`,
			PartitionGUID:   testUuids[idIdx],
			PartitionNumber: num,
			PartitionStart:  start,
			PartitionSize:   size,
		}
	}
	partition := func(typeIdx int, idIdx int, firstLBA uint64, lastLBA uint64) gpt.Partition {
		t := espType
		if typeIdx != -1 {
			t = gpt.PartType(testUuids[typeIdx])
		}
		return gpt.Partition{
			Type:     t,
			Id:       testUuids[idIdx],
			FirstLBA: firstLBA,
			LastLBA:  lastLBA,
		}
	}

	cases := []struct {
		label      string
		partitions []gpt.Partition
		exp        *efivars.BootEntry
		shouldFail bool
	}{
		{"no partitions", []gpt.Partition{}, nil, true},
		{"single partition (ESP)", []gpt.Partition{
			partition(-1, 1, 20, 30),
		}, bootEntry(1, 1, 20, 11), false},
		{"multiple ESPs", []gpt.Partition{
			partition(-1, 1, 20, 30),
			partition(-1, 3, 40, 50),
		}, nil, true},
		{"typical", []gpt.Partition{
			partition(2, 1, 20, 30),
			partition(-1, 4, 25, 26),
			partition(1, 4, 40, 50),
		}, bootEntry(4, 2, 25, 2), false},
		{"ignores overlap and duplicate IDs", []gpt.Partition{
			partition(2, 1, 20, 40),
			partition(-1, 4, 25, 26),
			partition(1, 4, 10, 50),
			partition(1, 4, 30, 42),
		}, bootEntry(4, 2, 25, 2), false},
	}

	for _, c := range cases {
		t.Run(c.label, func(t *testing.T) {
			act, actErr := efivars.NewBootEntry(`\test\path`, c.partitions)
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
