package model

import "chainmscan/db"

const TableNamePrefix_Block = "block"

/*
CREATE TABLE `block` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime DEFAULT NULL,
  `updated_at` datetime DEFAULT NULL,
  `block_height` bigint unsigned DEFAULT NULL,
  `block_hash` varchar(256) DEFAULT NULL,
  `chain_id` varchar(256) DEFAULT NULL,
  `pre_block_hash` varchar(256) DEFAULT NULL,
  `block_type` varchar(256) DEFAULT NULL,
  `block_version` int unsigned DEFAULT NULL,
  `pre_conf_height` bigint unsigned DEFAULT NULL,
  `tx_count` int unsigned DEFAULT NULL,
  `tx_root` varchar(256) DEFAULT NULL,
  `dag_hash` varchar(256) DEFAULT NULL,
  `rw_set_root` varchar(256) DEFAULT NULL,
  `block_timestamp` bigint DEFAULT NULL,
  `proposer_org_id` varchar(256) DEFAULT NULL,
  `consensus_args` varchar(256) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE INDEX `block_hash_index` (`block_hash`),
  INDEX `block_timestamp_index` (`block_timestamp`),
  INDEX `block_height_index` (`block_height`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
*/

type Block struct {
	db.CommonField
	BlockHeight    uint64 `json:"blockHeight" gorm:"index:block_height_index"`
	BlockHash      string `json:"blockHash" gorm:"uniqueIndex:block_hash_index"`
	ChainId        string `json:"chainId"`
	PreBlockHash   string `json:"preBlockHash"`
	BlockType      string `json:"blockType"`
	BlockVersion   uint32 `json:"blockVersion"`
	PreConfHeight  uint64 `json:"preConfHeight"`
	TxCount        uint32 `json:"txCount"`
	TxRoot         string `json:"txRoot"`
	DagHash        string `json:"dagHash"`
	RwSetRoot      string `json:"rwSetRoot"`
	BlockTimestamp int64  `json:"blockTimestamp" gorm:"index:block_timestamp_index"`
	ProposerOrgId  string `json:"proposerOrgId"`
	ConsensusArgs  string `json:"consensusArgs"`
}

func (t Block) TableName() string {
	return TableNamePrefix_Block
}

// func init() {
// 	t := new(Block)
// 	db.TableSlice = append(db.TableSlice, t)
// }
