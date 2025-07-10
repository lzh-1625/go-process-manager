package api

import (
	"github.com/lzh-1625/go_process_manager/internal/app/logic"
	"github.com/lzh-1625/go_process_manager/internal/app/model"

	"github.com/gin-gonic/gin"
)

type file struct{}

var FileApi = new(file)

func (f *file) FilePathHandler(ctx *gin.Context, req model.FilePathHandlerReq) []model.FileStruct {
	return logic.FileLogic.GetFileAndDirByPath(req.Path)
}

func (f *file) FileWriteHandler(ctx *gin.Context, _ any) (err error) {
	path := ctx.PostForm("filePath")
	fi, err := ctx.FormFile("data")
	if err != nil {
		return err
	}
	fiReader, _ := fi.Open()
	err = logic.FileLogic.UpdateFileData(path, fiReader, fi.Size)
	return
}

func (f *file) FileReadHandler(ctx *gin.Context, req model.FileReadHandlerReq) any {
	bytes, err := logic.FileLogic.ReadFileFromPath(req.FilePath)
	if err != nil {
		return err
	}
	return string(bytes)
}
