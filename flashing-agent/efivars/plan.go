package efivars

import (
	"fmt"
	"math"
)

var softmetalEntryDesc = "Softmetal (boot from disk)"

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
	for i := uint16(0); !foundID && i <= math.MaxUint16; i++ {
		if _, prs := usedByID[i]; !prs {
			foundID = true
			newID = uint16(i)
		}
	}

	newOrd := BootOrder{newID}
	for _, v := range oldOrd {
		if v != newID {
			newOrd = append(newOrd, v)
		}
	}

	return &Update{Write: map[uint16]BootEntry{newID: newEntry}, Order: newOrd}, nil
}
