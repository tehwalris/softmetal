package main

import (
	"io"
	"log"
	"net"

	pb "git.dolansoft.org/philippe/softmetal/pb"
	"google.golang.org/grpc"
)

type supervisorServer struct{}

func (s *supervisorServer) Supervise(
	client pb.FlashingSupervisor_SuperviseServer,
) error {
	log.Print("Flashing agent connected")
	// agentLog := log.New(os.Stderr, "AGENT: ", log.LstdFlags)
	for {
		in, e := client.Recv()
		if e == io.EOF {
			return nil
		}
		if e != nil {
			return e
		}
		log.Printf("Got message: %+v", in)
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
