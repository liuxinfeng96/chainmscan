package model

import "chainmscan/db"

const TableNamePrefix_Block = "block"

type Block struct {
	db.CommonField
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
	ConsensusArgs  string `json:"consensusArgs"`
}

func (t Block) TableName() string {
	return TableNamePrefix_BlockDetails
}
