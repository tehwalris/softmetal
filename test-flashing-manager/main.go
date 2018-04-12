package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"sync/atomic"

	pb "git.dolansoft.org/philippe/softmetal/pb"
	"google.golang.org/grpc"
)

type supervisorServer struct {
	agentIdCounter uint64
}

func (s *supervisorServer) Supervise(
	client pb.FlashingSupervisor_SuperviseServer,
) error {
	agentId := atomic.AddUint64(&s.agentIdCounter, 1)
	supervisionLog := log.New(os.Stderr, fmt.Sprintf("SUPER %v: ", agentId), log.LstdFlags)
	agentLog := log.New(os.Stderr, fmt.Sprintf("AGENT %v: ", agentId), log.LstdFlags)
	supervisionLog.Println("Agent connected")
	client.Send(&pb.FlashingCommand{
		Config: &pb.FlashingConfig{
			TargetDiskCombinedSerial: "TOSHIBA_THNSFJ256GCSU_46KS117IT8LW",
			PersistentPartitions:     []*pb.FlashingConfig_Partition{},
		},
		PowerOnCompletion: pb.PowerControlType_POWER_OFF,
	}) // TODO
	for {
		in, e := client.Recv()
		if e == io.EOF {
			supervisionLog.Println("Agent disconnected")
			return nil
		}
		if e != nil {
			return e
		}
		if genericLog := in.GetGenericLog(); genericLog != nil {
			agentLog.Println(genericLog.Log)
		} else {
			supervisionLog.Printf("Got message: %+v", in)
		}
	}
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	listenAddress := ":5051"
	log.Printf("Listening on %v.", listenAddress)
	lis, e := net.Listen("tcp", listenAddress)
	check(e)

	s := grpc.NewServer()
	pb.RegisterFlashingSupervisorServer(s, &supervisorServer{})
	e = s.Serve(lis)
	check(e)
}
