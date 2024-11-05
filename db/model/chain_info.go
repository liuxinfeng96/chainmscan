package model

import "chainmscan/db"

const TableName_ChainInfo = "chain_info"

type ChainInfo struct {
	db.CommonField
	GenHash     string `json:"genHash" gorm:"uniqueIndex:gen_hash_index"`
	ChainId     string `json:"chainId"`
	TableNum    int    `json:"tableNum"`
	TxAmount    int    `json:"txAmount"`
	BlockAmount int    `json:"blockAmount"`
}

func (t ChainInfo) TableName() string {
	return TableName_ChainInfo
}

func init() {
	t := new(ChainInfo)
	db.TableSlice = append(db.TableSlice, t)
}
