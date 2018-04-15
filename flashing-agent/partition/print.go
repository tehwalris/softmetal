package partition

import (
	"git.dolansoft.org/philippe/softmetal/flashing-agent/superlog"
	"github.com/rekby/gpt"
)

func PrintTable(table *gpt.Table, logger *superlog.Logger) {
	logger.Logf("GPT table:")
	logger.Logf("  Sector size: %v", table.SectorSize)
	logger.Logf("  Header: %v", table.Header)
	logger.Logf("  Partitions:")
	consecutiveEmpty := 0
	for i, p := range table.Partitions {
		if p.IsEmpty() {
			consecutiveEmpty += 1
			continue
		} else if consecutiveEmpty > 0 {
			logger.Logf("    (%v empty)\n", consecutiveEmpty)
			consecutiveEmpty = 0
		}
		logger.Logf("    %03d: %12d - %-12d %v (type %v)\n",
			i, p.FirstLBA, p.LastLBA, p.Id.String(), p.Type.String())
	}
	if consecutiveEmpty > 0 {
		logger.Logf("    (%v empty)\n", consecutiveEmpty)
	}
}
