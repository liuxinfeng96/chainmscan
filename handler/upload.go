package handler

import (
	"chainmscan/server"
	"os"
	"path"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UploadFileHandler struct {
}

type UploadFileResp struct {
	FileId string `json:"fileId"`
}

func (h *UploadFileHandler) Handle(s *server.Server) gin.HandlerFunc {
	return func(c *gin.Context) {

		file, err := c.FormFile("file")
		if err != nil {
			FailedJSONResp(RespMsgParamsTypeError, c)
			return
		}

		log, err := s.GetZapLogger("UploadFileHandler")
		if err != nil {
			FailedJSONResp(RespMsgLogServerError, c)
			return
		}

		fileId := uuid.New()

		filePath := path.Join(s.UploadFilePath(), fileId.String())

		err = c.SaveUploadedFile(file, filePath)
		if err != nil {
			log.Errorf("fail to save the file, err: [%s]\n", err.Error())
			FailedJSONResp(RespMsgServerError, c)
			os.RemoveAll(filePath)
			return
		}

		resp := new(UploadFileResp)
		resp.FileId = fileId.String()

		SuccessfulJSONResp(resp, "", c)
	}
}
