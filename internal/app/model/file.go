package model

type FileStruct struct {
	Name  string `json:"name"`
	IsDir bool   `json:"isDir"`
}

type FilePathHandlerReq struct {
	Path string `query:"path" binding:"required"`
}

type FileReadHandlerReq struct {
	FilePath string `query:"filePath" binding:"required"`
}
