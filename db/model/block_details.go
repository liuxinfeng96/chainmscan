package model

import "chainmscan/db"

const TableNamePrefix_BlockDetails = "block_details"

type BlockDetails struct {
	db.CommonField
	BlockHash         string
	ProposerBytes     []byte
	ProposerSignature string
	Dag               string
}

func (t BlockDetails) TableName() string {
	return TableNamePrefix_BlockDetails
}
