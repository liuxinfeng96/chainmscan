package handler

import (
	"chainmscan/db/dao"
	"chainmscan/server"

	"chainmaker.org/chainmaker/pb-go/v2/accesscontrol"
	"github.com/gin-gonic/gin"
	"github.com/gogo/protobuf/proto"
)

type BlockListHandler struct {
}

type BlockListReq struct {
	PageReq
	GenHash string `json:"genHash"`
}

type BlockListResp struct {
	BlockHeight    uint64 `json:"blockHeight"`
	BlockHash      string `json:"blockHash"`
	ChainId        string `json:"chainId"`
	PreBlockHash   string `json:"preBlockHash"`
	TxCount        uint32 `json:"txCount"`
	TxRoot         string `json:"txRoot"`
	BlockTimestamp int64  `json:"blockTimestamp"`
	ProposerOrgId  string `json:"proposerOrgId"`
}

func (h *BlockListHandler) Handle(s *server.Server) gin.HandlerFunc {
	return func(c *gin.Context) {

		req := new(BlockListReq)
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

		log, err := s.GetZapLogger("BlockListHandler")
		if err != nil {
			FailedJSONResp(RespMsgLogServerError, c)
			return
		}

		list, err := dao.GetBlockList(req.GenHash, req.Page, req.PageSize, s.Db())
		if err != nil {
			log.Errorf("fail to get block list, err: [%s], genHash: [%s]\n", err.Error(), req.GenHash)
			FailedJSONResp(RespMsgServerError, c)
			return
		}

		chainInfo, err := dao.GetChainInfo(req.GenHash, s.Db())
		if err != nil {
			log.Errorf("fail to get chainInfo, err: [%s], genHash: [%s]\n", err.Error(), req.GenHash)
			FailedJSONResp(RespMsgServerError, c)
			return
		}

		resp := make([]*BlockListResp, 0)

		for _, v := range list {
			bl := &BlockListResp{
				BlockHeight:    v.BlockHeight,
				BlockHash:      v.BlockHash,
				ChainId:        v.ChainId,
				PreBlockHash:   v.PreBlockHash,
				TxCount:        v.TxCount,
				TxRoot:         v.TxRoot,
				BlockTimestamp: v.BlockTimestamp,
				ProposerOrgId:  v.ProposerOrgId,
			}
			resp = append(resp, bl)
		}

		SuccessfulJSONRespWithPage(resp, int64(chainInfo.BlockAmount), c)
	}
}

type BlockDetailsHandler struct {
}

type BlockDetailsReq struct {
	GenHash     string `json:"genHash"`
	BlockHeight int64  `json:"blockHeight"`
	BlockHash   string `json:"blockHash"`
	Id          int    `json:"id"`
}

type BlockDetailsResp struct {
	BlockHeight    uint64 `json:"blockHeight"`
	BlockHash      string `json:"blockHash"`
	ChainId        string `json:"chainId"`
	PreBlockHash   string `json:"preBlockHash"`
	BlockType      string `json:"blockType"`
	BlockVersion   uint32 `json:"blockVersion"`
	PreConfHeight  uint64 `json:"preConfHeight"`
	TxCount        uint32 `json:"txCount"`
	TxRoot         string `json:"txRoot"`
	DagHash        string `json:"dagHash"`
	RwSetRoot      string `json:"rwSetRoot"`
	BlockTimestamp int64  `json:"blockTimestamp"`
	ProposerOrgId  string `json:"proposerOrgId"`
	Proposer       string `json:"proposer"`
}

func (h *BlockDetailsHandler) Handle(s *server.Server) gin.HandlerFunc {
	return func(c *gin.Context) {

		req := new(BlockDetailsReq)
		if err := c.ShouldBindJSON(req); err != nil {
			FailedJSONResp(RespMsgParamsTypeError, c)
			return
		}

		err := checkStringParamsEmpty(req.GenHash)
		if err != nil {
			FailedJSONResp(RespMsgParamsMissing, c)
			return
		}

		if len(req.BlockHash) == 0 && req.BlockHeight == -1 && req.Id == 0 {
			FailedJSONResp(RespMsgParamsMissing, c)
			return
		}

		log, err := s.GetZapLogger("BlockDetailsHandler")
		if err != nil {
			FailedJSONResp(RespMsgLogServerError, c)
			return
		}

		block, tableNum, err := dao.GetBlockInfo(req.GenHash, req.BlockHeight, req.BlockHash,
			req.Id, s.Db())
		if err != nil {
			log.Errorf("fail to get block info, err: [%s], req: [%+v]\n", err.Error(), req)
			FailedJSONResp(RespMsgServerError, c)
			return
		}

		details, err := dao.GetBlockDetails(block.BlockHash, tableNum, s.Db())
		if err != nil {
			log.Errorf("fail to get block details, err: [%s], req: [%+v]\n", err.Error(), req)
			FailedJSONResp(RespMsgServerError, c)
			return
		}

		var member accesscontrol.Member
		err = proto.Unmarshal(details.ProposerBytes, &member)
		if err != nil {
			log.Errorf("fail to unmarshal the proposer, err: [%s], req: [%+v]\n", err.Error(), req)
			FailedJSONResp(RespMsgServerError, c)
			return
		}

		resp := &BlockDetailsResp{
			BlockHeight:    block.BlockHeight,
			BlockHash:      block.BlockHash,
			ChainId:        block.ChainId,
			PreBlockHash:   block.PreBlockHash,
			BlockType:      block.BlockType,
			BlockVersion:   block.BlockVersion,
			PreConfHeight:  block.PreConfHeight,
			TxCount:        block.TxCount,
			TxRoot:         block.TxRoot,
			DagHash:        block.DagHash,
			RwSetRoot:      block.RwSetRoot,
			BlockTimestamp: block.BlockTimestamp,
			ProposerOrgId:  block.ProposerOrgId,
			Proposer:       string(member.MemberInfo),
		}

		SuccessfulJSONResp(resp, "", c)
	}
}
