package shutdown

import (
	"context"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/stsg/gophkeeper2/pkg/logger"
)

var (
	mu = &sync.Mutex{}
)

//go:generate mockgen -source=exit_handler.go -destination=../mocks/shutdown/exit_handler.go -package=shutdown

type ExitHandler interface {
	IsNewFuncExecutionAllowed() bool
	ShutdownHTTPServerBeforeExit(httpServer *http.Server)
	ShutdownGrpcServerBeforeExit(grpcServer *grpc.Server)
	AddFuncInProcessing(alias string)
	FuncFinished(alias string)
	ProperExitDefer() chan struct{}

	ToCancel([]context.CancelFunc)
	ToStop([]chan struct{})
	ToClose([]io.Closer)
	ToExecute([]func(ctx context.Context) error)
}

type exitHandler struct {
	mainCtxCanceler   context.CancelFunc
	log               *zap.SugaredLogger
	httpServer        *http.Server
	grpcServer        *grpc.Server
	toCancel          []context.CancelFunc
	toStop            []chan struct{}
	toClose           []io.Closer
	toExecute         []func(ctx context.Context) error
	funcsInProcessing sync.WaitGroup
	newFuncAllowed    bool
}

func NewExitHandlerWithCtx(mainCtxCanceler context.CancelFunc) ExitHandler {
	return &exitHandler{
		log:               logger.NewLogger("exit-hdr"),
		mainCtxCanceler:   mainCtxCanceler,
		newFuncAllowed:    true,
		funcsInProcessing: sync.WaitGroup{},
	}
}

func (eh *exitHandler) ToCancel(toCancel []context.CancelFunc) {
	eh.toCancel = toCancel
}

func (eh *exitHandler) ToStop(toStop []chan struct{}) {
	eh.toStop = toStop
}

func (eh *exitHandler) ToClose(toClose []io.Closer) {
	eh.toClose = toClose
}

func (eh *exitHandler) ToExecute(toExecute []func(ctx context.Context) error) {
	eh.toExecute = toExecute
}

func (eh *exitHandler) IsNewFuncExecutionAllowed() bool {
	mu.Lock()
	defer mu.Unlock()
	return eh.newFuncAllowed
}

func (eh *exitHandler) setNewFuncExecutionAllowed(value bool) {
	mu.Lock()
	defer mu.Unlock()
	eh.newFuncAllowed = value
}

func (eh *exitHandler) ShutdownHTTPServerBeforeExit(httpServer *http.Server) {
	eh.httpServer = httpServer
}

func (eh *exitHandler) ShutdownGrpcServerBeforeExit(grpcServer *grpc.Server) {
	eh.grpcServer = grpcServer
}

func (eh *exitHandler) AddFuncInProcessing(alias string) {
	mu.Lock()
	defer mu.Unlock()
	eh.log.Infof("'%s' func is started and added to exit handler", alias)
	eh.funcsInProcessing.Add(1)
}

func (eh *exitHandler) FuncFinished(alias string) {
	mu.Lock()
	defer mu.Unlock()
	eh.log.Infof("'%s' func is finished and removed from exit handler", alias)
	eh.funcsInProcessing.Add(-1)
}

func (eh *exitHandler) ProperExitDefer() chan struct{} {
	eh.log.Info("Graceful exit handler is activated")
	signals := make(chan os.Signal, 1)
	signal.Notify(signals,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)
	exit := make(chan struct{})
	go func() {
		s := <-signals
		eh.log.Infof("Received a signal '%s'", s)
		exit <- struct{}{}
		eh.setNewFuncExecutionAllowed(false)
		eh.Shutdown()
	}()
	return exit
}

func (eh *exitHandler) Shutdown() {
	successfullyFinished := make(chan struct{})
	go func() {
		eh.waitForShutdownServer()
		eh.waitForFinishFunc()
		eh.endHeldObjects()
		successfullyFinished <- struct{}{}
	}()
	select {
	case <-successfullyFinished:
		log.Println("System finished work, graceful shutdown")
		eh.mainCtxCanceler()
	case <-time.After(1 * time.Minute):
		log.Println("System has not shutdown in time '1m', shutdown with interruption")
		os.Exit(1)
	}
}

func (eh *exitHandler) waitForFinishFunc() {
	log.Println("Waiting for functions finish work...")
	eh.funcsInProcessing.Wait()
	log.Println("All functions finished work successfully")
}

func (eh *exitHandler) waitForShutdownServer() {
	if eh.httpServer != nil {
		log.Println("Waiting for shutdown http server...")
		err := eh.httpServer.Shutdown(context.Background())
		log.Println("Http Server shutdown complete")
		if err != nil {
			eh.log.Infof("failed to shutdown server: %v", err)
		}
	}
	if eh.grpcServer != nil {
		log.Println("Waiting for shutdown proto server...")
		eh.grpcServer.GracefulStop()
		log.Println("Grpc Server shutdown complete")
	}
}

func (eh *exitHandler) endHeldObjects() {
	if len(eh.toExecute) > 0 {
		log.Println("ToExecute final funcs")
		for _, execute := range eh.toExecute {
			err := execute(context.Background())
			if err != nil {
				eh.log.Infof("func error: %v", err)
			}
		}
	}
	if len(eh.toCancel) > 0 {
		log.Println("ToCancel active contexts")
		for _, cancel := range eh.toCancel {
			cancel()
		}
	}
	if len(eh.toStop) > 0 {
		log.Println("ToStop active goroutines")
		for _, toStop := range eh.toStop {
			close(toStop)
		}
	}
	if len(eh.toClose) > 0 {
		log.Println("ToClose active resources")
		for _, toClose := range eh.toClose {
			err := toClose.Close()
			if err != nil {
				eh.log.Infof("failed to close an resource: %v", err)
			}
		}
	}
	log.Println("Success end final work")
}
