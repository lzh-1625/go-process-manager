package model

type FileStruct struct {
	Name  string `json:"name"`
	IsDir bool   `json:"isDir"`
}

type FilePathHandlerReq struct {
	Path string `form:"path"`
}

type FileReadHandlerReq struct {
	FilePath string `form:"filePath"`
}
