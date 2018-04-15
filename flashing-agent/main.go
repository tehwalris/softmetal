//go:generate protoc -I ../pb ../pb/flashing-supervisor.proto --go_out=plugins=grpc:../pb

package main

import (
	"context"
	"flag"
	"log"
	"os"

	"git.dolansoft.org/philippe/softmetal/flashing-agent/disk"
	"git.dolansoft.org/philippe/softmetal/flashing-agent/partition"
	"git.dolansoft.org/philippe/softmetal/flashing-agent/superlog"
	pb "git.dolansoft.org/philippe/softmetal/pb"
	"google.golang.org/grpc"
)

var managerHP = flag.String("manager", "", "host and GRPC port of flashing manager (required)")

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
		logger.Logf("Using new blank GPT table (no table found on disk).")
	} else {
		logger.Logf("Using existing GPT table from disk.")
	}

	partition.PrintTable(table, logger)
	// merge gpt
	// write gpt

	return nil
}

func listen(logger *superlog.Logger) (pb.PowerControlType, error) {
	var defaultPowerControl pb.PowerControlType

	logger.Logf("connecting to manager: %v", *managerHP)
	conn, e := grpc.Dial(*managerHP, grpc.WithInsecure())
	if e != nil {
		return defaultPowerControl, e
	}

	c := pb.NewFlashingSupervisorClient(conn)
	cmd, e := c.GetCommand(context.Background(), &pb.Empty{})
	if e != nil {
		return defaultPowerControl, e
	}
	logger.AttachSupervisor(c, cmd.SessionId)
	defer logger.DetachSupervisor()

	if e = flash(logger, cmd.Config); e != nil {
		// log this here so that supervisor gets it, since it will be detatched later
		logger.Logf("flashing error: %v", e)
	}
	return cmd.PowerOnCompletion, e
}

func main() {
	flag.Parse()
	if *managerHP == "" {
		log.Fatalf("missing required arguments")
	}

	logger := superlog.New(log.New(os.Stderr, "", log.LstdFlags))
	powerControl, e := listen(logger)
	logger.Logf("flashing error: %v", e)
	logger.Logf("would power control: %v", powerControl)
}
