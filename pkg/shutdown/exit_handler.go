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

// ExitHandler interface
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

// NewExitHandlerWithCtx creates a new instance of the ExitHandler interface with the given main context canceler.
//
// Parameters:
// - mainCtxCanceler: A function that cancels the main context.
//
// Returns:
// - ExitHandler: An instance of the ExitHandler interface.
func NewExitHandlerWithCtx(mainCtxCanceler context.CancelFunc) ExitHandler {
	return &exitHandler{
		log:               logger.NewLogger("exit-hdr"),
		mainCtxCanceler:   mainCtxCanceler,
		newFuncAllowed:    true,
		funcsInProcessing: sync.WaitGroup{},
	}
}

// ToCancel sets the toCancel field of the exitHandler struct with the provided slice of context.CancelFunc.
//
// Parameters:
// - toCancel: A slice of context.CancelFunc to assign to the toCancel field.
func (eh *exitHandler) ToCancel(toCancel []context.CancelFunc) {
	eh.toCancel = toCancel
}

// ToStop sets the list of channels to stop active goroutines.
//
// Parameters:
// - toStop: A slice of channels to stop active goroutines.
//
// Returns:
// - None.
func (eh *exitHandler) ToStop(toStop []chan struct{}) {
	eh.toStop = toStop
}

// ToClose assigns a slice of io.Closer to the exitHandler's toClose field.
//
// Parameters:
// - toClose: A slice of io.Closer to be assigned.
func (eh *exitHandler) ToClose(toClose []io.Closer) {
	eh.toClose = toClose
}

// ToExecute assigns a slice of functions to the exitHandler's toExecute field.
//
// Parameters:
// - toExecute: A slice of functions with context that return an error to be assigned.
func (eh *exitHandler) ToExecute(toExecute []func(ctx context.Context) error) {
	eh.toExecute = toExecute
}

// IsNewFuncExecutionAllowed returns whether new function execution is allowed.
//
// It acquires a lock, checks the value of the newFuncAllowed field, and releases the lock.
//
// Returns:
// - bool: Whether new function execution is allowed.
func (eh *exitHandler) IsNewFuncExecutionAllowed() bool {
	mu.Lock()
	defer mu.Unlock()
	return eh.newFuncAllowed
}

// setNewFuncExecutionAllowed sets the value of the newFuncAllowed field in the exitHandler struct.
//
// Parameters:
// - value: A boolean value indicating whether new function execution is allowed.
//
// Returns:
// None.
func (eh *exitHandler) setNewFuncExecutionAllowed(value bool) {
	mu.Lock()
	defer mu.Unlock()
	eh.newFuncAllowed = value
}

// ShutdownHTTPServerBeforeExit sets the provided http.Server to the exitHandler's httpServer field.
//
// Parameters:
// - httpServer: A pointer to an http.Server object to be set as the exitHandler's httpServer field.
func (eh *exitHandler) ShutdownHTTPServerBeforeExit(httpServer *http.Server) {
	eh.httpServer = httpServer
}

// ShutdownGrpcServerBeforeExit sets the provided grpc.Server to the exitHandler's grpcServer field.
//
// Parameters:
// - grpcServer: A pointer to a grpc.Server object.
//
// Returns:
// None.
func (eh *exitHandler) ShutdownGrpcServerBeforeExit(grpcServer *grpc.Server) {
	eh.grpcServer = grpcServer
}

// AddFuncInProcessing adds a function to the exit handler's internal processing.
//
// It acquires a lock, logs the start of the function, and increments the
// number of functions in processing.
//
// Parameters:
// - alias: A string representing the alias of the function being added.
//
// Returns: None.
func (eh *exitHandler) AddFuncInProcessing(alias string) {
	mu.Lock()
	defer mu.Unlock()
	eh.log.Infof("'%s' func is started and added to exit handler", alias)
	eh.funcsInProcessing.Add(1)
}

// FuncFinished removes the provided function alias from the exit handler's internal processing.
//
// Parameters:
// - alias: A string representing the alias of the function being finished.
//
// Returns:
// None.
func (eh *exitHandler) FuncFinished(alias string) {
	mu.Lock()
	defer mu.Unlock()
	eh.log.Infof("'%s' func is finished and removed from exit handler", alias)
	eh.funcsInProcessing.Add(-1)
}

// ProperExitDefer activates the graceful exit handler and returns a channel that
// will be closed when a proper exit signal is received. The exit handler
// listens for SIGINT, SIGTERM, and SIGQUIT signals and performs the following
// actions:
// - Logs the activation of the graceful exit handler.
// - Waits for a signal to be received on the signals channel.
// - Logs the received signal.
// - Closes the exit channel.
// - Sets the new function execution flag to false.
// - Calls the Shutdown method of the exit handler.
//
// Returns:
// - A channel of type struct{} that will be closed when a proper exit signal is received.
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

// Shutdown gracefully shuts down the system by waiting for the shutdown server,
// finishing all functions, and ending held objects. It returns when the system
// has finished work and gracefully shut down, or after 1 minute if the system
// has not shut down in time.
//
// No parameters.
// No return values.
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

// waitForFinishFunc waits for all functions to finish work before logging that
// they have finished successfully.
//
// No parameters.
// No return values.
func (eh *exitHandler) waitForFinishFunc() {
	log.Println("Waiting for functions finish work...")
	eh.funcsInProcessing.Wait()
	log.Println("All functions finished work successfully")
}

// waitForShutdownServer waits for the shutdown HTTP server and the shutdown
// gRPC server to complete their tasks and gracefully stop. It logs the progress
// of the shutdown process and logs an error if any occurs.
//
// No parameters.
// No return values.
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

// endHeldObjects performs cleanup by executing final functions, canceling active contexts,
// stopping active goroutines, closing active resources, and logging the successful completion of final work.
//
// No parameters.
// No return values.
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
