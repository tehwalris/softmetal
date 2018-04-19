package copyimg

import (
	"fmt"
	"io"
	"io/ioutil"
	"sort"

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
// When CopyToSeeker fails, the value of destination bytes is undefined.
// Destination bytes which are not referenced by tasks never change (even after errors).
func CopyToSeeker(dst io.WriteSeeker, src io.Reader, unsortedTasks []Task) error {
	var tasks tasks
	tasks.d = make([]Task, len(unsortedTasks))
	copy(tasks.d, unsortedTasks)

	tasks.useDst = true
	sort.Sort(tasks)
	var i int64
	for _, t := range tasks.d {
		delta := int64(t.Dst) - i
		if delta < 0 {
			return fmt.Errorf("destination regions of tasks overlap")
		}
		i = int64(t.Dst + t.Size)
	}

	tasks.useDst = false
	sort.Sort(tasks)
	i = 0
	for _, t := range tasks.d {
		srcDelta := int64(t.Src) - i
		if srcDelta < 0 {
			return fmt.Errorf("source regions of tasks overlap")
		}
		n, e := io.CopyN(ioutil.Discard, src, srcDelta)
		if e != nil {
			return e
		}
		i += n

		if _, e := dst.Seek(int64(t.Dst), io.SeekStart); e != nil {
			return e
		}

		n, e = io.CopyN(dst, src, int64(t.Size))
		if e != nil {
			return e
		}
		i += n
	}
	return nil
}

// Tasks wraps []Task, so that sort.Interface can be implemented.
type tasks struct {
	d      []Task
	useDst bool
}

func (c tasks) Len() int { return len(c.d) }

func (c tasks) Less(i, j int) bool {
	if c.useDst {
		return c.d[i].Dst < c.d[j].Dst
	}
	return c.d[i].Src < c.d[j].Src
}

func (c tasks) Swap(i, j int) {
	ti := c.d[i]
	c.d[i] = c.d[j]
	c.d[j] = ti
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
