package efivars

import (
	"fmt"
	"math"
	"strings"

	"github.com/rekby/gpt"
)

var softmetalEntryDesc = "Softmetal (boot from disk)"
var espGUIDStr = "C12A7328-F81F-11D2-BA4B-00A0C93EC93B"

// Update specifies a set of modifications to EFI boot variables.
type Update struct {
	Write map[uint16]BootEntry // Create or update these variables
	Order BootOrder            // Set BootOrder to this
}

// PlanUpdate creates or updates the softmetal boot entry to match newEntry and adjusts
// the boot order to have that entry load first. PlanUpdate does not write any EFI variables itself.
// The newEntry argument should be a fully configured boot entry with an empty description.
// If PlanUpdate plans to create a new boot entry, it uses the lowest free ID (Update.Write key).
// Only the Description field is required in the supplied boot entries (oldEntries).
// PlanUpdate recognizes the softmetal boot entry by description only.
// If there are multiple softmetal boot entries, PlanUpdate will fail.
func PlanUpdate(oldOrd BootOrder, oldEntries map[uint16]BootEntry, newEntry BootEntry) (*Update, error) {
	if newEntry.Description != "" {
		return nil, fmt.Errorf("newEntry.Description must be empty, got %v", newEntry.Description)
	}
	newEntry.Description = softmetalEntryDesc

	var usedByID = make(map[uint16]struct{})
	for k := range oldEntries {
		usedByID[k] = struct{}{}
	}

	var foundID bool
	var newID uint16
	for k, v := range oldEntries {
		if v.Description == softmetalEntryDesc {
			if foundID {
				return nil, fmt.Errorf("found muliple existing softmetal boot entries")
			}
			foundID = true
			newID = k
		}
	}
	for i := 0; !foundID && i <= math.MaxUint16; i++ {
		if _, prs := usedByID[uint16(i)]; !prs {
			foundID = true
			newID = uint16(i)
		}
	}
	if !foundID {
		return nil, fmt.Errorf("no free boot entry IDs (%v boot entries exist)", len(oldEntries))
	}

	newOrd := BootOrder{newID}
	for _, v := range oldOrd {
		if v != newID {
			newOrd = append(newOrd, v)
		}
	}

	return &Update{Write: map[uint16]BootEntry{newID: newEntry}, Order: newOrd}, nil
}

// NewBootEntry creates a boot entry for softmetal by finding required
// information about the ESP partition on disk. The partitions argument
// should contain the final partitions stored on the disk, not the ones in the image.
// NewBootEntry fills all fields of BootEntry except Description.
// If there are n != 1 ESP partitions, NewBootEntry fails.
// ESP partitions are detected by their type: EFI System (see espGUIDStr).
func NewBootEntry(path string, partitions []gpt.Partition) (*BootEntry, error) {
	// TODO need to confirm that parition numbers on linux increment with empty entries in partition table

	var targetIdx int
	var found int
	for i, p := range partitions {
		if strings.ToLower(p.Type.String()) == strings.ToLower(espGUIDStr) {
			targetIdx = i
			found++
		}
	}
	if found != 1 {
		return nil, fmt.Errorf("found %v ESP partitions on disk, want exactly 1", found)
	}
	p := partitions[targetIdx]
	return &BootEntry{
		Path:            path,
		PartitionGUID:   p.Id,
		PartitionNumber: uint32(targetIdx + 1),
		PartitionStart:  p.FirstLBA,
		PartitionSize:   p.LastLBA - p.FirstLBA + 1,
	}, nil
}
