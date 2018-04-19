package copyimg_test

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"reflect"
	"testing"

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
			act, actErr := copyimg.PlanFromGPTs(&c.src, &c.dst)
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
			actErr := copyimg.CopyToSeeker(actBuf, bytes.NewReader(srcData), c.tasks)
			act := actBuf.buf
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
