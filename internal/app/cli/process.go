package cli

import (
	"net/url"
	"os"
	"strconv"

	"github.com/gorilla/websocket"
	"github.com/lzh-1625/go_process_manager/config"
	"github.com/lzh-1625/go_process_manager/internal/app/eum"
	"github.com/lzh-1625/go_process_manager/internal/app/model"
	"github.com/olekukonko/tablewriter"
)

type ProcessCli struct {
	uri string
}

func NewProcessCli() *ProcessCli {
	return &ProcessCli{}
}

func (p *ProcessCli) GetProcessList() error {
	result, err := Get[[]model.ProcessInfo]("/api/process")
	if err != nil {
		return err
	}

	table := tablewriter.NewWriter(os.Stdout)

	table.Header([]string{
		"NAME",
		"UUID",
		"STATUS",
		"USER",
		"CPU(%)",
		"MEMORY(KB)",
	})

	getStateString := func(state eum.ProcessState) string {
		switch state {
		case eum.ProcessStateStart:
			return "Starting"
		case eum.ProcessStateRunning:
			return "Running"
		case eum.ProcessStateStop:
			return "Stopped"
		case eum.ProcessStateWarnning:
			return "Warning"
		default:
			return "Unknown"
		}
	}

	getUsageString := func(usage []float64) string {
		if len(usage) == 0 {
			return "-"
		}
		return strconv.FormatFloat(usage[len(usage)-1], 'f', 2, 64)
	}
	for _, process := range *result {
		table.Append([]string{
			process.Name,
			strconv.Itoa(process.UUID),
			getStateString(process.State.State),
			process.User,
			getUsageString(process.Usage.Cpu),
			getUsageString(process.Usage.Mem),
		})
	}

	table.Render()
	return nil
}

func (p *ProcessCli) ProcessExec(uuid int) error {
	u := url.URL{
		Scheme:   "ws",
		Host:     "localhost" + config.CF.Listen,
		Path:     "/api/ws",
		RawQuery: url.Values{"uuid": {"1"}, "token": {GetJwt()}}.Encode(),
	}
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return err
	}

	defer conn.Close()
	go func() {
		buf := make([]byte, 1024)
		for {
			n, err := os.Stdin.Read(buf)
			if err != nil {
				return
			}
			conn.WriteMessage(websocket.TextMessage, buf[:n])
		}
	}()

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			return err
		}
		os.Stdout.Write(message)
	}
}
