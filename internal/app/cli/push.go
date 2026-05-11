package cli

import (
	"os"
	"strconv"

	"github.com/lzh-1625/go_process_manager/internal/app/model"
	"github.com/olekukonko/tablewriter"
)

type PushCli struct{}

func NewPushCli() *PushCli {
	return &PushCli{}
}

func (p *PushCli) GetList() error {
	result, err := Get[[]*model.Push]("/api/push/list", nil)
	if err != nil {
		return err
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.Header([]string{
		"ID",
		"METHOD",
		"URL",
		"REMARK",
		"ENABLE",
	})

	for _, push := range *result {
		table.Append([]string{
			strconv.FormatInt(push.ID, 10),
			push.Method,
			push.Url,
			push.Remark,
			strconv.FormatBool(push.Enable),
		})
	}

	table.Render()
	return nil
}

func (p *PushCli) GetByID(id int) error {
	result, err := Get[model.Push]("/api/push", map[string]string{"id": strconv.Itoa(id)})
	if err != nil {
		return err
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.Header([]string{"FIELD", "VALUE"})
	table.Append([]string{"ID", strconv.FormatInt(result.ID, 10)})
	table.Append([]string{"Method", result.Method})
	table.Append([]string{"URL", result.Url})
	table.Append([]string{"Body", result.Body})
	table.Append([]string{"Remark", result.Remark})
	table.Append([]string{"Enable", strconv.FormatBool(result.Enable)})

	table.Render()
	return nil
}

func (p *PushCli) Create(push model.Push) error {
	_, err := Post[struct{}]("/api/push", push)
	if err != nil {
		return err
	}
	return nil
}

func (p *PushCli) Update(push model.Push) error {
	_, err := Put[struct{}]("/api/push", push)
	if err != nil {
		return err
	}
	return nil
}

func (p *PushCli) Delete(id int) error {
	_, err := Delete[struct{}]("/api/push", map[string]string{"id": strconv.Itoa(id)})
	if err != nil {
		return err
	}
	return nil
}
