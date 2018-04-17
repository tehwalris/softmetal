package disk

import (
	"bytes"
	"io"
)

// WritePMBR writes a protective MBR to the given disk.
// It overwrites anything at the MBR location.
func WritePMBR(f io.WriteSeeker, sizeBytes uint64) error {
	// TODO also write "last logical block"
	// TODO size field is number of sectors!

	var sizeField = uint32(sizeBytes - 1)
	if sizeBytes > 0x100000000 {
		sizeField = 0xFFFFFFFF
	}

	mbr := make([]byte, 512)
	// http://www.uefi.org/sites/default/files/resources/UEFI%20Spec%202_6.pdf Section 5.2.3 Protective MBR
	// http://www.jonrajewski.com/data/Presentations/CEIC2013/Partition_Table_Documentation_Compressed.pdf
	mbr[448] = 0x02
	mbr[450] = 0xEE
	mbr[451] = 0xFF
	mbr[452] = 0xFF
	mbr[453] = 0xFF
	mbr[454] = 0x01
	mbr[458] = byte(sizeField & 0xFF)
	mbr[459] = byte((sizeField >> 8) & 0xFF)
	mbr[460] = byte((sizeField >> 16) & 0xFF)
	mbr[461] = byte((sizeField >> 24) & 0xFF)
	mbr[510] = 0x55
	mbr[511] = 0xAA

	if _, e := f.Seek(0, io.SeekStart); e != nil {
		return e
	}
	_, e := io.Copy(f, bytes.NewReader(mbr))
	return e
}
