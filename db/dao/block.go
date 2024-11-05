package dao

import (
	dbModel "chainmscan/db/model"
	"database/sql"
	"fmt"

	"gorm.io/gorm"
)

func MaxBlockHeightInDb(genHash string, gormDb *gorm.DB) (int64, error) {

	tableNum, err := getChainTableNum(genHash, gormDb)
	if err != nil {
		return 0, err
	}

	type Result struct {
		MaxBlockHeight *int64
	}

	var max Result

	if err := gormDb.Table(fmt.Sprintf(dbModel.TableNamePrefix_Block+"_%02d", tableNum)).
		Select("MAX(block_height) AS max_block_height").
		Scan(&max).Error; err != nil {
		return 0, err
	}

	if max.MaxBlockHeight == nil {
		// 新链没有0号区块
		return -1, nil
	}

	return *max.MaxBlockHeight, nil
}

func GetBlockList(genHash string, page, pageSize int32,
	gormDb *gorm.DB) ([]*dbModel.Block, error) {

	var list []*dbModel.Block

	tableNum, err := getChainTableNum(genHash, gormDb)
	if err != nil {
		return list, err
	}

	offset := (page - 1) * pageSize

	err = gormDb.Table(fmt.Sprintf(dbModel.TableNamePrefix_Block+"_%02d", tableNum)).
		Limit(int(pageSize)).Offset(int(offset)).Order("block_height desc").
		Find(&list).Error
	if err != nil {
		return list, err
	}

	return list, nil
}

func GetBlockInfo(genHash string, blockHeight int64, blockHash string,
	gormDb *gorm.DB) (*dbModel.Block, int, error) {

	var block dbModel.Block

	tableNum, err := getChainTableNum(genHash, gormDb)
	if err != nil {
		return nil, 0, err
	}

	queryDb := gormDb.Table(fmt.Sprintf(dbModel.TableNamePrefix_Block+"_%02d", tableNum))

	if len(blockHash) != 0 {
		queryDb = queryDb.Where("block_hash = ?", blockHash)
	}

	if len(blockHash) == 0 && blockHeight > -1 {
		queryDb = queryDb.Where("block_height = ?", blockHeight)
	}

	err = queryDb.First(&block).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, 0, nil
		}

		return nil, 0, err
	}

	return &block, tableNum, nil
}

func GetBlockTxCount(genHash string, blockHeight int64,
	gormDb *gorm.DB) (int64, error) {

	tableNum, err := getChainTableNum(genHash, gormDb)
	if err != nil {
		return 0, err
	}

	var txCount sql.NullInt64

	err = gormDb.Table(fmt.Sprintf(dbModel.TableNamePrefix_Block+"_%02d", tableNum)).
		Select("tx_count").Where("block_height = ?", blockHeight).
		Scan(&txCount).Error
	if err != nil {
		return 0, err
	}

	return txCount.Int64, nil
}
