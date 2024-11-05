package model

import "chainmscan/db"

const TableNamePrefix_BlockDetails = "block_details"

/*
CREATE TABLE `block_details` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime DEFAULT NULL,
  `updated_at` datetime DEFAULT NULL,
  `block_hash` varchar(256) DEFAULT NULL,
  `proposer_bytes` longblob,
  `proposer_signature` longtext,
  `dag` longtext,
  PRIMARY KEY (`id`),
  UNIQUE INDEX `block_hash_index` (`block_hash`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
*/

type BlockDetails struct {
	db.CommonField
	BlockHash         string `gorm:"uniqueIndex:block_hash_index"`
	ProposerBytes     []byte `gorm:"type:longblob"`
	ProposerSignature string `gorm:"type:longtext"`
	Dag               string `gorm:"type:longtext"`
}

func (t BlockDetails) TableName() string {
	return TableNamePrefix_BlockDetails
}

// func init() {
// 	t := new(BlockDetails)
// 	db.TableSlice = append(db.TableSlice, t)
// }
