package disk

import (
	"bytes"
	"io"
)

// WritePMBR writes a protective MBR to the given disk.
// It overwrites anything at the MBR location.
func WritePMBR(f io.WriteSeeker, sizeBytes uint64) error {
	mbr := make([]byte, 512)
	// http://www.jonrajewski.com/data/Presentations/CEIC2013/Partition_Table_Documentation_Compressed.pdf
	mbr[448] = 0x02
	mbr[450] = 0xEE
	mbr[451] = 0xFF
	mbr[452] = 0xFF
	mbr[453] = 0xFF
	mbr[454] = 0x01
	mbr[458] = 0xFF
	mbr[459] = 0xFF
	mbr[460] = 0xFF
	mbr[461] = 0xFF
	mbr[510] = 0x55
	mbr[511] = 0xAA

	if _, e := f.Seek(0, io.SeekStart); e != nil {
		return e
	}
	_, e := io.Copy(f, bytes.NewReader(mbr))
	return e
}
