package model

import "chainmscan/db"

const TableName_ChainInfo = "chain_info"

type ChainInfo struct {
	db.CommonField
	GenHash     string
	ChainId     string
	TableNum    int
	TxAmount    int
	BlockAmount int
}

func (t ChainInfo) TableName() string {
	return TableName_ChainInfo
}
