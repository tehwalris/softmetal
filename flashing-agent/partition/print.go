package partition

import (
	"fmt"

	"github.com/rekby/gpt"
)

func PrintTable(table *gpt.Table) {
	fmt.Printf("Table:\n  Sector size: %v\n  Header: %+v\n  Partitions:\n",
		table.SectorSize, table.Header)
	consecutiveEmpty := 0
	for i, p := range table.Partitions {
		if p.IsEmpty() {
			consecutiveEmpty += 1
			continue
		} else if consecutiveEmpty > 0 {
			fmt.Printf("    (%v empty)\n", consecutiveEmpty)
			consecutiveEmpty = 0
		}
		fmt.Printf("    %03d: %12d - %-12d %v (type %v)\n",
			i, p.FirstLBA, p.LastLBA, p.Id.String(), p.Type.String())
	}
	if consecutiveEmpty > 0 {
		fmt.Printf("    (%v empty)\n", consecutiveEmpty)
	}
}
