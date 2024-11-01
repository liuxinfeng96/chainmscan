package model

import "chainmscan/db"

const TableNamePrefix_Transaction = "transaction"

type Transaction struct {
	db.CommonField
	TxId           string `json:"txId"`
	BlockHeight    uint64 `json:"blockHeight"`
	ChainId        string `json:"chainId"`
	ContractName   string `json:"contractName"`
	Method         string `json:"method"`
	TxType         string `json:"txType"`
	Timestamp      int64  `json:"timestamp"`
	ExpirationTime int64  `json:"expirationTime"`
	Sequence       uint64 `json:"sequence"`
	GasLimit       uint64 `json:"gasLimit"`
	SenderOrgId    string `json:"senderOrgId"`
	TxStatusCode   string `json:"txStatusCode"`
}

func (t Transaction) TableName() string {
	return TableNamePrefix_Transaction
}
