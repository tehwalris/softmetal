package copyimg

import (
	"fmt"
	"io"

	"git.dolansoft.org/philippe/softmetal/flashing-agent/partition"
	"github.com/rekby/gpt"
)

// Task describes a copy operation on a continous section of data.
type Task struct {
	Src  uint64 // index where first byte is copied from
	Dst  uint64 // index where first byte is copied to
	Size uint64 // number of bytes to copy
}

// CopyToSeeker copies specified regions from a non-seekable source to a seekable destination.
// The specified copy tasks do not have to be ordered.
// CopyToSeeker fails if two tasks have overlapping source regions.
// CopyToSeeker fails if two tasks have overlapping destination regions.
// CopyToSeeker fails if any source or target region is out of range.
func CopyToSeeker(src io.Reader, dst io.ReadSeeker, tasks []Task) error {
	return fmt.Errorf("not implemented")
}

// PlanFromGPTs plans copy operations to transfer data for all partitions which are in both GPT tables.
// Partitions which are only in one table are not copied.
// Empty partitions (GPT type 0) are ignored.
// Nothing except the partition contents is copied (not even the GPT tables themselves).
// Only partition sizes from the source are used. Sizes of destination partitions are ignored.
// The passed GPT tables are generally not validated (eg. for duplicate partitions).
// Because of this, it is not guaranteed that copy tasks do not overlap.
func PlanFromGPTs(src *gpt.Table, dst *gpt.Table) ([]Task, error) {
	var out []Task
	for _, s := range src.Partitions {
		if s.IsEmpty() {
			continue
		}
		for _, d := range dst.Partitions {
			if d.IsEmpty() || !partition.EqGUID(s.Id, d.Id) {
				continue
			}
			if s.FirstLBA > s.LastLBA {
				return nil, fmt.Errorf(
					"got partition %v with (FirstLBA: %v, LastLBA: %v)",
					s.Id.String(), s.FirstLBA, s.LastLBA)
			}
			out = append(out, Task{
				Src:  s.FirstLBA * src.SectorSize,
				Dst:  d.FirstLBA * dst.SectorSize,
				Size: (s.LastLBA - s.FirstLBA + 1) * src.SectorSize,
			})
		}
	}
	return out, nil
}
