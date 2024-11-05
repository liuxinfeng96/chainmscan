package model

import "chainmscan/db"

const TableNamePrefix_Transaction = "transaction"

/*
CREATE TABLE `transaction` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime DEFAULT NULL,
  `updated_at` datetime DEFAULT NULL,
  `tx_id` varchar(256) DEFAULT NULL,
  `block_height` bigint unsigned DEFAULT NULL,
  `chain_id` varchar(256) DEFAULT NULL,
  `contract_name` varchar(256) DEFAULT NULL,
  `method` varchar(256) DEFAULT NULL,
  `tx_type` varchar(256) DEFAULT NULL,
  `timestamp` bigint DEFAULT NULL,
  `expiration_time` bigint DEFAULT NULL,
  `sequence` bigint unsigned DEFAULT NULL,
  `gas_limit` bigint unsigned DEFAULT NULL,
  `sender_org_id` varchar(256) DEFAULT NULL,
  `tx_status_code` varchar(256) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE INDEX `tx_id_index` (`tx_id`),
  INDEX `block_height_index` (`block_height`),
  INDEX `contract_name_index` (`contract_name`),
  INDEX `timestamp_index` (`timestamp`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
*/

type Transaction struct {
	db.CommonField
	TxId           string `json:"txId" gorm:"uniqueIndex:tx_id_index"`
	BlockHeight    uint64 `json:"blockHeight" gorm:"index:block_height_index"`
	ChainId        string `json:"chainId"`
	ContractName   string `json:"contractName" gorm:"index:contract_name_index"`
	Method         string `json:"method"`
	TxType         string `json:"txType"`
	Timestamp      int64  `json:"timestamp" gorm:"index:timestamp_index"`
	ExpirationTime int64  `json:"expirationTime"`
	Sequence       uint64 `json:"sequence"`
	GasLimit       uint64 `json:"gasLimit"`
	SenderOrgId    string `json:"senderOrgId"`
	TxStatusCode   string `json:"txStatusCode"`
}

func (t Transaction) TableName() string {
	return TableNamePrefix_Transaction
}

// func init() {
// 	t := new(Transaction)
// 	db.TableSlice = append(db.TableSlice, t)
// }
