package partition

import (
	pb "git.dolansoft.org/philippe/softmetal/pb"
	"github.com/rekby/gpt"
)

func partitionSectorSize(p *pb.FlashingConfig_Partition, sectorSize uint64) uint64 {
	size := p.Size / sectorSize
	if p.Size%sectorSize != 0 {
		size += 1
	}
	return size
}

// Does not set FirstLBA and LastLBA
func convertPartition(p *pb.FlashingConfig_Partition) (*gpt.Partition, error) {
	var gptType gpt.PartType
	var id gpt.Guid
	res, e := StringToGuid(p.GptType)
	if e != nil {
		return nil, e
	}
	gptType = gpt.PartType(res)
	res, e = StringToGuid(p.PartUuid)
	if e != nil {
		return nil, e
	}
	id = gpt.Guid(res)
	diskP := gpt.Partition{
		Id:   id,
		Type: gptType,
		// TODO part name and other fields
	}
	return &diskP, nil
}

func AddPersistentIfMissing(table *gpt.Table, p *pb.FlashingConfig_Partition) error {
	if !ContainsId(table.Partitions, &p.PartUuid) {
		size := partitionSectorSize(p, table.SectorSize)
		diskP, e := convertPartition(p)
		if e != nil {
			return e
		}
		diskP.FirstLBA = 0
		diskP.LastLBA = size - 1
		if e := AddFindSpace(table, diskP, End); e != nil {
			return e
		}
	}
	return nil
}
