package main

import (
	"context"
	"flag"
	"log"
	"net"
	"net/http"
	"sync/atomic"

	pb "git.dolansoft.org/philippe/softmetal/pb"
	"google.golang.org/grpc"
)

var httpListen = flag.String("http-listen", ":8080", "address and port to listen for HTTP on")
var grpcListen = flag.String("grpc-listen", ":6781", "address and port to listen for GRPC on")
var imageURL = flag.String("image", "", "URL of the disk image to flash (required)")

type supervisorServer struct {
	agentIDCounter uint64
}

func (s *supervisorServer) GetCommand(ctx context.Context, r *pb.Empty) (*pb.FlashingCommand, error) {
	sid := atomic.AddUint64(&s.agentIDCounter, 1)
	log.Printf("SUPER %v: agent connected", sid)
	return &pb.FlashingCommand{
		SessionId: sid,
		Config: &pb.FlashingConfig{
			TargetDiskCombinedSerial: "QEMU_HARDDISK_QM00001",
			ImageConfig: &pb.FlashingConfig_ImageConfig{
				Url:        *imageURL,
				SectorSize: 512,
			},
		},
		PowerOnCompletion: pb.PowerControlType_POWER_OFF,
	}, nil
}

func (s *supervisorServer) RecordLog(ctx context.Context, r *pb.RecordLogRequest) (*pb.Empty, error) {
	log.Printf("AGENT %v LOG: %v", r.SessionId, r.Log)
	return &pb.Empty{}, nil
}

func (s *supervisorServer) RecordProgress(ctx context.Context, r *pb.RecordProgressRequest) (*pb.Empty, error) {
	log.Printf("AGENT %v PROGRESS: %v", r.SessionId, r.Progress)
	return &pb.Empty{}, nil
}

func (s *supervisorServer) RecordFinished(ctx context.Context, r *pb.RecordFinishedRequest) (*pb.Empty, error) {
	log.Printf("AGENT %v FINISHED: ok: %v", r.SessionId, r.Ok)
	return &pb.Empty{}, nil
}

func handleHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("walrus!"))
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	flag.Parse()
	if *imageURL == "" {
		log.Fatalf("missing required arguments, see -help")
	}

	lis, e := net.Listen("tcp", *grpcListen)
	check(e)
	s := grpc.NewServer()
	pb.RegisterFlashingSupervisorServer(s, &supervisorServer{})

	http.HandleFunc("/", handleHTTP)
	http.HandleFunc("/agent-linux-amd64", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "../flashing-agent/flashing-agent")
	})

	log.Printf("listening (HTTP on %v, GRPC on %v)", *httpListen, *grpcListen)
	go func() { check(http.ListenAndServe(*httpListen, nil)) }()
	check(s.Serve(lis))
}
