package superlog

import (
	"context"
	"fmt"
	"log"
	"time"

	pb "git.dolansoft.org/philippe/softmetal/pb"
)

var timeout = time.Millisecond * 500

type Logger struct {
	baseLogger      *log.Logger
	superviseClient pb.FlashingSupervisorClient
	sessID          uint64
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
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		c.RecordLog(ctx, &pb.RecordLogRequest{
			SessionId: l.sessID,
			Log:       fmt.Sprintf("%v%v", l.baseLogger.Prefix(), msg),
		})
	}
}

func (l *Logger) logString(msg string) {
	l.baseLogger.Println(msg)
	l.trySendToSupervisor(msg)
}

func (l *Logger) Logf(format string, v ...interface{}) {
	l.logString(fmt.Sprintf(format, v...))
}

func (l *Logger) Progress(p float32) {
	l.baseLogger.Printf("Progress: %v", p)
	c := l.superviseClient
	if c != nil {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		c.RecordProgress(ctx, &pb.RecordProgressRequest{
			SessionId: l.sessID,
			Progress:  p,
		})
	}
}

func (l *Logger) AttachSupervisor(client pb.FlashingSupervisorClient, sessID uint64) {
	l.superviseClient = client
	l.sessID = sessID
}

func (l *Logger) DetachSupervisor() {
	l.superviseClient = nil
	l.sessID = 0
}
