package handler

import (
	"chainmscan/blockchain"
	"chainmscan/server"

	"github.com/gin-gonic/gin"
)

type SubscribeHandler struct {
}

type SubscribeReq struct {
	ChainId          string `json:"chainId"`
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

		err := checkStringParamsEmpty(req.ChainId, req.OrgId, req.SignKeyPem,
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
		}

		client, err := blockchain.NewChainmakerClient(config)
		if err != nil {
			log.Errorf("failed to create chainmaker client, err: [%s], chainId: [%s]\n",
				err.Error(), req.ChainId)
			FailedJSONResp(RespMsgServerError, c)
			return
		}

		err = s.Subscribe(client, s.Db())
		if err != nil {
			log.Errorf("failed to subscribe, err: [%s], genHash: [%s]\n",
				err.Error(), client.GetChainGenHash())
			FailedJSONResp(RespMsgServerError, c)
			return
		}

		SuccessfulJSONResp(client.GetChainGenHash(), "", c)
	}
}
