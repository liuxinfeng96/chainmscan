package handler

import (
	"chainmscan/db/dao"
	"chainmscan/server"
	"time"

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

type TxDetailsHandler struct {
}

type TxDetailsReq struct {
	GenHash string `json:"genHash"`
	TxId    string `json:"txId"`
	Id      int    `json:"id"`
}

type TxDetailsResp struct {
	TxId                  string `json:"txId"`
	BlockHeight           uint64 `json:"blockHeight"`
	ChainId               string `json:"chainId"`
	ContractName          string `json:"contractName"`
	Method                string `json:"method"`
	TxType                string `json:"txType"`
	Timestamp             int64  `json:"timestamp"`
	ExpirationTime        int64  `json:"expirationTime"`
	GasLimit              uint64 `json:"gasLimit"`
	SenderOrgId           string `json:"senderOrgId"`
	TxStatusCode          string `json:"txStatusCode"`
	TxParameters          string `json:"txParameters"`
	RwSetHash             string `json:"rwSetHash"`
	TxMessage             string `json:"txMessage"`
	ContractResultCode    uint32 `json:"contractResultCode"`
	ContractResult        string `json:"contractResult"`
	ContractResultMessage string `json:"contractResultMessage"`
	GasUsed               uint64 `json:"gasUsed"`
}

func (h *TxDetailsHandler) Handle(s *server.Server) gin.HandlerFunc {
	return func(c *gin.Context) {

		req := new(TxDetailsReq)
		if err := c.ShouldBindJSON(req); err != nil {
			FailedJSONResp(RespMsgParamsTypeError, c)
			return
		}

		err := checkStringParamsEmpty(req.GenHash, req.TxId)
		if err != nil {
			FailedJSONResp(RespMsgParamsMissing, c)
			return
		}

		if len(req.TxId) == 0 && req.Id == 0 {
			FailedJSONResp(RespMsgParamsMissing, c)
			return
		}

		log, err := s.GetZapLogger("TxDetailsHandler")
		if err != nil {
			FailedJSONResp(RespMsgLogServerError, c)
			return
		}

		txInfo, tableNum, err := dao.GetTxInfo(req.GenHash, req.TxId, req.Id, s.Db())
		if err != nil {
			log.Errorf("fail to get tx info, err: [%s], genHash: [%s], txId: [%s]\n",
				err.Error(), req.GenHash, req.TxId)
			FailedJSONResp(RespMsgServerError, c)
			return
		}

		txDetails, err := dao.GetTxDetails(req.TxId, tableNum, s.Db())
		if err != nil {
			log.Errorf("fail to get tx details, err: [%s], genHash: [%s], txId: [%s]\n",
				err.Error(), req.GenHash, req.TxId)
			FailedJSONResp(RespMsgServerError, c)
			return
		}

		resp := &TxDetailsResp{
			TxId:                  txInfo.TxId,
			BlockHeight:           txInfo.BlockHeight,
			ChainId:               txInfo.ChainId,
			ContractName:          txInfo.ContractName,
			Method:                txInfo.Method,
			TxType:                txInfo.TxType,
			Timestamp:             txInfo.Timestamp,
			ExpirationTime:        txInfo.ExpirationTime,
			GasLimit:              txInfo.GasLimit,
			SenderOrgId:           txInfo.SenderOrgId,
			TxStatusCode:          txInfo.TxStatusCode,
			TxParameters:          string(txDetails.TxParameters),
			RwSetHash:             txDetails.RwSetHash,
			TxMessage:             txDetails.TxMessage,
			ContractResultCode:    txDetails.ContractResultCode,
			ContractResult:        string(txDetails.ContractResult),
			ContractResultMessage: txDetails.ContractResultMessage,
			GasUsed:               txDetails.GasUsed,
		}

		SuccessfulJSONResp(resp, "", c)
	}
}

type TxAmountByTimeHandler struct {
}

type TxAmountByTimeReq struct {
	GenHash string `json:"genHash"`
}

type TxAmountByTimeResp struct {
	Timestamp int64 `json:"timestamp"`
	TxAmount  int64 `json:"txAmount"`
}

func (h *TxAmountByTimeHandler) Handle(s *server.Server) gin.HandlerFunc {
	return func(c *gin.Context) {

		req := new(TxAmountByTimeReq)
		if err := c.ShouldBindJSON(req); err != nil {
			FailedJSONResp(RespMsgParamsTypeError, c)
			return
		}

		err := checkStringParamsEmpty(req.GenHash)
		if err != nil {
			FailedJSONResp(RespMsgParamsMissing, c)
			return
		}

		log, err := s.GetZapLogger("TxAmountByTimeHandler")
		if err != nil {
			FailedJSONResp(RespMsgLogServerError, c)
			return
		}

		resp := make([]*TxAmountByTimeResp, 0)

		t := time.Now().Add(24 * time.Hour * (-1))

		for i := 0; i < 24; i++ {
			startTime := t.Add(time.Hour * time.Duration(i)).Unix()
			endTime := t.Add(time.Hour * time.Duration(i+1)).Unix()

			txAmount, err := dao.GetTxAmountByTime(req.GenHash, startTime, endTime, s.Db())
			if err != nil {
				log.Errorf("fail to get tx amount by time, err: [%s], genHash: [%s]\n",
					err.Error(), req.GenHash)
				FailedJSONResp(RespMsgServerError, c)
				return
			}

			resp = append(resp, &TxAmountByTimeResp{
				Timestamp: endTime,
				TxAmount:  txAmount,
			})
		}

		SuccessfulJSONResp(resp, "", c)
	}
}
