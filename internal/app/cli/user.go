package cli

import (
	"os"

	"github.com/lzh-1625/go_process_manager/internal/app/model"
	"github.com/lzh-1625/go_process_manager/internal/app/types"
	"github.com/olekukonko/tablewriter"
)

type UserCli struct{}

func NewUserCli() *UserCli {
	return &UserCli{}
}

func getRoleString(role types.Role) string {
	switch role {
	case types.RoleRoot:
		return "Root"
	case types.RoleAdmin:
		return "Admin"
	case types.RoleUser:
		return "User"
	case types.RoleGuest:
		return "Guest"
	default:
		return "Unknown"
	}
}

func (u *UserCli) GetList() {
	result, err := Get[[]*model.User]("/api/user", nil)

	checkError(err)

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
}

func (u *UserCli) Delete(account string) {
	_, err := Delete[struct{}]("/api/user", map[string]string{"account": account})
	checkError(err)
}
