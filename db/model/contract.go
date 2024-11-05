package model

import "chainmscan/db"

const TableNamePrefix_Contract = "contract"

/*
CREATE TABLE `contract` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime DEFAULT NULL,
  `updated_at` datetime DEFAULT NULL,
  `name` varchar(256) DEFAULT NULL,
  `version` varchar(256) DEFAULT NULL,
  `chain_id` varchar(256) DEFAULT NULL,
  `runtime_type` varchar(256) DEFAULT NULL,
  `state` varchar(256) DEFAULT NULL,
  `creator_org_id` varchar(256) DEFAULT NULL,
  `address` varchar(256) DEFAULT NULL,
  `tx_id` varchar(256) DEFAULT NULL,
  `height` bigint unsigned DEFAULT NULL,
  `tx_timestamp` bigint DEFAULT NULL,
  `creator_bytes` longblob,
  PRIMARY KEY (`id`),
  INDEX `tx_timestamp_index` (`tx_timestamp`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
*/

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
	TxTimestamp  int64  `json:"txTimestamp" gorm:"index:tx_timestamp_index"`
	CreatorBytes []byte `json:"creatorBytes" gorm:"type:longblob"`
}

func (t Contract) TableName() string {
	return TableNamePrefix_Contract
}

// func init() {
// 	t := new(Contract)
// 	db.TableSlice = append(db.TableSlice, t)
// }
