package partition

import (
	"log"

	"github.com/rekby/gpt"
)

func PrintTable(table *gpt.Table, logger *log.Logger) {
	logger.Println("GPT table:")
	logger.Printf("  Sector size: %v", table.SectorSize)
	logger.Printf("  Header: %v", table.Header)
	logger.Println("  Partitions:")
	consecutiveEmpty := 0
	for i, p := range table.Partitions {
		if p.IsEmpty() {
			consecutiveEmpty += 1
			continue
		} else if consecutiveEmpty > 0 {
			logger.Printf("    (%v empty)\n", consecutiveEmpty)
			consecutiveEmpty = 0
		}
		logger.Printf("    %03d: %12d - %-12d %v (type %v)\n",
			i, p.FirstLBA, p.LastLBA, p.Id.String(), p.Type.String())
	}
	if consecutiveEmpty > 0 {
		logger.Printf("    (%v empty)\n", consecutiveEmpty)
	}
}
