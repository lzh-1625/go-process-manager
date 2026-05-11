package cli

import (
	"os"

	"github.com/lzh-1625/go_process_manager/internal/app/eum"
	"github.com/lzh-1625/go_process_manager/internal/app/model"
	"github.com/olekukonko/tablewriter"
)

type UserCli struct{}

func NewUserCli() *UserCli {
	return &UserCli{}
}

func getRoleString(role eum.Role) string {
	switch role {
	case eum.RoleRoot:
		return "Root"
	case eum.RoleAdmin:
		return "Admin"
	case eum.RoleUser:
		return "User"
	case eum.RoleGuest:
		return "Guest"
	default:
		return "Unknown"
	}
}

func (u *UserCli) GetList() error {
	result, err := Get[[]*model.User]("/api/user", nil)
	if err != nil {
		return err
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.Header([]string{
		"ACCOUNT",
		"ROLE",
		"CREATE TIME",
		"REMARK",
	})

	for _, user := range *result {
		table.Append([]string{
			user.Account,
			getRoleString(user.Role),
			user.CreateTime.Format("2006-01-02 15:04:05"),
			user.Remark,
		})
	}

	table.Render()
	return nil
}

func (u *UserCli) Delete(account string) error {
	_, err := Delete[struct{}]("/api/user", map[string]string{"account": account})
	if err != nil {
		return err
	}
	return nil
}
