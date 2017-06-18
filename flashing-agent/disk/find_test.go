package disk

import (
	"strings"
	"testing"

	"github.com/jaypipes/ghw"
)

func TestFindDisk(t *testing.T) {
	var cases = []struct {
		disks             []ghw.Disk
		targetSerial      string
		expectedDiskIndex uint
		shouldFind        bool
		shouldFail        bool
		shouldContain     string
	}{
		{[]ghw.Disk{}, "testTargetSerial", 0, false, false, ""},
		{[]ghw.Disk{}, "", 0, false, true, "Bad serial"},
		{[]ghw.Disk{}, "unknown", 0, false, true, "Bad serial"},
		{[]ghw.Disk{
			{SerialNumber: "testDisk0Serial"},
		}, "testDisk0Serial", 0, true, false, ""},
		{[]ghw.Disk{
			{SerialNumber: "testDisk0Serial"},
		}, "testDisk2Serial", 0, false, false, ""},
		{[]ghw.Disk{
			{SerialNumber: "testDisk0Serial"},
			{SerialNumber: "testDisk1Serial"},
			{SerialNumber: "testDisk2Serial"},
			{SerialNumber: "testDisk3Serial"},
		}, "testDisk2Serial", 2, true, false, ""},
		{[]ghw.Disk{
			{SerialNumber: "testDisk0Serial"},
			{SerialNumber: "testDisk2Serial"},
			{SerialNumber: "testDisk2Serial"},
			{SerialNumber: "testDisk3Serial"},
		}, "testDisk2Serial", 0, false, true, "duplicate serial"},
	}

	for i, c := range cases {
		var diskPointers = make([]*ghw.Disk, len(c.disks))
		for j, _ := range c.disks {
			diskPointers[j] = &c.disks[j]
		}
		var expectedDisk *ghw.Disk = nil
		if c.shouldFind {
			expectedDisk = diskPointers[c.expectedDiskIndex]
		}
		var blockInfo = ghw.BlockInfo{Disks: diskPointers}
		disk, found, e := FindDisk(&blockInfo, c.targetSerial)

		if disk != expectedDisk {
			t.Errorf("Test case %v: Returned disk did not match expected.")
		}

		if found != c.shouldFind {
			t.Errorf("Test case %v: Found flag not correct. Expected/actual: %v/%v.",
				i, c.shouldFind, found)
		}

		if c.shouldFail {
			if e == nil {
				t.Errorf("Test case %v: Excpected error, but none occured", i)
			} else if !strings.Contains(e.Error(), c.shouldContain) {
				t.Errorf("Test case %v: Excpected error to contain \"%v\", but it didn't."+
					"Instead error was: %v", i, c.shouldContain, e)
			}
		} else {
			if e != nil {
				t.Errorf("Test case %v: Excpected no error, but got: %v", i, e)
			}
		}
	}
}
