package superlog

import (
	"fmt"
	"log"

	pb "git.dolansoft.org/philippe/softmetal/pb"
)

type Logger struct {
	baseLogger      *log.Logger
	superviseClient *pb.FlashingSupervisor_SuperviseClient
}

func New(baseLogger *log.Logger) *Logger {
	return &Logger{
		baseLogger:      baseLogger,
		superviseClient: nil,
	}
}

func (l *Logger) trySendToSupervisor(msg string) {
	c := l.superviseClient
	if c != nil {
		(*c).Send(&pb.FlashingStatusUpdate{
			Update: &pb.FlashingStatusUpdate_GenericLog_{
				GenericLog: &pb.FlashingStatusUpdate_GenericLog{
					Log: fmt.Sprintf("%v%v", l.baseLogger.Prefix(), msg),
				},
			},
		})
	}
}

func (l *Logger) logString(msg string) {
	l.baseLogger.Println(msg)
	l.trySendToSupervisor(msg)
}

func (l *Logger) Log(v ...interface{}) {
	l.logString(fmt.Sprint(v...))
}

func (l *Logger) Logf(format string, v ...interface{}) {
	l.logString(fmt.Sprintf(format, v...))
}

func (l *Logger) AttachSupervisor(client *pb.FlashingSupervisor_SuperviseClient) {
	l.superviseClient = client
}

func (l *Logger) DetachSupervisor() {
	l.superviseClient = nil
}
