package cli

import (
	"os"
	"strconv"

	"github.com/lzh-1625/go_process_manager/internal/app/model"
	"github.com/olekukonko/tablewriter"
)

type TaskCli struct{}

func NewTaskCli() *TaskCli {
	return &TaskCli{}
}

func (t *TaskCli) GetList() error {
	result, err := Get[[]model.TaskVo]("/api/task/all", nil)
	if err != nil {
		return err
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.Header([]string{
		"ID",
		"NAME",
		"CRON",
		"START TIME",
		"RUNNING",
	})

	for _, task := range *result {
		table.Append([]string{
			strconv.Itoa(task.ID),
			task.Name,
			task.CronExpression,
			task.StartTime,
			strconv.FormatBool(task.Running),
		})
	}

	table.Render()
	return nil
}

func (t *TaskCli) Delete(id int) error {
	_, err := Delete[struct{}]("/api/task", map[string]string{"id": strconv.Itoa(id)})
	if err != nil {
		return err
	}
	return nil
}

func (t *TaskCli) Start(id int) error {
	_, err := Get[struct{}]("/api/task/start", map[string]string{"id": strconv.Itoa(id)})
	if err != nil {
		return err
	}
	return nil
}

func (t *TaskCli) Stop(id int) error {
	_, err := Get[struct{}]("/api/task/stop", map[string]string{"id": strconv.Itoa(id)})
	if err != nil {
		return err
	}
	return nil
}
