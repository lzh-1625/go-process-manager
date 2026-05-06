package api

import (
	"github.com/labstack/echo"
	"github.com/lzh-1625/go_process_manager/internal/app/logic"
	"github.com/lzh-1625/go_process_manager/internal/app/model"
)

type file struct{}

var FileApi = new(file)

func (f *file) FilePathHandler(ctx echo.Context) error {
	var req model.FilePathHandlerReq
	if err := ctx.Bind(&req); err != nil {
		return err
	}
	return ctx.JSON(200, logic.FileLogic.GetFileAndDirByPath(req.Path))
}

func (f *file) FileWriteHandler(ctx echo.Context) error {
	path := ctx.FormValue("filePath")
	fi, err := ctx.FormFile("data")
	if err != nil {
		return err
	}
	fiReader, _ := fi.Open()
	return logic.FileLogic.UpdateFileData(path, fiReader, fi.Size)
}

func (f *file) FileReadHandler(ctx echo.Context) error {
	var req model.FileReadHandlerReq
	if err := ctx.Bind(&req); err != nil {
		return err
	}
	bytes, err := logic.FileLogic.ReadFileFromPath(req.FilePath)
	if err != nil {
		return err
	}
	return ctx.JSON(200, string(bytes))
}
