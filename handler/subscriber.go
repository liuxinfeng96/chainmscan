package handler

import (
	"chainmscan/blockchain"
	"chainmscan/db/dao"
	"chainmscan/server"
	"os"
	"path"

	"github.com/gin-gonic/gin"
)

type SubscribeHandler struct {
}

type SubscribeReq struct {
	ChainId          string `json:"chainId"`
	ChainName        string `json:"chainName"`
	OrgId            string `json:"orgId"`
	NodeAddr         string `json:"nodeAddr"`
	NodeCaCertPem    string `json:"nodeCaCertPem"`
	NodeTlsHostName  string `json:"nodeTlsHostName"`
	NodeUseTls       bool   `json:"nodeUseTls"`
	SignCertPem      string `json:"signCertPem"`
	SignKeyPem       string `json:"signKeyPem"`
	TlsCertPem       string `json:"tlsCertPem"`
	TlsKeyPem        string `json:"tlsKeyPem"`
	ArchiveCenterUrl string `json:"archiveCenterUrl"`
}

func (h *SubscribeHandler) Handle(s *server.Server) gin.HandlerFunc {
	return func(c *gin.Context) {
		req := new(SubscribeReq)
		if err := c.ShouldBindJSON(req); err != nil {
			FailedJSONResp(RespMsgParamsTypeError, c)
			return
		}

		err := checkStringParamsEmpty(req.ChainId, req.ChainName, req.OrgId, req.SignKeyPem,
			req.NodeAddr)
		if err != nil {
			FailedJSONResp(RespMsgParamsMissing, c)
			return
		}

		log, err := s.GetZapLogger("SubscribeHandler")
		if err != nil {
			FailedJSONResp(RespMsgLogServerError, c)
			return
		}

		// TODO client 加缓存？
		config := &blockchain.ClientConfig{
			ChainId:       req.ChainId,
			OrgId:         req.OrgId,
			SignKeyBytes:  []byte(req.SignKeyPem),
			SignCertBytes: []byte(req.SignCertPem),
			NodeConfs: []*blockchain.NodeConnConfig{
				{
					Addr:        req.NodeAddr,
					CaCertPem:   []string{req.NodeCaCertPem},
					TlsHostName: req.NodeTlsHostName,
					UseTls:      req.NodeUseTls,
				},
			},
			TlsKeyBytes:      []byte(req.TlsKeyPem),
			TlsCertBytes:     []byte(req.TlsCertPem),
			ArchiveCenterUrl: req.ArchiveCenterUrl,
			Logger:           log,
		}

		client, err := blockchain.NewChainmakerClient(config)
		if err != nil {
			log.Errorf("fail to create chainmaker client, err: [%s], chainId: [%s]\n",
				err.Error(), req.ChainId)
			FailedJSONResp(RespMsgServerError, c)
			return
		}

		err = s.Subscribe(client, req.ChainName, s.Db())
		if err != nil {
			log.Errorf("fail to subscribe, err: [%s], genHash: [%s]\n",
				err.Error(), client.GetChainGenHash())
			FailedJSONResp(RespMsgServerError, c)
			return
		}

		SuccessfulJSONResp(client.GetChainGenHash(), "", c)
	}
}

type UnSubscribeHandler struct {
}

type UnSubscribeReq struct {
	GenHash string `json:"genHash"`
}

func (h *UnSubscribeHandler) Handle(s *server.Server) gin.HandlerFunc {
	return func(c *gin.Context) {
		req := new(UnSubscribeReq)
		if err := c.ShouldBindJSON(req); err != nil {
			FailedJSONResp(RespMsgParamsTypeError, c)
			return
		}

		err := checkStringParamsEmpty(req.GenHash)
		if err != nil {
			FailedJSONResp(RespMsgParamsMissing, c)
			return
		}

		log, err := s.GetZapLogger("UnSubscribeHandler")
		if err != nil {
			FailedJSONResp(RespMsgLogServerError, c)
			return
		}

		err = s.UnSubscribe(req.GenHash, s.Db())
		if err != nil {
			log.Errorf("fail to unsubscribe, err: [%s], genHash: [%s]\n",
				err.Error(), req.GenHash)
			FailedJSONResp(RespMsgServerError, c)
			return
		}

		SuccessfulJSONResp(req.GenHash, "", c)
	}
}

type SubscribeByFileHandler struct {
}

type SubscribeByFileReq struct {
	ChainId          string `json:"chainId"`
	ChainName        string `json:"chainName"`
	OrgId            string `json:"orgId"`
	NodeAddr         string `json:"nodeAddr"`
	NodeCaCertFileId string `json:"nodeCaCertFileId"`
	NodeTlsHostName  string `json:"nodeTlsHostName"`
	NodeUseTls       bool   `json:"nodeUseTls"`
	SignCertFileId   string `json:"signCertFileId"`
	SignKeyFileId    string `json:"signKeyFileId"`
	TlsCertFileId    string `json:"tlsCertFileId"`
	TlsKeyFileId     string `json:"tlsKeyFileId"`
	ArchiveCenterUrl string `json:"archiveCenterUrl"`
}

func (h *SubscribeByFileHandler) Handle(s *server.Server) gin.HandlerFunc {
	return func(c *gin.Context) {
		req := new(SubscribeByFileReq)
		if err := c.ShouldBindJSON(req); err != nil {
			FailedJSONResp(RespMsgParamsTypeError, c)
			return
		}

		err := checkStringParamsEmpty(req.ChainId, req.OrgId,
			req.ChainName, req.SignKeyFileId, req.NodeAddr)
		if err != nil {
			FailedJSONResp(RespMsgParamsMissing, c)
			return
		}

		log, err := s.GetZapLogger("SubscribeHandler")
		if err != nil {
			FailedJSONResp(RespMsgLogServerError, c)
			return
		}

		signKeyPem, err := readFileBytes(req.SignKeyFileId, s.UploadFilePath())
		if err != nil {
			log.Errorf("fail to read file, err: [%s], fileId: [%s]\n",
				err.Error(), req.SignKeyFileId)
			FailedJSONResp(RespMsgServerError, c)
			return
		}
		SignCertPem, err := readFileBytes(req.SignCertFileId, s.UploadFilePath())
		if err != nil {
			log.Errorf("fail to read file, err: [%s], fileId: [%s]\n",
				err.Error(), req.SignCertFileId)
			FailedJSONResp(RespMsgServerError, c)
			return
		}
		tlsKeyPem, err := readFileBytes(req.TlsKeyFileId, s.UploadFilePath())
		if err != nil {
			log.Errorf("fail to read file, err: [%s], fileId: [%s]\n",
				err.Error(), req.TlsKeyFileId)
			FailedJSONResp(RespMsgServerError, c)
			return
		}
		tlsCertPem, err := readFileBytes(req.TlsCertFileId, s.UploadFilePath())
		if err != nil {
			log.Errorf("fail to read file, err: [%s], fileId: [%s]\n",
				err.Error(), req.TlsCertFileId)
			FailedJSONResp(RespMsgServerError, c)
			return
		}
		nodeCaCertPem, err := readFileBytes(req.NodeCaCertFileId, s.UploadFilePath())
		if err != nil {
			log.Errorf("fail to read file, err: [%s], fileId: [%s]\n",
				err.Error(), req.NodeCaCertFileId)
			FailedJSONResp(RespMsgServerError, c)
			return
		}

		// TODO client 加缓存？
		config := &blockchain.ClientConfig{
			ChainId:       req.ChainId,
			OrgId:         req.OrgId,
			SignKeyBytes:  signKeyPem,
			SignCertBytes: SignCertPem,
			NodeConfs: []*blockchain.NodeConnConfig{
				{
					Addr:        req.NodeAddr,
					CaCertPem:   []string{string(nodeCaCertPem)},
					TlsHostName: req.NodeTlsHostName,
					UseTls:      req.NodeUseTls,
				},
			},
			TlsKeyBytes:      tlsKeyPem,
			TlsCertBytes:     tlsCertPem,
			ArchiveCenterUrl: req.ArchiveCenterUrl,
			Logger:           log,
		}

		client, err := blockchain.NewChainmakerClient(config)
		if err != nil {
			log.Errorf("fail to create chainmaker client, err: [%s], chainId: [%s]\n",
				err.Error(), req.ChainId)
			FailedJSONResp(RespMsgServerError, c)
			return
		}

		err = s.Subscribe(client, req.ChainName, s.Db())
		if err != nil {
			log.Errorf("fail to subscribe, err: [%s], genHash: [%s]\n",
				err.Error(), client.GetChainGenHash())
			FailedJSONResp(RespMsgServerError, c)
			return
		}

		SuccessfulJSONResp(client.GetChainGenHash(), "", c)
	}
}

type SubscriptionListHandler struct {
}

type SubscriptionListResp struct {
	GenHash   string `json:"genHash"`
	ChainName string `json:"chainName"`
	ChainId   string `json:"chainId"`
}

func (h *SubscriptionListHandler) Handle(s *server.Server) gin.HandlerFunc {
	return func(c *gin.Context) {

		log, err := s.GetZapLogger("SubscriptionListHandler")
		if err != nil {
			FailedJSONResp(RespMsgLogServerError, c)
			return
		}

		list, err := dao.GetAllSubscription(s.Db())
		if err != nil {
			log.Errorf("fail to get all subscription, err: [%s]\n", err.Error())
			FailedJSONResp(RespMsgServerError, c)
			return
		}

		resp := make([]*SubscriptionListResp, 0)

		for _, v := range list {
			resp = append(resp, &SubscriptionListResp{
				GenHash:   v.GenHash,
				ChainId:   v.ChainId,
				ChainName: v.ChainName,
			})
		}

		SuccessfulJSONResp(resp, "", c)
	}
}

func readFileBytes(filedId, dirPath string) ([]byte, error) {

	filePath := path.Join(dirPath, filedId)

	defer os.RemoveAll(filePath)

	return os.ReadFile(filePath)
}
