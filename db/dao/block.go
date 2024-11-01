package dao

import (
	dbModel "chainmscan/db/model"
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
		return 0, nil
	}

	return *max.MaxBlockHeight, nil
}
