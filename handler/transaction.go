package handler

import (
	"chainmscan/db/dao"
	"chainmscan/server"
	"time"

	"chainmaker.org/chainmaker/pb-go/v2/common"
	"github.com/gin-gonic/gin"
	"github.com/gogo/protobuf/proto"
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
	Id           uint   `json:"id"`
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
			txList, err := dao.GetLatestTxListByContractName(req.GenHash, req.ContractName, req.Page, req.PageSize, s.Db())
			if err != nil {
				log.Errorf("fail to get tx by contract, err: [%s], genHash: [%s], contractName: [%s]\n",
					err.Error(), req.GenHash, req.ContractName)
				FailedJSONResp(RespMsgServerError, c)
				return
			}

			resp := make([]*TxListResp, 0)

			if txList == nil {
				SuccessfulJSONRespWithPage(resp, 0, c)
				return
			}

			txCount, err := dao.GetLatestTxCountByContractName(req.GenHash, req.ContractName, TxListByContractNameCount, s.Db())
			if err != nil {
				log.Errorf("fail to get tx count by contract, err: [%s], genHash: [%s], contractName: [%s]\n",
					err.Error(), req.GenHash, req.ContractName)
				FailedJSONResp(RespMsgServerError, c)
				return
			}

			for _, v := range txList {
				tlreq := &TxListResp{
					Id:           v.ID,
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

			SuccessfulJSONRespWithPage(resp, int64(txCount), c)

		} else {
			txList, err := dao.GetTxList(req.GenHash, req.Page, req.PageSize, req.BlockHeight, s.Db())
			if err != nil {
				log.Errorf("fail to get tx list, err: [%s], genHash: [%s], height: [%d]\n",
					err.Error(), req.GenHash, req.BlockHeight)
				FailedJSONResp(RespMsgServerError, c)
				return
			}

			resp := make([]*TxListResp, 0)

			if txList == nil {
				SuccessfulJSONRespWithPage(resp, 0, c)
				return
			}

			for _, v := range txList {
				tlreq := &TxListResp{
					Id:           v.ID,
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
	SenderInfo            string `json:"senderInfo"`
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

		err := checkStringParamsEmpty(req.GenHash)
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

		if txInfo == nil {
			SuccessfulJSONResp(&TxDetailsResp{}, "", nil)
			return
		}

		txDetails, err := dao.GetTxDetails(txInfo.TxId, tableNum, s.Db())
		if err != nil {
			log.Errorf("fail to get tx details, err: [%s], genHash: [%s], txId: [%s]\n",
				err.Error(), req.GenHash, req.TxId)
			FailedJSONResp(RespMsgServerError, c)
			return
		}

		var sender common.EndorsementEntry

		if len(txDetails.SenderBytes) != 0 {
			err := proto.Unmarshal(txDetails.SenderBytes, &sender)
			if err != nil {
				log.Errorf("fail to unmarshal the sender bytes, err: [%s], genHash: [%s], txId: [%s]\n",
					err.Error(), req.GenHash, req.TxId)
				FailedJSONResp(RespMsgServerError, c)
				return
			}
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
			SenderInfo:            string(sender.Signer.MemberInfo),
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
