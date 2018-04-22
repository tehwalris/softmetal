package disk

import (
	"bytes"
	"fmt"
	"io"
	"math"
)

// WritePMBR writes a protective MBR to the given disk.
// It overwrites the whole first logical block (sectorSizeBytes) of the disk.
func WritePMBR(f io.WriteSeeker, sectorSizeBytes uint64, diskSizeBytes uint64) error {
	// Useful reference pages:
	// http://www.uefi.org/sites/default/files/resources/UEFI%20Spec%202_6.pdf Section 5.2.3 Protective MBR
	// http://www.jonrajewski.com/data/Presentations/CEIC2013/Partition_Table_Documentation_Compressed.pdf
	// http://thestarman.pcministry.com/asm/mbr/PartTables.htm

	if diskSizeBytes == 0 || sectorSizeBytes < 512 || diskSizeBytes%sectorSizeBytes != 0 {
		return fmt.Errorf("invalid or incompatible disk and sector sizes (disk: %v, sector: %v)", diskSizeBytes, sectorSizeBytes)
	}

	// See partRecord for details.
	sizeInLBA := diskSizeBytes/sectorSizeBytes - 1
	if sizeInLBA > math.MaxUint32 {
		sizeInLBA = math.MaxUint32
	}

	// First (and only) partition record. Names and comments taken from
	// "Table 17. Protective MBR Partition Record protecting the entire disk"
	// in http://www.uefi.org/sites/default/files/resources/UEFI%20Spec%202_6.pdf
	partRecord := []byte{
		// BootIndicator
		// Set to 0x00 to indicate a non-bootable partition. If set to
		// any value other than 0x00 the behavior of this flag on
		// non-UEFI systems is undefined. Must be ignored by
		// UEFI implementations.
		0x00,

		// StartingCHS
		// Set to 0x000200, corresponding to the Starting LBA
		// field.
		0x00,
		0x02,
		0x00,

		// OSType
		// Set to 0xEE (i.e., GPT Protective)
		0xEE,

		// EndingCHS
		// Set to the CHS address of the last logical block on the
		// disk. Set to 0xFFFFFF if it is not possible to represent
		// the value in this field.
		// HACK
		// Not even trying to calculate this. This field is not
		// relevant at all for disks larger than about 8GB. Even for
		// smaller disks, it is very likely that no useful software
		// cares about this field (instead uses only StartingLBA and
		// SizeInLBA fields). Calculating EndingCHS would also require
		// knowing the disk geometry, which is hard to reliably find
		// in a golang application.
		0xFF,
		0xFF,
		0xFF,

		// StartingLBA
		// Set to 0x00000001 (i.e., the LBA of the GPT Partition
		// Header).
		0x01,
		0x00,
		0x00,
		0x00,

		// SizeInLBA
		// Set to the size of the disk minus one. Set to
		// 0xFFFFFFFF if the size of the disk is too large to be
		// represented in this field.
		byte(sizeInLBA & 0xFF),
		byte((sizeInLBA >> 8) & 0xFF),
		byte((sizeInLBA >> 16) & 0xFF),
		byte((sizeInLBA >> 24) & 0xFF),
	}

	mbr := make([]byte, sectorSizeBytes)
	for i, v := range partRecord {
		mbr[446+i] = v
	}
	// MBR Signature
	mbr[510] = 0x55
	mbr[511] = 0xAA

	if _, e := f.Seek(0, io.SeekStart); e != nil {
		return e
	}
	_, e := io.Copy(f, bytes.NewReader(mbr))
	return e
}
