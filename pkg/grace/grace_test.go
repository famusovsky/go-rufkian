package grace

import (
	"os"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

type mockProcess struct {
	runCalled      bool
	shutdownCalled bool
	wg             sync.WaitGroup
}

func newMockProcess() *mockProcess {
	mp := &mockProcess{}
	mp.wg.Add(1)
	return mp
}

func (m *mockProcess) Run() {
	m.runCalled = true
	m.wg.Done()
}

func (m *mockProcess) Shutdown() {
	m.shutdownCalled = true
}

func (m *mockProcess) WaitForRun() {
	m.wg.Wait()
}

func TestHandle(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	mp := newMockProcess()

	go func() {
		mp.WaitForRun()

		time.Sleep(100 * time.Millisecond)

		p, _ := os.FindProcess(os.Getpid())
		p.Signal(syscall.SIGINT)
	}()

	Handle(mp, logger)

	assert.True(t, mp.runCalled)
	assert.True(t, mp.shutdownCalled)
}

func TestHandleWithTermSignal(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	mp := newMockProcess()

	go func() {
		mp.WaitForRun()

		time.Sleep(100 * time.Millisecond)

		p, _ := os.FindProcess(os.Getpid())
		p.Signal(syscall.SIGTERM)
	}()

	Handle(mp, logger)

	assert.True(t, mp.runCalled)
	assert.True(t, mp.shutdownCalled)
}

func TestHandleWithMultipleSignals(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	mp := newMockProcess()

	go func() {
		mp.WaitForRun()

		time.Sleep(50 * time.Millisecond)

		p, _ := os.FindProcess(os.Getpid())
		p.Signal(syscall.SIGINT)

		time.Sleep(50 * time.Millisecond)

		p.Signal(syscall.SIGTERM)
	}()

	Handle(mp, logger)

	assert.True(t, mp.runCalled)
	assert.True(t, mp.shutdownCalled)
}

type autoShutdownProcess struct {
	runCalled      bool
	shutdownCalled bool
	wg             sync.WaitGroup
	sigChan        chan os.Signal
}

func newAutoShutdownProcess() *autoShutdownProcess {
	p := &autoShutdownProcess{
		sigChan: make(chan os.Signal, 1),
	}
	p.wg.Add(1)
	return p
}

func (p *autoShutdownProcess) Run() {
	p.runCalled = true
	p.wg.Done()

	go func() {
		time.Sleep(100 * time.Millisecond)
		p.sigChan <- syscall.SIGINT
	}()
}

func (p *autoShutdownProcess) Shutdown() {
	p.shutdownCalled = true
}

func (p *autoShutdownProcess) WaitForRun() {
	p.wg.Wait()
}

func TestHandleWithProcessInitiatedShutdown(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	mp := newAutoShutdownProcess()

	go func() {
		mp.WaitForRun()

		time.Sleep(50 * time.Millisecond)

		// No need to notify the channel as we're directly sending a signal
	}()

	Handle(mp, logger)

	assert.True(t, mp.runCalled)
	assert.True(t, mp.shutdownCalled)
}
