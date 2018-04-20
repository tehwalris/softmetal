package copyimg_test

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"reflect"
	"testing"
	"time"

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
		{"empty to empty",
			gpt.Table{},
			gpt.Table{},
			[]copyimg.Task{},
			false},
		{"single partition (same size and location)",
			gpt.Table{SectorSize: 1024, Partitions: []gpt.Partition{
				{Id: testUuids[1], Type: gpt.PartType(testUuids[2]), FirstLBA: 30, LastLBA: 40},
			}},
			gpt.Table{SectorSize: 1024, Partitions: []gpt.Partition{
				{Id: testUuids[1], Type: gpt.PartType(testUuids[2]), FirstLBA: 30, LastLBA: 40},
			}},
			[]copyimg.Task{{Src: 30 * 1024, Dst: 30 * 1024, Size: 11 * 1024}},
			false},
		{"single partition (same size, but shifted)",
			gpt.Table{SectorSize: 1024, Partitions: []gpt.Partition{
				{Id: testUuids[1], Type: gpt.PartType(testUuids[2]), FirstLBA: 30, LastLBA: 40},
			}},
			gpt.Table{SectorSize: 1024, Partitions: []gpt.Partition{
				{Id: testUuids[1], Type: gpt.PartType(testUuids[2]), FirstLBA: 13, LastLBA: 23},
			}},
			[]copyimg.Task{{Src: 30 * 1024, Dst: 13 * 1024, Size: 11 * 1024}},
			false},
		{"single partition (different LBA and sector sizes)",
			gpt.Table{SectorSize: 512, Partitions: []gpt.Partition{
				{Id: testUuids[1], Type: gpt.PartType(testUuids[2]), FirstLBA: 30, LastLBA: 40},
			}},
			gpt.Table{SectorSize: 666, Partitions: []gpt.Partition{
				{Id: testUuids[1], Type: gpt.PartType(testUuids[2]), FirstLBA: 13, LastLBA: 23},
			}},
			[]copyimg.Task{{Src: 30 * 512, Dst: 13 * 666, Size: 11 * 512}},
			false},
		{"single partition (ignores different partition types)",
			gpt.Table{SectorSize: 1024, Partitions: []gpt.Partition{
				{Id: testUuids[1], Type: gpt.PartType(testUuids[1]), FirstLBA: 30, LastLBA: 40},
			}},
			gpt.Table{SectorSize: 1024, Partitions: []gpt.Partition{
				{Id: testUuids[1], Type: gpt.PartType(testUuids[3]), FirstLBA: 30, LastLBA: 40},
			}},
			[]copyimg.Task{{Src: 30 * 1024, Dst: 30 * 1024, Size: 11 * 1024}},
			false},
		{"two partitions (reordered in slice)",
			gpt.Table{SectorSize: 1024, Partitions: []gpt.Partition{
				{Id: testUuids[1], Type: gpt.PartType(testUuids[2]), FirstLBA: 30, LastLBA: 40},
				{Id: testUuids[2], Type: gpt.PartType(testUuids[2]), FirstLBA: 50, LastLBA: 65},
			}},
			gpt.Table{SectorSize: 1024, Partitions: []gpt.Partition{
				{Id: testUuids[2], Type: gpt.PartType(testUuids[2]), FirstLBA: 50, LastLBA: 65},
				{Id: testUuids[1], Type: gpt.PartType(testUuids[2]), FirstLBA: 30, LastLBA: 40},
			}},
			[]copyimg.Task{
				{Src: 30 * 1024, Dst: 30 * 1024, Size: 11 * 1024},
				{Src: 50 * 1024, Dst: 50 * 1024, Size: 16 * 1024},
			},
			false},
		{"single partition (one LBA in size)",
			gpt.Table{SectorSize: 512, Partitions: []gpt.Partition{
				{Id: testUuids[1], Type: gpt.PartType(testUuids[2]), FirstLBA: 30, LastLBA: 30},
			}},
			gpt.Table{SectorSize: 512, Partitions: []gpt.Partition{
				{Id: testUuids[1], Type: gpt.PartType(testUuids[2]), FirstLBA: 13, LastLBA: 13},
			}},
			[]copyimg.Task{{Src: 30 * 512, Dst: 13 * 512, Size: 512}},
			false},
		{"skips missing partitions (in both source and destination)",
			gpt.Table{SectorSize: 1024, Partitions: []gpt.Partition{
				{Id: testUuids[1], Type: gpt.PartType(testUuids[2]), FirstLBA: 10, LastLBA: 15},
				{Id: testUuids[3], Type: gpt.PartType(testUuids[2]), FirstLBA: 30, LastLBA: 40},
				{Id: testUuids[2], Type: gpt.PartType(testUuids[0]), FirstLBA: 50, LastLBA: 65},
			}},
			gpt.Table{SectorSize: 1024, Partitions: []gpt.Partition{
				{Id: testUuids[1], Type: gpt.PartType(testUuids[0]), FirstLBA: 10, LastLBA: 15},
				{Id: testUuids[2], Type: gpt.PartType(testUuids[2]), FirstLBA: 50, LastLBA: 65},
				{Id: testUuids[3], Type: gpt.PartType(testUuids[2]), FirstLBA: 30, LastLBA: 40},
			}},
			[]copyimg.Task{{Src: 30 * 1024, Dst: 30 * 1024, Size: 11 * 1024}},
			false},
		{"fails when copied source partition has FirstLBA < LastLBA",
			gpt.Table{SectorSize: 512, Partitions: []gpt.Partition{
				{Id: testUuids[1], Type: gpt.PartType(testUuids[2]), FirstLBA: 30, LastLBA: 29},
			}},
			gpt.Table{SectorSize: 512, Partitions: []gpt.Partition{
				{Id: testUuids[1], Type: gpt.PartType(testUuids[2]), FirstLBA: 13, LastLBA: 13},
			}},
			nil,
			true},
		{"doesn't fail when destination, uncopied source or empty partitions have FirstLBA < LastLBA",
			gpt.Table{SectorSize: 512, Partitions: []gpt.Partition{
				{Id: testUuids[1], Type: gpt.PartType(testUuids[2]), FirstLBA: 10, LastLBA: 15},
				{Id: testUuids[2], Type: gpt.PartType(testUuids[2]), FirstLBA: 50, LastLBA: 49}, // uncopied
				{Id: testUuids[3], Type: gpt.PartType(testUuids[0]), FirstLBA: 30, LastLBA: 29}, // empty
			}},
			gpt.Table{SectorSize: 512, Partitions: []gpt.Partition{
				{Id: testUuids[1], Type: gpt.PartType(testUuids[2]), FirstLBA: 10, LastLBA: 9},
				{Id: testUuids[3], Type: gpt.PartType(testUuids[2]), FirstLBA: 30, LastLBA: 29},
			}},
			[]copyimg.Task{{Src: 10 * 512, Dst: 10 * 512, Size: 6 * 512}},
			false},
	}

	for _, c := range cases {
		t.Run(c.label, func(t *testing.T) {
			act, actErr := copyimg.PlanFromGPTs(&c.dst, &c.src)
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

func TestCopyToSeeker(t *testing.T) {
	srcData := []byte{0x03, 0x88, 0x45, 0xAA, 0x88, 0x99, 0xFE, 0x72}
	dstOrigData := []byte{0x45, 0x77, 0x89, 0x82, 0x56, 0x32, 0xAA, 0xBC}
	cases := []struct {
		label      string
		tasks      []copyimg.Task
		exp        []byte
		shouldFail bool
	}{
		{"nothing copied",
			[]copyimg.Task{},
			dstOrigData, false},
		{"everything copied (using one task)",
			[]copyimg.Task{{Src: 0, Dst: 0, Size: 8}},
			srcData, false},
		{"part copied (at start)",
			[]copyimg.Task{{Src: 0, Dst: 0, Size: 3}},
			[]byte{0x03, 0x88, 0x45, 0x82, 0x56, 0x32, 0xAA, 0xBC}, false},
		{"parts copied (src == dst, multiple tasks, not at start)",
			[]copyimg.Task{{Src: 1, Dst: 1, Size: 1}, {Src: 3, Dst: 3, Size: 3}},
			[]byte{0x45, 0x88, 0x89, 0xAA, 0x88, 0x99, 0xAA, 0xBC}, false},
		{"parts copied (src == dst, multiple tasks, reverse order)",
			[]copyimg.Task{{Src: 3, Dst: 3, Size: 3}, {Src: 1, Dst: 1, Size: 1}},
			[]byte{0x45, 0x88, 0x89, 0xAA, 0x88, 0x99, 0xAA, 0xBC}, false},
		{"part copied (src != dst)",
			[]copyimg.Task{{Src: 1, Dst: 5, Size: 3}},
			[]byte{0x45, 0x77, 0x89, 0x82, 0x56, 0x88, 0x45, 0xAA}, false},
		{"everything copied (using multiple tasks)",
			[]copyimg.Task{
				{Src: 0, Dst: 0, Size: 3},
				{Src: 3, Dst: 3, Size: 1},
				{Src: 4, Dst: 4, Size: 4},
			},
			srcData, false},
		{"reverse source (single byte copies, random order)",
			[]copyimg.Task{
				{Src: 2, Dst: 5, Size: 1},
				{Src: 5, Dst: 2, Size: 1},
				{Src: 0, Dst: 7, Size: 1},
				{Src: 3, Dst: 4, Size: 1},
				{Src: 1, Dst: 6, Size: 1},
				{Src: 4, Dst: 3, Size: 1},
				{Src: 7, Dst: 0, Size: 1},
				{Src: 6, Dst: 1, Size: 1},
			},
			[]byte{0x72, 0xFE, 0x99, 0x88, 0xAA, 0x45, 0x88, 0x03}, false},
		{"overlapping src",
			[]copyimg.Task{{Src: 0, Dst: 0, Size: 3}, {Src: 2, Dst: 5, Size: 1}},
			nil, true},
		{"overlapping src (reversed)",
			[]copyimg.Task{{Src: 2, Dst: 5, Size: 1}, {Src: 0, Dst: 0, Size: 3}},
			nil, true},
		{"almost overlapping src",
			[]copyimg.Task{{Src: 0, Dst: 0, Size: 3}, {Src: 3, Dst: 5, Size: 1}},
			[]byte{0x03, 0x88, 0x45, 0x82, 0x56, 0xAA, 0xAA, 0xBC}, false},
		{"overlapping dst",
			[]copyimg.Task{{Src: 0, Dst: 0, Size: 3}, {Src: 4, Dst: 2, Size: 1}},
			nil, true},
		{"almost overlapping dst",
			[]copyimg.Task{{Src: 0, Dst: 0, Size: 3}, {Src: 4, Dst: 3, Size: 1}},
			[]byte{0x03, 0x88, 0x45, 0x88, 0x56, 0x32, 0xaa, 0xbc}, false},
		{"src out of range (starts outside)",
			[]copyimg.Task{{Src: 8, Dst: 0, Size: 1}},
			nil, true},
		{"dst out of range (starts outside)",
			[]copyimg.Task{{Src: 0, Dst: 8, Size: 1}},
			nil, true},
		{"src out of range (starts inside)",
			[]copyimg.Task{{Src: 5, Dst: 0, Size: 4}},
			nil, true},
		{"dst out of range (starts inside)",
			[]copyimg.Task{{Src: 0, Dst: 5, Size: 4}},
			nil, true},
	}

	for _, c := range cases {
		t.Run(c.label, func(t *testing.T) {
			origCopy := make([]byte, len(dstOrigData))
			copy(origCopy, dstOrigData)
			actBuf := NewWB(origCopy)

			progC := make(chan uint64)
			var prog []uint64
			progCloseC := make(chan struct{})
			go func() {
				for v := range progC {
					prog = append(prog, v)
				}
				close(progCloseC)
			}()

			actErr := copyimg.CopyToSeeker(actBuf, bytes.NewReader(srcData), c.tasks, progC)
			act := actBuf.buf

			select {
			case <-progCloseC:
				// do nothing
			case <-time.After(10 * time.Millisecond):
				t.Errorf("progress channel not closed")
			}

			if c.shouldFail {
				if actErr == nil {
					t.Errorf("got no error, want some error")
				}
			} else {
				if actErr != nil {
					t.Errorf("unexpected error: %v", actErr)
				}
				if !reflect.DeepEqual(act, c.exp) {
					t.Errorf("got %v, want %v", hex.EncodeToString(act), hex.EncodeToString(c.exp))
				}
				if len(prog) != len(c.tasks) {
					t.Errorf("got %v progress messages, want %v", len(prog), len(c.tasks))
				}
				var taskTot uint64
				for _, t := range c.tasks {
					taskTot += t.Size
				}
				var progTot uint64
				for _, v := range prog {
					progTot += v
				}
				if progTot != taskTot {
					t.Errorf("got progress messages for %v bytes (%+v), want %v bytes", progTot, prog, taskTot)
				}
			}

			for i, o := range dstOrigData {
				var hasTask bool
				for _, t := range c.tasks {
					if int(t.Dst) <= i && i < int(t.Dst+t.Size) {
						hasTask = true
					}
				}
				if hasTask {
					continue
				}
				if act[i] != o {
					t.Errorf("got act[%v]=0x%x, want 0x%x (original value, since there's no task for this byte)",
						i, act[i], o)
				}
				if len(c.exp) != 0 && c.exp[i] != o {
					t.Errorf("got exp[%v]=0x%x, want 0x%x (original value, since there's no task for this byte)",
						i, c.exp[i], o)
				}
			}
		})
	}
}

type WritableBuf struct {
	buf     []byte
	offset  int64
	lastErr error
}

func NewWB(data []byte) *WritableBuf {
	return &WritableBuf{buf: data}
}

func (t *WritableBuf) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case io.SeekStart:
		t.offset = offset
	case io.SeekCurrent:
		t.offset += offset
	case io.SeekEnd:
		return t.offset, fmt.Errorf("this Seeker does not support io.SeekEnd")
	default:
		return t.offset, fmt.Errorf("invalid whence argument to Seek: %v", whence)
	}
	if t.offset < 0 {
		e := fmt.Errorf("cannot seek to negative offset %v, capped to 0", t.offset)
		t.offset = 0
		return t.offset, e
	}
	return t.offset, nil
}

func (t *WritableBuf) Write(p []byte) (int, error) {
	lenb := len(t.buf)
	for i := range p {
		j := int(t.offset) + i
		if j >= lenb {
			return i, fmt.Errorf("write is (partially) of bounds")
		}
		t.buf[j] = p[i]
	}
	return len(p), nil
}

func TestSplitTasks(t *testing.T) {
	cases := []struct {
		label  string
		input  []copyimg.Task
		n      int
		outMin int // outMin <= len(output)
		outMax int // len(output) <= outMax
	}{
		{"no tasks", []copyimg.Task{}, 1, 0, 0},
		{"no tasks", []copyimg.Task{}, 10, 0, 0},
		{"one task", []copyimg.Task{
			{Src: 45, Dst: 22, Size: 10},
		}, 1, 1, 1},
		{"one task, size == n", []copyimg.Task{
			{Src: 45, Dst: 22, Size: 10},
		}, 10, 10, 10},
		{"multiple tasks, equal sizes, n == len(tasks)", []copyimg.Task{
			{Src: 45, Dst: 22, Size: 10},
			{Src: 99, Dst: 12, Size: 10},
			{Src: 3, Dst: 112, Size: 10},
		}, 3, 3, 3},
		{"one task, size exact multiple of n", []copyimg.Task{
			{Src: 45, Dst: 22, Size: 90},
		}, 10, 10, 10},
		{"one task, size just smaller than multiple of n", []copyimg.Task{
			{Src: 45, Dst: 22, Size: 89},
		}, 10, 9, 12},
		{"one task, size just larger than multiple of n", []copyimg.Task{
			{Src: 45, Dst: 22, Size: 91},
		}, 10, 8, 11},
		{"one task, size < n", []copyimg.Task{
			{Src: 45, Dst: 22, Size: 9},
		}, 10, 9, 9},
		{"one task, size == 0", []copyimg.Task{
			{Src: 45, Dst: 22, Size: 0},
		}, 10, 1, 1},
		{"multiple tasks, different sizes", []copyimg.Task{
			{Src: 45, Dst: 22, Size: 5},
			{Src: 3, Dst: 112, Size: 32},
			{Src: 99, Dst: 12, Size: 10},
		}, 4, 4, 6},
		{"multiple tasks, n < len(tasks)", []copyimg.Task{
			{Src: 45, Dst: 22, Size: 5},
			{Src: 3, Dst: 112, Size: 32},
			{Src: 99, Dst: 12, Size: 10},
			{Src: 5, Dst: 3, Size: 99},
		}, 2, 4, 4},
	}

	for _, c := range cases {
		t.Run(fmt.Sprintf("%v (n=%v)", c.label, c.n), func(t *testing.T) {
			input := make([]copyimg.Task, len(c.input))
			copy(input, c.input)
			act := copyimg.SplitTasks(c.input, c.n)
			if !reflect.DeepEqual(input, c.input) {
				t.Errorf("mutated input")
			}
			if c.outMin > len(act) || len(act) > c.outMax {
				t.Errorf("got len(act) = %v tasks, want %v <= len(act) <= %v",
					len(act), c.outMin, c.outMax)
			}
			var expBytes []copyimg.Task
			for _, t := range c.input {
				for i := uint64(0); i < t.Size; i++ {
					expBytes = append(expBytes, copyimg.Task{Src: t.Src + i, Dst: t.Dst + i, Size: 1})
				}
			}
			var actBytes []copyimg.Task
			for _, t := range act {
				for i := uint64(0); i < t.Size; i++ {
					actBytes = append(actBytes, copyimg.Task{Src: t.Src + i, Dst: t.Dst + i, Size: 1})
				}
			}
			if len(actBytes) != len(expBytes) {
				t.Errorf("planned to copy %v bytes, want %v", len(actBytes), len(expBytes))
			}
			for _, e := range expBytes {
				var found bool
				for _, a := range actBytes {
					if a == e {
						found = true
					}
				}
				if !found {
					t.Errorf("byte (Src: %v, Dst: %v) not copied", e.Src, e.Dst)
				}
			}
			// TODO compare all (src, dst) byte location pairs for act/exp
			/*
				var bufLen int
				for _, t := range c.input {
					s := int(t.Src + t.Size)
					d := int(t.Dst + t.Size)
					if bufLen < s {
						bufLen = s
					}
					if bufLen < d {
						bufLen = d
					}
				}
				srcBuf := make([]byte, bufLen)
				dstBuf := make([]byte, bufLen)
				for _, t := range c.input {
					for i := 0; i < t.Size; i++ {
						srcBuf[i+int(t.Src)] = 1
						srcBuf[i+int(t.Dst)] = 1
					}
				}
			*/
		})
	}
}
