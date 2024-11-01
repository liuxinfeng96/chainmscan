package model

import "chainmscan/db"

const TableNamePrefix_Contract = "contract"

type Contract struct {
	db.CommonField
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
	CreatorBytes []byte
}

func (t Contract) TableName() string {
	return TableNamePrefix_Contract
}
