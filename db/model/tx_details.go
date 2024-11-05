package model

import "chainmscan/db"

const TableNamePrefix_TxDetails = "tx_details"

/*
CREATE TABLE `tx_details` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime DEFAULT NULL,
  `updated_at` datetime DEFAULT NULL,
  `tx_id` varchar(256) DEFAULT NULL,
  `tx_parameters` longblob,
  `sender_bytes` longblob,
  `endorsers_bytes` longblob,
  `tx_status_code` varchar(256) DEFAULT NULL,
  `rw_set_hash` varchar(256) DEFAULT NULL,
  `tx_message` varchar(256) DEFAULT NULL,
  `contract_result_code` int unsigned DEFAULT NULL,
  `contract_result` longblob,
  `contract_result_message` varchar(256) DEFAULT NULL,
  `gas_used` bigint unsigned DEFAULT NULL,
  `contract_event_bytes` longblob,
  `tx_reads_bytes` longblob,
  `tx_writes_bytes` longblob,
  PRIMARY KEY (`id`),
  UNIQUE INDEX `tx_id_index` (`tx_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
*/

type TxDetails struct {
	db.CommonField
	TxId                  string `gorm:"uniqueIndex:tx_id_index"`
	TxParameters          []byte `gorm:"type:longblob"`
	SenderBytes           []byte `gorm:"type:longblob"`
	EndorsersBytes        []byte `gorm:"type:longblob"`
	TxStatusCode          string
	RwSetHash             string
	TxMessage             string
	ContractResultCode    uint32
	ContractResult        []byte `gorm:"type:longblob"`
	ContractResultMessage string
	GasUsed               uint64
	ContractEventBytes    []byte `gorm:"type:longblob"`
	TxReadsBytes          []byte `gorm:"type:longblob"`
	TxWritesBytes         []byte `gorm:"type:longblob"`
}

func (t TxDetails) TableName() string {
	return TableNamePrefix_TxDetails
}

// func init() {
// 	t := new(TxDetails)
// 	db.TableSlice = append(db.TableSlice, t)
// }
