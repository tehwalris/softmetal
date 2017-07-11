package main

import (
	"context"
	"log"
	"os"

	"google.golang.org/grpc"

	"git.dolansoft.org/philippe/softmetal/flashing-agent/disk"
	"git.dolansoft.org/philippe/softmetal/flashing-agent/partition"
	"git.dolansoft.org/philippe/softmetal/flashing-agent/superlog"
	pb "git.dolansoft.org/philippe/softmetal/pb"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	logger := superlog.New(log.New(os.Stderr, "", log.LstdFlags))
	managerAddress := "localhost:5051"
	targetDiskSerial := "TOSHIBA_THNSFJ256GCSU_46KS117IT8LW"

	logger.Logf("Connecting to manager (%v).", managerAddress)
	conn, e := grpc.Dial(managerAddress, grpc.WithInsecure())
	check(e)

	c := pb.NewFlashingSupervisorClient(conn)
	superviseClient, e := c.Supervise(context.Background())
	defer superviseClient.CloseSend()
	logger.AttachSupervisor(&superviseClient)
	check(e)
	e = superviseClient.Send(&pb.FlashingStatusUpdate{
		Update: &pb.FlashingStatusUpdate_StateChange_{
			StateChange: &pb.FlashingStatusUpdate_StateChange{
				State: pb.FlashingStatusUpdate_StateChange_READY_IDLE,
			},
		},
	})
	check(e)

	logger.Logf("Using disk with serial %v.", targetDiskSerial)
	f, diskInfo, e := disk.OpenBySerial(targetDiskSerial)
	check(e)
	table, didCreateGpt, e := disk.GetOrCreateGpt(f, diskInfo)
	check(e)
	if didCreateGpt {
		logger.Log("Using new blank GPT table (no table found on disk).")
	} else {
		logger.Log("Using existing GPT table from disk.")
	}

	partition.PrintTable(table, logger)
	// merge gpt
	// write gpt
}
