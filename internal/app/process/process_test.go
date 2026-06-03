package process

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/lzh-1625/go_process_manager/config"
	"github.com/lzh-1625/go_process_manager/internal/app/eum"
	"github.com/lzh-1625/go_process_manager/internal/app/model"
)

func TestProcess(t *testing.T) {
	var log string
	process := NewProcessPty(model.Process{
		Name:      "test",
		Cmd:       "sh -c 'sleep 1 && echo test'",
		Cwd:       ".",
		LogReport: true,
	},
		SetLogHandler(false, func(p *ProcessBase, content []byte) {
			log = string(content)
		}),
	)
	process.Start()
	<-process.StopChan
	if strings.TrimSpace(log) != "test" {
		t.Errorf("log is not test, got %s", log)
	}
}

func TestPipeLogHandler(t *testing.T) {
	var log string
	process := NewProcessPty(model.Process{
		Name:      "test",
		Cmd:       "sh",
		Cwd:       ".",
		LogReport: true,
	},
		SetLogHandler(true, func(p *ProcessBase, content []byte) {
			log = string(content)
		}),
	)
	process.Start()
	testString := "abcdefghijklmnopqrstuvwxyz"
	for _, r := range []byte(testString) {
		time.Sleep(time.Millisecond * 50)
		process.WriteBytes([]byte{r})
	}
	process.Kill()
	if !strings.HasSuffix(log, testString) {
		t.Errorf("log is not test, got %s", log)
	}
}

type testWriter struct {
	buf   *bytes.Buffer
	close bool
}

func (t *testWriter) Write(p []byte) (n int, err error) {
	return t.buf.Write(p)
}
func (t *testWriter) Close() error {
	if t.close {
		return errors.New("close already")
	}
	t.close = true
	return nil
}

func TestProcessWriter(t *testing.T) {
	config.CF.KillWaitTime = 1
	process := NewProcessPty(model.Process{
		Name: "test",
		Cmd:  "sh",
		Cwd:  ".",
	})
	process.Start()
	tw := &testWriter{
		buf: bytes.NewBuffer(make([]byte, 1024)),
	}
	process.AddWriter("test", tw)

	textString := "abcdefghijklmnopqrstuvwxyz"
	for _, v := range []byte(textString) {
		time.Sleep(time.Millisecond * 50)
		process.WriteBytes([]byte{v})
	}
	process.ProcessControl("user")
	process.Kill()
	if str := tw.buf.String(); !strings.Contains(str, textString) || !tw.close {
		t.Errorf("log is not test, got %s, close: %t", str, tw.close)
	}
}

func TestProcessAtomic(t *testing.T) {
	config.CF.KillWaitTime = 1
	process := NewProcessPty(model.Process{
		Name: "test",
		Cmd:  "sleep 100",
		Cwd:  ".",
	},
		SetStateHook(func(p *ProcessBase, state eum.ProcessState) {
			t.Logf("state: %v", state)
		}),
	)

	wg := sync.WaitGroup{}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	wg.Go(func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				err := process.Start()
				if err != nil {
					t.Logf("process start failed, err: %v", err)
				} else {
					t.Log("process start success")
				}
				time.Sleep(time.Millisecond * 10)
			}
		}
	})
	wg.Go(func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				err := process.Kill()
				if err != nil {
					t.Logf("process kill failed, err: %v", err)
				} else {
					t.Log("process kill success")
				}
				time.Sleep(time.Millisecond * 10)
			}
		}
	})
	wg.Wait()
}
