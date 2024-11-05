package handler

import (
	"chainmscan/db/dao"
	"chainmscan/server"

	"github.com/gin-gonic/gin"
)

type TxListHandler struct {
}

const TxListByContractNameCount = 100

type TxListReq struct {
	PageReq
	GenHash      string `json:"genHash"`
	BlockHeight  int64  `json:"blockHeight"`
	ContractName string `json:"contractName"`
}

type TxListResp struct {
	TxId         string `json:"txId"`
	BlockHeight  uint64 `json:"blockHeight"`
	ChainId      string `json:"chainId"`
	ContractName string `json:"contractName"`
	Method       string `json:"method"`
	TxType       string `json:"txType"`
	Timestamp    int64  `json:"timestamp"`
	SenderOrgId  string `json:"senderOrgId"`
	TxStatusCode string `json:"txStatusCode"`
}

func (h *TxListHandler) Handle(s *server.Server) gin.HandlerFunc {
	return func(c *gin.Context) {

		req := new(TxListReq)
		if err := c.ShouldBindJSON(req); err != nil {
			FailedJSONResp(RespMsgParamsTypeError, c)
			return
		}

		err := checkStringParamsEmpty(req.GenHash)
		if err != nil {
			FailedJSONResp(RespMsgParamsMissing, c)
			return
		}

		checkPageReq(&req.PageReq)

		log, err := s.GetZapLogger("TxListHandler")
		if err != nil {
			FailedJSONResp(RespMsgLogServerError, c)
			return
		}

		if len(req.ContractName) != 0 {
			// 通过合约查询交易，考虑性能，只提供最新100条交易列表
			txList, err := dao.GetLatestTxListByContractName(req.GenHash, req.ContractName, TxListByContractNameCount, s.Db())
			if err != nil {
				log.Errorf("fail to get tx by contract, err: [%s], genHash: [%s], contractName: [%s]\n",
					err.Error(), req.GenHash, req.ContractName)
				FailedJSONResp(RespMsgServerError, c)
				return
			}

			resp := make([]*TxListResp, 0)

			for _, v := range txList {
				tlreq := &TxListResp{
					TxId:         v.TxId,
					BlockHeight:  v.BlockHeight,
					ChainId:      v.ChainId,
					ContractName: v.ContractName,
					Method:       v.Method,
					TxType:       v.TxType,
					Timestamp:    v.Timestamp,
					SenderOrgId:  v.SenderOrgId,
					TxStatusCode: v.TxStatusCode,
				}

				resp = append(resp, tlreq)
			}

			SuccessfulJSONRespWithPage(resp, TxListByContractNameCount, c)

		} else {
			txList, err := dao.GetTxList(req.GenHash, req.Page, req.PageSize, req.BlockHeight, s.Db())
			if err != nil {
				log.Errorf("fail to get tx list, err: [%s], genHash: [%s], height: [%d]\n",
					err.Error(), req.GenHash, req.BlockHeight)
				FailedJSONResp(RespMsgServerError, c)
				return
			}

			resp := make([]*TxListResp, 0)

			for _, v := range txList {
				tlreq := &TxListResp{
					TxId:         v.TxId,
					BlockHeight:  v.BlockHeight,
					ChainId:      v.ChainId,
					ContractName: v.ContractName,
					Method:       v.Method,
					TxType:       v.TxType,
					Timestamp:    v.Timestamp,
					SenderOrgId:  v.SenderOrgId,
					TxStatusCode: v.TxStatusCode,
				}

				resp = append(resp, tlreq)
			}

			var txCount int64

			if req.BlockHeight > 0 {
				// 考虑性能，从区块中获取交易数量
				txCount, err = dao.GetBlockTxCount(req.GenHash, req.BlockHeight, s.Db())
				if err != nil {
					log.Errorf("fail to get tx count, err: [%s], genHash: [%s], height: [%d]\n",
						err.Error(), req.GenHash, req.BlockHeight)
					FailedJSONResp(RespMsgServerError, c)
					return
				}
			} else {
				txCount, err = dao.GetChainTxAmount(req.GenHash, s.Db())
				if err != nil {
					log.Errorf("fail to get tx count, err: [%s], genHash: [%s]\n",
						err.Error(), req.GenHash)
					FailedJSONResp(RespMsgServerError, c)
					return
				}
			}

			SuccessfulJSONRespWithPage(resp, txCount, c)
		}

	}
}
