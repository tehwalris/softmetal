//go:generate protoc -I ../pb ../pb/flashing-supervisor.proto --go_out=plugins=grpc:../pb

package main

import (
	"context"
	"io"
	"log"
	"os"

	"git.dolansoft.org/philippe/softmetal/flashing-agent/disk"
	"git.dolansoft.org/philippe/softmetal/flashing-agent/partition"
	"git.dolansoft.org/philippe/softmetal/flashing-agent/superlog"
	pb "git.dolansoft.org/philippe/softmetal/pb"
	"google.golang.org/grpc"
)

func flash(logger *superlog.Logger, config *pb.FlashingConfig) error {
	targetDiskSerial := "TOSHIBA_THNSFJ256GCSU_46KS117IT8LW"
	// TODO load and check inputs

	logger.Logf("Using disk with serial %v.", targetDiskSerial)
	f, diskInfo, e := disk.OpenBySerial(targetDiskSerial)
	if e != nil {
		return e
	}
	table, didCreateGpt, e := disk.GetOrCreateGpt(f, diskInfo)
	if e != nil {
		return e
	}
	if didCreateGpt {
		logger.Log("Using new blank GPT table (no table found on disk).")
	} else {
		logger.Log("Using existing GPT table from disk.")
	}

	partition.PrintTable(table, logger)
	// merge gpt
	// write gpt

	return nil
}

func listen(logger *superlog.Logger) (pb.PowerControlType, error) {
	managerAddress := "localhost:5051"
	defaultPowerControl := pb.PowerControlType_REBOOT

	logger.Logf("Connecting to manager (%v).", managerAddress)
	conn, e := grpc.Dial(managerAddress, grpc.WithInsecure())
	if e != nil {
		return defaultPowerControl, e
	}

	c := pb.NewFlashingSupervisorClient(conn)
	superviseClient, e := c.Supervise(context.Background())
	if e != nil {
		return defaultPowerControl, e
	}
	defer superviseClient.CloseSend()
	defer logger.DetachSupervisor()
	logger.AttachSupervisor(&superviseClient)

	in, e := superviseClient.Recv()
	if e == io.EOF {
		logger.Log("Manager disconnected")
		logger.DetachSupervisor()
		return defaultPowerControl, nil
	}
	if e != nil {
		return defaultPowerControl, e
	}
	e, powerControl := flash(logger, in.Config), in.PowerOnCompletion
	superviseClient.CloseSend()
	superviseClient.Recv()
	return powerControl, e
}

func main() {
	logger := superlog.New(log.New(os.Stderr, "", log.LstdFlags))
	powerControl, e := listen(logger)
	if e != nil {
		logger.Logf("Exited with error: %v", e)
	} else {
		logger.Log("Exited cleanly.")
	}
	logger.Logf("Would power control: %v", powerControl) // TODO
}
