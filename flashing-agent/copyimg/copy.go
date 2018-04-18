package copyimg

import (
	"fmt"
	"io"

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
// Nothing except the parition contents is copied (not even the GPT tables themselves).
// The given GPT tables are not validated (eg. for duplicate partitions).
// Because of this, it is not guaranteed that copy tasks do not overlap.
func PlanFromGPTs(src *gpt.Table, dst *gpt.Table) []Task {
	return []Task{}
}
