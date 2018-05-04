package main

import pb "git.dolansoft.org/philippe/softmetal/pb"

var machines = map[string]pb.FlashingConfig{
	"okne": {
		TargetDiskCombinedSerial: "Samsung_SSD_950_PRO_512GB_S2GMNCAGB17541H",
		PersistentPartitions: []*pb.FlashingConfig_Partition{
			{
				PartUuid: "C277D159-0819-4E41-9F3D-B22143CECFCD",
				GptType:  "E6D6D379-F507-44C2-A23C-238F2A3DF928",
				Size:     927068815 * 512,
			},
		},
	},
	"zaba": {
		TargetDiskCombinedSerial: "Samsung_SSD_840_EVO_500GB_S1DHNSAF443735Z",
	},
}
