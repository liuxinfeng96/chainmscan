package model

import "chainmscan/db"

const TableNamePrefix_TxDetails = "tx_details"

type TxDetails struct {
	db.CommonField
	TxId                  string
	TxParameters          []byte
	SenderBytes           []byte
	EndorsersBytes        []byte
	TxStatusCode          string
	RwSetHash             string
	TxMessage             string
	ContractResultCode    uint32
	ContractResult        string
	ContractResultMessage string
	GasUsed               uint64
	ContractEventBytes    []byte
	TxReadsBytes          []byte
	TxWritesBytes         []byte
}

func (t TxDetails) TableName() string {
	return TableNamePrefix_TxDetails
}
