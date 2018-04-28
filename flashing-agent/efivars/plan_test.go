package efivars_test

import (
	"reflect"
	"testing"

	"git.dolansoft.org/philippe/softmetal/flashing-agent/efivars"
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
