package efivars_test

import (
	"encoding/hex"
	"reflect"
	"testing"

	"git.dolansoft.org/philippe/softmetal/flashing-agent/efivars"
	"github.com/rekby/gpt"
)

var testUuids = []gpt.Guid{
	{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
	{0x1f, 0xe6, 0x90, 0x41, 0xfc, 0xda, 0xb9, 0x4d, 0x83, 0x21, 0xa5, 0xc9, 0x28, 0x47, 0xf7, 0x6b},
	{0x79, 0xdd, 0x87, 0xb1, 0x5f, 0xb8, 0x02, 0x44, 0x88, 0xe8, 0x6d, 0xe0, 0xf9, 0x33, 0x16, 0x62},
}

var testUuidStrings = []string{
	"00000000-0000-0000-0000-000000000000",
	"4190E61F-DAFC-4DB9-8321-A5C92847F76B",
	"B187DD79-B85F-4402-88E8-6DE0F9331662",
}

func TestGuidStringsMatch(t *testing.T) {
	if len(testUuids) != len(testUuidStrings) {
		t.Errorf("length mismatch")
	}
	for i, v := range testUuids {
		a := v.String()
		b := testUuidStrings[i]
		if a != b {
			t.Errorf("got %v, want %v", a, b)
		}
	}
}

func TestMarshalBootEntry(t *testing.T) {
	var longStrBuf []byte
	for i := 0; i < 33000; i++ {
		longStrBuf = append(longStrBuf, 'a')
	}
	longStr := string(longStrBuf)

	cases := []struct {
		label        string
		input        efivars.BootEntry
		exp          []byte
		shouldFail   bool
		ignoreResult bool
	}{
		{"typical",
			efivars.BootEntry{
				Description:     "Linux Boot Manager",
				DiskGUID:        testUuids[2],
				Path:            `\EFI\systemd\systemd-bootx64.efi`,
				PartitionNumber: 1,
				PartitionStart:  0x0800,
				PartitionSize:   0x02F800,
			},
			[]byte{
				0x07, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00 /**/, 0x74, 0x00, 0x4c, 0x00, 0x69, 0x00, 0x6e, 0x00,
				0x75, 0x00, 0x78, 0x00, 0x20, 0x00, 0x42, 0x00 /**/, 0x6f, 0x00, 0x6f, 0x00, 0x74, 0x00, 0x20, 0x00,
				0x4d, 0x00, 0x61, 0x00, 0x6e, 0x00, 0x61, 0x00 /**/, 0x67, 0x00, 0x65, 0x00, 0x72, 0x00, 0x00, 0x00,
				0x04, 0x01, 0x2a, 0x00, 0x01, 0x00, 0x00, 0x00 /**/, 0x00, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0xf8, 0x02, 0x00, 0x00, 0x00, 0x00, 0x00 /**/, 0x79, 0xdd, 0x87, 0xb1, 0x5f, 0xb8, 0x02, 0x44,
				0x88, 0xe8, 0x6d, 0xe0, 0xf9, 0x33, 0x16, 0x62 /**/, 0x02, 0x02, 0x04, 0x04, 0x46, 0x00, 0x5c, 0x00,
				0x45, 0x00, 0x46, 0x00, 0x49, 0x00, 0x5c, 0x00 /**/, 0x73, 0x00, 0x79, 0x00, 0x73, 0x00, 0x74, 0x00,
				0x65, 0x00, 0x6d, 0x00, 0x64, 0x00, 0x5c, 0x00 /**/, 0x73, 0x00, 0x79, 0x00, 0x73, 0x00, 0x74, 0x00,
				0x65, 0x00, 0x6d, 0x00, 0x64, 0x00, 0x2d, 0x00 /**/, 0x62, 0x00, 0x6f, 0x00, 0x6f, 0x00, 0x74, 0x00,
				0x78, 0x00, 0x36, 0x00, 0x34, 0x00, 0x2e, 0x00 /**/, 0x65, 0x00, 0x66, 0x00, 0x69, 0x00, 0x00, 0x00,
				0x7f, 0xff, 0x04, 0x00,
			}, false, false},
		{"empty Description", efivars.BootEntry{
			DiskGUID:        testUuids[2],
			Path:            `\EFI\systemd\systemd-bootx64.efi`,
			PartitionNumber: 1,
			PartitionStart:  0x0800,
			PartitionSize:   0x02F800,
		}, nil, true, false},
		{"empty DiskGUID", efivars.BootEntry{
			Description:     "Linux Boot Manager",
			Path:            `\EFI\systemd\systemd-bootx64.efi`,
			PartitionNumber: 1,
			PartitionStart:  0x0800,
			PartitionSize:   0x02F800,
		}, nil, true, false},
		{"empty Path", efivars.BootEntry{
			Description:     "Linux Boot Manager",
			DiskGUID:        testUuids[2],
			PartitionNumber: 1,
			PartitionStart:  0x0800,
			PartitionSize:   0x02F800,
		}, nil, true, false},
		{"empty PartitionNumber", efivars.BootEntry{
			Description:    "Linux Boot Manager",
			DiskGUID:       testUuids[2],
			Path:           `\EFI\systemd\systemd-bootx64.efi`,
			PartitionStart: 0x0800,
			PartitionSize:  0x02F800,
		}, nil, true, false},
		{"empty PartitionStart", efivars.BootEntry{
			Description:     "Linux Boot Manager",
			DiskGUID:        testUuids[2],
			Path:            `\EFI\systemd\systemd-bootx64.efi`,
			PartitionNumber: 1,
			PartitionSize:   0x02F800,
		}, nil, true, false},
		{"empty PartitionSize", efivars.BootEntry{
			Description:     "Linux Boot Manager",
			DiskGUID:        testUuids[2],
			Path:            `\EFI\systemd\systemd-bootx64.efi`,
			PartitionNumber: 1,
			PartitionStart:  0x0800,
		}, nil, true, false},
		{"too large (long Path)",
			efivars.BootEntry{
				Description:     "some description",
				DiskGUID:        testUuids[2],
				Path:            longStr,
				PartitionNumber: 1,
				PartitionStart:  0x0800,
				PartitionSize:   0x02F800,
			}, nil, true, true},
		{"not too large (long Description)",
			efivars.BootEntry{
				Description:     longStr,
				DiskGUID:        testUuids[2],
				Path:            "some short path",
				PartitionNumber: 1,
				PartitionStart:  0x0800,
				PartitionSize:   0x02F800,
			}, nil, false, true},
	}
	for _, c := range cases {
		t.Run(c.label, func(t *testing.T) {
			act, actErr := c.input.Marshal()
			if c.shouldFail && actErr == nil {
				t.Errorf("got no error, want some error")
			}
			if !c.shouldFail && actErr != nil {
				t.Errorf("unexpected error: %v", actErr)
			}
			if c.ignoreResult {
				return
			}
			if !reflect.DeepEqual(act, c.exp) {
				t.Errorf("got %+v, want %+v", hex.EncodeToString(act), hex.EncodeToString(c.exp))
			}
			var unequalCount int
			for i, a := range act {
				e := c.exp[i]
				if a != e {
					t.Errorf("byte %v is unequal: got 0x%x, want 0x%x", i, a, e)
					unequalCount++
				}
				if unequalCount > 20 {
					break
				}
			}
		})
	}
}

func TestMarshalBootOrder(t *testing.T) {
	cases := []struct {
		label string
		input efivars.BootOrder
		exp   []byte
	}{
		{"empty", efivars.BootOrder{}, []byte{0x07, 0x00, 0x00, 0x00}},
		{"typical",
			efivars.BootOrder{0x0000, 0x0010, 0x0011, 0x0012, 0x0013, 0x0017, 0x0018, 0x0019, 0x001A, 0x001B, 0x001C},
			[]byte{
				0x07, 0x00, 0x00, 0x00, 0x00, 0x00, 0x10, 0x00,
				0x11, 0x00, 0x12, 0x00, 0x13, 0x00, 0x17, 0x00,
				0x18, 0x00, 0x19, 0x00, 0x1A, 0x00, 0x1B, 0x00,
				0x1C, 0x00}},
		{"high numbers", efivars.BootOrder{0xFF00, 0x1234, 0xFFFF},
			[]byte{0x07, 0x00, 0x00, 0x00, 0x00, 0xFF, 0x34, 0x12, 0xFF, 0xFF}},
		{"no special duplicate handling", efivars.BootOrder{0x1234, 0x1234},
			[]byte{0x07, 0x00, 0x00, 0x00, 0x34, 0x12, 0x34, 0x12}},
	}
	for _, c := range cases {
		t.Run(c.label, func(t *testing.T) {
			act := c.input.Marhsal()
			if !reflect.DeepEqual(act, c.exp) {
				t.Errorf("got %+v, want %+v", hex.EncodeToString(act), hex.EncodeToString(c.exp))
			}
		})
	}
}

func TestUnmarshalBootOrder(t *testing.T) {
	cases := []struct {
		label      string
		input      []byte
		exp        *efivars.BootOrder
		shouldFail bool
	}{
		{"empty", []byte{0x07, 0x00, 0x00, 0x00}, &efivars.BootOrder{}, false},
		{"typical",
			[]byte{
				0x07, 0x00, 0x00, 0x00, 0x00, 0x00, 0x10, 0x00,
				0x11, 0x00, 0x12, 0x00, 0x13, 0x00, 0x17, 0x00,
				0x18, 0x00, 0x19, 0x00, 0x1A, 0x00, 0x1B, 0x00,
				0x1C, 0x00},
			&efivars.BootOrder{0x0000, 0x0010, 0x0011, 0x0012, 0x0013, 0x0017, 0x0018, 0x0019, 0x001A, 0x001B, 0x001C},
			false},
		{"too short (0 bytes)", []byte{}, nil, true},
		{"too short (3 bytes)", []byte{0x07, 0x00, 0x00}, nil, true},
		{"ignores attributes", []byte{0x12, 0xB4, 0xF3, 0x20}, &efivars.BootOrder{}, false},
		{"uneven length", []byte{0x12, 0xB4, 0xF3, 0x20, 0x34, 0x12, 0x00}, nil, true},
		{"high numbers",
			[]byte{0x07, 0x00, 0x00, 0x00, 0x00, 0xFF, 0x34, 0x12, 0xFF, 0xFF},
			&efivars.BootOrder{0xFF00, 0x1234, 0xFFFF},
			false},
		{"no special duplicate handling",
			[]byte{0x07, 0x00, 0x00, 0x00, 0x34, 0x12, 0x34, 0x12},
			&efivars.BootOrder{0x1234, 0x1234},
			false},
	}
	for _, c := range cases {
		t.Run(c.label, func(t *testing.T) {
			act, actErr := efivars.UnmarshalBootOrder(c.input)
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
