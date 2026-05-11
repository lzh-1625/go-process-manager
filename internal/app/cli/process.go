package cli

import (
	"fmt"
	"math"
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

func (p *ProcessCli) GetList() error {
	result, err := Get[[]model.ProcessInfo]("/api/process", nil)
	if err != nil {
		return err
	}

	table := tablewriter.NewWriter(os.Stdout)

	table.Header([]string{
		"NAME",
		"UUID",
		"STATUS",
		"USER",
		"CPU",
		"MEMORY",
		"START TIME",
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

	getCPUUsageString := func(usage []float64) string {
		if len(usage) == 0 {
			return "-"
		}
		return fmt.Sprintf("%f%%", usage[len(usage)-1])
	}
	getMemoryUsageString := func(usage []float64) string {
		if len(usage) == 0 {
			return "-"
		}

		return formatBytes(int64(usage[len(usage)-1]))
	}
	for _, process := range *result {
		table.Append([]string{
			process.Name,
			strconv.Itoa(process.UUID),
			getStateString(process.State.State),
			process.User,
			getCPUUsageString(process.Usage.Cpu),
			getMemoryUsageString(process.Usage.Mem),
			process.StartTime,
		})
	}

	table.Render()
	return nil
}

func formatBytes(bytes int64) string {
	if bytes == 0 {
		return "0B"
	}

	units := []string{"B", "K", "M", "G", "T", "P"}
	level := math.Floor(math.Log(float64(bytes)) / math.Log(1024))
	if int(level) >= len(units) {
		level = float64(len(units) - 1)
	}

	size := float64(bytes) / math.Pow(1024, level)
	return fmt.Sprintf("%.1f%s", size, units[int(level)])
}

func (p *ProcessCli) Exec(uuid int) error {
	u := url.URL{
		Scheme:   "ws",
		Host:     "localhost" + config.CF.Listen,
		Path:     "/api/ws",
		RawQuery: url.Values{"uuid": {strconv.Itoa(uuid)}, "token": {GetJwt()}}.Encode(),
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

func (p *ProcessCli) Start(uuid int) error {
	_, err := Put[struct{}]("/api/process", map[string]any{"uuid": uuid})
	if err != nil {
		return err
	}
	return nil
}

func (p *ProcessCli) Stop(uuid int) error {
	_, err := Delete[struct{}]("/api/process", map[string]string{"uuid": strconv.Itoa(uuid)})
	if err != nil {
		return err
	}
	return nil
}
