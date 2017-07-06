package partition

import (
	"strings"

	pb "git.dolansoft.org/philippe/softmetal/pb"

	"github.com/rekby/gpt"
)

func sizeBytes(part *gpt.Partition, sectorSize uint64) uint64 {
	return (part.LastLBA - part.FirstLBA + 1) * sectorSize
}

func Matches(
	real *gpt.Partition, search *pb.FlashingConfig_Partition, sectorSize uint64,
) bool {
	return strings.ToLower(real.Id.String()) == strings.ToLower(search.PartUuid) &&
		strings.ToLower(real.Type.String()) == strings.ToLower(search.GptType) &&
		sizeBytes(real, sectorSize) == search.Size
}

func MatchesId(
	real *gpt.Partition, partUuid *string,
) bool {
	return strings.ToLower(real.Id.String()) == strings.ToLower(*partUuid)
}

func ContainsId(
	partitions []gpt.Partition, partUuid *string,
) bool {
	for i, _ := range partitions {
		p := &partitions[i]
		if !p.IsEmpty() && MatchesId(p, partUuid) {
			return true
		}
	}
	return false
}
