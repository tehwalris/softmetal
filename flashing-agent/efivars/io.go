package efivars

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"strconv"
)

// The functions in this file read and write variables
// from the Linux efivars filesystem.
var efivarsPath = "/sys/firmware/efi/efivars/"
var efivarsPerms os.FileMode = 0644

// All standard EFI boot variables have this suffix.
// It is the EFI_GLOBAL_VARIABLE VendorGUID.
var efiGlobalSuffix = "-8be4df61-93ca-11d2-aa0d-00e098032b8c"

// ReadBootOrder reads the EFI boot order variable.
func ReadBootOrder() (*BootOrder, error) {
	d, e := ioutil.ReadFile(path.Join(efivarsPath, "BootOrder"+efiGlobalSuffix))
	if e != nil {
		return nil, e
	}
	return UnmarshalBootOrder(d)
}

// WriteBootOrder overwrites the EFI boot order variable.
func WriteBootOrder(ord BootOrder) error {
	p := path.Join(efivarsPath, "BootOrder"+efiGlobalSuffix)
	return ioutil.WriteFile(p, ord.Marhsal(), efivarsPerms)
}

// ReadBootEntries reads all existing EFI boot entries,
// even if they are not in the boot order.
func ReadBootEntries() (map[uint16]BootEntry, error) {
	r := regexp.MustCompile("^Boot([0-9A-F]{4})" + efiGlobalSuffix + "$")
	entries, e := ioutil.ReadDir(efivarsPath)
	if e != nil {
		return nil, e
	}
	out := make(map[uint16]BootEntry)
	for _, v := range entries {
		m := r.FindStringSubmatch(v.Name())
		if v.IsDir() || len(m) == 0 {
			continue
		}
		id, e := strconv.ParseUint(m[1], 16, 16)
		if e != nil {
			return nil, e
		}
		d, e := ioutil.ReadFile(path.Join(efivarsPath, v.Name()))
		if e != nil {
			return nil, e
		}
		parsed, e := UnmarshalBootEntry(d)
		if e != nil {
			return nil, e
		}
		out[uint16(id)] = *parsed
	}
	return out, nil
}

// WriteBootEntries overwrites or creates the specified EFI boot entries.
// It does not modify or delete any other boot entries.
func WriteBootEntries(entries map[uint16]BootEntry) error {
	for k, v := range entries {
		p := path.Join(efivarsPath, fmt.Sprintf("Boot%04X%s", k, efiGlobalSuffix))
		d, e := v.Marshal()
		if e != nil {
			return e
		}
		if e := ioutil.WriteFile(p, d, efivarsPerms); e != nil {
			return e
		}
	}
	return nil
}

// IsEFIBooted checks that the machine is booted in EFI mode and
// that the efivars filesystem is readable.
func IsEFIBooted() bool {
	_, e := ioutil.ReadDir(efivarsPath)
	return e == nil
}
