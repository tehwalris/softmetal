package partition

import (
	"testing"

	"github.com/rekby/gpt"
)

// TODO more test cases
func TestAssertGptCompatible(t *testing.T) {
	disk := gpt.Table{
		SectorSize: 512,
		Header:     gpt.Header{},
		Partitions: []gpt.Partition{},
	}
	img := gpt.Table{
		SectorSize: 1024,
		Header:     gpt.Header{},
		Partitions: []gpt.Partition{},
	}
	e := AssertGptCompatible(disk, img)
	if e == nil {
		t.Fatalf("Mismated sector sizes not detected.")
	}
}

// TODO TestAssertValidLayout
