//go:build slim

package bleve

import "github.com/lzh-1625/go_process_manager/internal/app/repository/search"

func NewBleveSearch() search.ILogLogic {
	panic("bleve not support in slim mode")
}
