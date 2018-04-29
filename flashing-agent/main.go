//go:generate protoc -I ../pb ../pb/flashing-supervisor.proto --go_out=plugins=grpc:../pb

package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/jaypipes/ghw"
	"github.com/rekby/gpt"

	"git.dolansoft.org/philippe/softmetal/flashing-agent/copyimg"
	"git.dolansoft.org/philippe/softmetal/flashing-agent/disk"
	"git.dolansoft.org/philippe/softmetal/flashing-agent/efivars"
	"git.dolansoft.org/philippe/softmetal/flashing-agent/partition"
	"git.dolansoft.org/philippe/softmetal/flashing-agent/superlog"
	pb "git.dolansoft.org/philippe/softmetal/pb"
	"google.golang.org/grpc"
)

var managerHP = flag.String("manager", "", "host and GRPC port of flashing manager (required)")

// gptBufferSize is the maximum number of bytes to load from
// the start of the image for extracting the GPT.
const gptBufferSize = 1000 * 1000

const initialRetryCount = 10
const initialRetryDelay = 5 * time.Second

func flash(logger *superlog.Logger, config *pb.FlashingConfig) error {
	if config.ImageConfig == nil {
		return fmt.Errorf("FlashingConfig.ImageConfig is required")
	}

	isEFI := efivars.IsEFIBooted()
	if !isEFI {
		log.Printf("WARNING: machine not booted in EFI mode or efivars filesystem unavailable")
	}

	bootEnt := config.ImageConfig.BootEntry
	if bootEnt == nil {
		log.Printf("WARING: no boot entry specified in ImageConfig")
	} else if !isEFI {
		return fmt.Errorf("machine must be EFI booted to set boot entries")
	}

	logger.Logf("using disk with serial %v", config.TargetDiskCombinedSerial)
	diskF, diskInfo, e := disk.OpenBySerial(config.TargetDiskCombinedSerial)
	if e != nil {
		return e
	}
	defer func() {
		if e := diskF.Close(); e != nil {
			logger.Logf("while closing disk: %v", e)
		}
	}()

	table, didCreateGpt, e := disk.GetOrCreateGpt(diskF, diskInfo)
	if e != nil {
		return e
	}
	if didCreateGpt {
		logger.Logf("using new blank GPT table (no table found on disk)")
	} else {
		logger.Logf("using existing GPT table from disk")
	}
	partition.PrintTable(table, logger, "Old GPT table from disk")

	imgURL := config.ImageConfig.Url
	logger.Logf("using image: %v", imgURL)
	imgRes, e := http.Get(imgURL)
	if e != nil {
		return fmt.Errorf("while getting image (first): %v", e)
	}

	var imgBuf bytes.Buffer
	imgBuf.Grow(gptBufferSize)
	_, e = io.CopyN(&imgBuf, imgRes.Body, gptBufferSize)
	if e != nil && e != io.EOF {
		return fmt.Errorf("while buffering: %v", e)
	}
	if e := imgRes.Body.Close(); e != nil {
		return fmt.Errorf("while closing image: %v", e)
	}

	imgSS := config.ImageConfig.SectorSize
	if imgSS < 512 {
		return fmt.Errorf("image has invalid sector size: %v", imgSS)
	}
	imgBufR := bytes.NewReader(imgBuf.Bytes())
	if _, e := imgBufR.Seek(int64(imgSS), io.SeekStart); e != nil {
		return fmt.Errorf("while seeking buffer: %v", e)
	}
	imgTable, e := gpt.ReadTable(imgBufR, uint64(imgSS))
	if e != nil {
		return fmt.Errorf("while reading GPT from image: %v", e)
	}
	partition.PrintTable(&imgTable, logger, "GPT table from image")

	pers := make([]pb.FlashingConfig_Partition, len(config.PersistentPartitions))
	for i, p := range config.PersistentPartitions {
		pers[i] = *p
	}
	if e := copyimg.MergeGpt(table, &imgTable, pers); e != nil {
		return fmt.Errorf("while merging GPT: %v", e)
	}
	partition.PrintTable(table, logger, "Merged GPT")

	if e := table.Write(diskF); e != nil {
		return fmt.Errorf("while writing disk-start GPT: %v", e)
	}
	if e := table.CreateOtherSideTable().Write(diskF); e != nil {
		return fmt.Errorf("while writing disk-end GPT: %v", e)
	}
	if e := disk.WritePMBR(diskF, diskInfo.SectorSizeBytes, diskInfo.SizeBytes); e != nil {
		return fmt.Errorf("while writing protective MBR: %v", e)
	}

	cpTasks, e := copyimg.PlanFromGPTs(table, &imgTable)
	if e != nil {
		return fmt.Errorf("while planning copy: %v", e)
	}
	cpTasks = copyimg.SplitTasks(cpTasks, 100)
	imgRes, e = http.Get(imgURL)
	if e != nil {
		return fmt.Errorf("while getting image (second): %v", e)
	}

	var total uint64
	for _, t := range cpTasks {
		total += t.Size
	}
	progC := make(chan uint64, 50)
	var cur uint64
	go func() {
		for v := range progC {
			cur += v
			logger.Progress(float32(cur) / float32(total))
		}
	}()

	fmt.Sprintf("WARNING: skipping copy")
	/*
		if e := copyimg.CopyToSeeker(diskF, imgRes.Body, cpTasks, progC); e != nil {
			return fmt.Errorf("during main copy operation: %v", e)
		}
	*/
	if e := imgRes.Body.Close(); e != nil {
		log.Printf("WARNING: failed to close image (%v) after copy: %v", imgURL, e)
	}

	if bootEnt != nil {
		logger.Logf("configuring boot entries")
		oldOrd, e := efivars.ReadBootOrder()
		if e != nil {
			return fmt.Errorf("while reading boot order: %v", e)
		}
		oldEnts, e := efivars.ReadBootEntries()
		if e != nil {
			return fmt.Errorf("while reading boot entries: %v", e)
		}
		logger.Logf("old boot order: %04X", oldOrd)
		logger.Logf("old boot entries:")
		for k, v := range oldEnts {
			logger.Logf(" %04X %v", k, v.Description)
		}

		newEnt, e := efivars.NewBootEntry(bootEnt.Path, table.Partitions)
		if e != nil {
			return fmt.Errorf("while creating boot entry in-memory: %v", e)
		}

		up, e := efivars.PlanUpdate(*oldOrd, oldEnts, *newEnt)
		logger.Logf("boot config changes: %+v", up)

		if e != nil {
			return fmt.Errorf("while planning update: %v", e)
		}
		if e := efivars.WriteBootEntries(up.Write); e != nil {
			return fmt.Errorf("while writing boot entries: %v", e)
		}
		if e := efivars.WriteBootOrder(up.Order); e != nil {
			return fmt.Errorf("while writing boot order: %v", e)
		}
	}

	// TODO alignment of merged partitions!
	// TODO check that image is equal on second read

	return nil
}

func logSysinfo(logger *superlog.Logger) error {
	bl, e := ghw.Block()
	if e != nil {
		return e
	}
	logger.Logf("block info: %v", bl.String())
	logger.Logf("  disks: ")
	for _, v := range bl.Disks {
		logger.Logf("    %v", v.String())
	}
	n, e := ghw.Network()
	if e != nil {
		return e
	}
	logger.Logf("network info: %v", n.String())
	logger.Logf("  NICs: ")
	for _, v := range n.NICs {
		logger.Logf("    %v", v.String())
	}
	return nil
}

func listen(logger *superlog.Logger) (pb.PowerControlType, error) {
	var ok bool
	var defaultPowerControl pb.PowerControlType

	var conn *grpc.ClientConn
	var e error
	for i := 0; i < initialRetryCount; i++ {
		logger.Logf("connecting to manager: %v (attempt #%v)", *managerHP, i+1)
		conn, e = grpc.Dial(*managerHP, grpc.WithInsecure(), grpc.WithBlock())
		if e == nil {
			break
		}
		logger.Logf("while connecting: %v", e)
		time.Sleep(initialRetryDelay)
	}
	if e != nil {
		return defaultPowerControl, e
	}

	c := pb.NewFlashingSupervisorClient(conn)
	cmd, e := c.GetCommand(context.Background(), &pb.Empty{})
	if e != nil {
		return defaultPowerControl, e
	}
	logger.AttachSupervisor(c, cmd.SessionId)
	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(error); ok {
				logger.Logf("recovered from panic: \n%v", e)
			} else {
				logger.Logf("recovered from panic: \n%v", r)
			}
		}
		logger.DetachSupervisor()
		_, e := c.RecordFinished(context.Background(), &pb.RecordFinishedRequest{
			SessionId: cmd.SessionId,
			Ok:        ok,
		})
		if e != nil {
			logger.Logf("failed to report finished: %v", e)
		}
	}()

	if e = logSysinfo(logger); e != nil {
		logger.Logf("failed to get system info: %v", e)
	}

	if e = flash(logger, cmd.Config); e != nil {
		// Log this here so that supervisor gets it, since it will be detatched later.
		logger.Logf("flashing error: %v", e)
		return cmd.PowerOnCompletion, e
	}

	ok = true
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
