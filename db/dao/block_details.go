package dao

import (
	dbModel "chainmscan/db/model"
	"fmt"

	"gorm.io/gorm"
)

func GetBlockDetails(blockHash string, tableNum int,
	gormDb *gorm.DB) (*dbModel.BlockDetails, error) {

	var blockDetails dbModel.BlockDetails

	err := gormDb.Table(fmt.Sprintf(dbModel.TableNamePrefix_BlockDetails+"_%02d", tableNum)).
		Where("block_hash = ?", blockHash).First(&blockDetails).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}

		return nil, err
	}

	return &blockDetails, nil
}
