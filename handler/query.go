package handler

import (
	"chainmscan/db/dao"
	"chainmscan/server"
	"strconv"

	"github.com/gin-gonic/gin"
)

type SearchHandler struct {
}

type SearchReq struct {
	Keyword string `json:"keyword"`
	GenHash string `json:"genHash"`
}

const (
	SearchType_Tx = iota
	SearchType_Block
	SearchType_Contract
	SearchType_Unknown
)

type SearchResp struct {
	Type uint32 `json:"type"`
	Id   uint   `json:"id"`
}

func (h *SearchHandler) Handle(s *server.Server) gin.HandlerFunc {
	return func(c *gin.Context) {

		req := new(SearchReq)
		if err := c.ShouldBindJSON(req); err != nil {
			FailedJSONResp(RespMsgParamsTypeError, c)
			return
		}

		err := checkStringParamsEmpty(req.GenHash, req.Keyword)
		if err != nil {
			FailedJSONResp(RespMsgParamsMissing, c)
			return
		}

		if len(req.Keyword) != 64 {
			blockHeight, err := strconv.Atoi(req.Keyword)
			if err == nil {
				block, _, _ := dao.GetBlockInfo(req.GenHash, int64(blockHeight), "", 0, s.Db())
				if block != nil {
					SuccessfulJSONResp(&SearchResp{
						Id:   block.ID,
						Type: SearchType_Block,
					}, "", c)
					return
				}
			}
		} else {
			tx, _, _ := dao.GetTxInfo(req.GenHash, req.Keyword, 0, s.Db())
			if tx != nil {
				SuccessfulJSONResp(&SearchResp{
					Id:   tx.ID,
					Type: SearchType_Tx,
				}, "", c)
				return
			}

			block, _, _ := dao.GetBlockInfo(req.GenHash, 0, req.Keyword, 0, s.Db())
			if block != nil {
				SuccessfulJSONResp(&SearchResp{
					Id:   block.ID,
					Type: SearchType_Block,
				}, "", c)
				return
			}
		}

		SuccessfulJSONResp(&SearchResp{
			Type: SearchType_Unknown,
		}, "", c)
	}
}

type OverviewHandler struct {
}

type OverviewReq struct {
	GenHash string `json:"genHash"`
}

type OverviewResp struct {
	BlockAmount    int `json:"blockAmount"`
	TxAmount       int `json:"txAmount"`
	ContractAmount int `json:"contractAmount"`
}

func (h *OverviewHandler) Handle(s *server.Server) gin.HandlerFunc {
	return func(c *gin.Context) {

		req := new(OverviewReq)
		if err := c.ShouldBindJSON(req); err != nil {
			FailedJSONResp(RespMsgParamsTypeError, c)
			return
		}

		err := checkStringParamsEmpty(req.GenHash)
		if err != nil {
			FailedJSONResp(RespMsgParamsMissing, c)
			return
		}

		log, err := s.GetZapLogger("OverviewHandler")
		if err != nil {
			FailedJSONResp(RespMsgLogServerError, c)
			return
		}

		chainInfo, err := dao.GetChainInfo(req.GenHash, s.Db())
		if err != nil {
			log.Errorf("fail to get chain info, err: [%s], genHash: [%s]\n", err.Error(), req.GenHash)
			FailedJSONResp(RespMsgServerError, c)
			return
		}

		if chainInfo == nil {
			log.Errorf("fail to get chain info, the chain info does not exist, genHash: [%s]\n", req.GenHash)
			FailedJSONResp(RespMsgServerError, c)
			return
		}

		_, contractAmount, err := dao.GetContractList(req.GenHash, 1, 10, s.Db())
		if err != nil {
			log.Errorf("fail to get contract list, err: [%s], genHash: [%s]\n", err.Error(), req.GenHash)
			FailedJSONResp(RespMsgServerError, c)
			return
		}

		SuccessfulJSONResp(&OverviewResp{
			TxAmount:       chainInfo.TxAmount,
			BlockAmount:    chainInfo.BlockAmount,
			ContractAmount: int(contractAmount),
		}, "", c)
	}
}
