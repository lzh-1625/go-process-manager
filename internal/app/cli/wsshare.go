package cli

import (
	"os"
	"strconv"

	"github.com/lzh-1625/go_process_manager/internal/app/model"
	"github.com/olekukonko/tablewriter"
)

type WSShareCli struct{}

func NewWSShareCli() *WSShareCli {
	return &WSShareCli{}
}

func (w *WSShareCli) GetList() error {
	result, err := Get[[]*model.WsShare]("/api/ws/token/list", nil)
	if err != nil {
		return err
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.Header([]string{
		"ID",
		"PID",
		"WRITE",
		"TOKEN",
		"CREATE BY",
		"CREATED AT",
		"EXPIRE TIME",
		"LAST LINK",
	})

	for _, share := range *result {
		table.Append([]string{
			strconv.Itoa(share.ID),
			strconv.Itoa(share.Pid),
			strconv.FormatBool(share.Write),
			share.Token,
			share.CreateBy,
			share.CreatedAt.Format("2006-01-02 15:04:05"),
			share.ExpireTime.Format("2006-01-02 15:04:05"),
			share.LastLink.Format("2006-01-02 15:04:05"),
		})
	}

	table.Render()
	return nil
}

func (w *WSShareCli) Delete(id int) error {
	_, err := Delete[struct{}]("/api/ws/token", map[string]string{"id": strconv.Itoa(id)})
	if err != nil {
		return err
	}
	return nil
}
