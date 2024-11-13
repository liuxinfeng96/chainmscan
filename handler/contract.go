package handler

import (
	"chainmscan/db/dao"
	"chainmscan/server"

	"chainmaker.org/chainmaker/pb-go/v2/accesscontrol"
	"github.com/gin-gonic/gin"
	"github.com/gogo/protobuf/proto"
)

type ContractListHandler struct {
}

type ContractListReq struct {
	PageReq
	GenHash string `json:"genHash"`
}

type ContractListResp struct {
	Id           uint   `json:"id"`
	Name         string `json:"name"`
	Version      string `json:"version"`
	ChainId      string `json:"chainId"`
	RuntimeType  string `json:"runtimeType"`
	State        string `json:"state"`
	CreatorOrgId string `json:"creatorOrgId"`
	Height       uint64 `json:"height"`
	TxTimestamp  int64  `json:"txTimestamp"`
}

func (h *ContractListHandler) Handle(s *server.Server) gin.HandlerFunc {
	return func(c *gin.Context) {

		req := new(ContractListReq)
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

		log, err := s.GetZapLogger("ContractListHandler")
		if err != nil {
			FailedJSONResp(RespMsgLogServerError, c)
			return
		}

		list, total, err := dao.GetContractList(req.GenHash, req.Page, req.PageSize, s.Db())
		if err != nil {
			log.Errorf("fail to get contract list, err: [%s], genHash: [%s]\n", err.Error(), req.GenHash)
			FailedJSONResp(RespMsgServerError, c)
			return
		}

		resp := make([]*ContractListResp, 0)

		for _, v := range list {
			bl := &ContractListResp{
				Id:           v.ID,
				Name:         v.Name,
				Version:      v.Version,
				ChainId:      v.ChainId,
				RuntimeType:  v.RuntimeType,
				State:        v.State,
				CreatorOrgId: v.CreatorOrgId,
				Height:       v.Height,
				TxTimestamp:  v.TxTimestamp,
			}
			resp = append(resp, bl)
		}

		SuccessfulJSONRespWithPage(resp, total, c)
	}
}

type ContractDetailsHandler struct {
}

type ContractDetailsReq struct {
	GenHash      string `json:"genHash"`
	ContractId   uint   `json:"contractId"`
	ContractName string `json:"contractName"`
}

type ContractDetailsResp struct {
	Name         string `json:"name"`
	Version      string `json:"version"`
	ChainId      string `json:"chainId"`
	RuntimeType  string `json:"runtimeType"`
	State        string `json:"state"`
	CreatorOrgId string `json:"creatorOrgId"`
	Address      string `json:"address"`
	TxId         string `json:"txId"`
	Height       uint64 `json:"height"`
	TxTimestamp  int64  `json:"txTimestamp"`
	Creator      string `json:"creator"`
}

func (h *ContractDetailsHandler) Handle(s *server.Server) gin.HandlerFunc {
	return func(c *gin.Context) {

		req := new(ContractDetailsReq)
		if err := c.ShouldBindJSON(req); err != nil {
			FailedJSONResp(RespMsgParamsTypeError, c)
			return
		}

		err := checkStringParamsEmpty(req.GenHash)
		if err != nil {
			FailedJSONResp(RespMsgParamsMissing, c)
			return
		}

		log, err := s.GetZapLogger("ContractDetailsHandler")
		if err != nil {
			FailedJSONResp(RespMsgLogServerError, c)
			return
		}

		contract, err := dao.GetContractInfo(req.GenHash, req.ContractName, req.ContractId, s.Db())
		if err != nil {
			log.Errorf("fail to get contract info, err: [%s], contractId: [%d]\n", err.Error(),
				req.GenHash, req.ContractId)
			FailedJSONResp(RespMsgServerError, c)
			return
		}

		var creator accesscontrol.MemberFull

		err = proto.Unmarshal(contract.CreatorBytes, &creator)
		if err != nil {
			log.Errorf("fail to unmarshal contract creator, err: [%s], contractId: [%d]\n", err.Error(),
				req.GenHash, req.ContractId)
			FailedJSONResp(RespMsgServerError, c)
			return
		}

		resp := &ContractDetailsResp{
			Name:         contract.Name,
			Version:      contract.Version,
			ChainId:      contract.ChainId,
			RuntimeType:  contract.RuntimeType,
			State:        contract.State,
			CreatorOrgId: contract.CreatorOrgId,
			Address:      contract.CreatorOrgId,
			TxId:         contract.TxId,
			Height:       contract.Height,
			TxTimestamp:  contract.TxTimestamp,
			Creator:      string(creator.MemberInfo),
		}

		SuccessfulJSONResp(resp, "", c)
	}
}
