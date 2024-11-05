package dao

import (
	dbModel "chainmscan/db/model"
	"fmt"

	"gorm.io/gorm"
)

func GetTxList(genHash string, page, pageSize int32, blockHeight int64,
	gormDb *gorm.DB) ([]*dbModel.Transaction, error) {

	var list []*dbModel.Transaction

	tableNum, err := getChainTableNum(genHash, gormDb)
	if err != nil {
		return list, err
	}

	offset := (page - 1) * pageSize

	queryDb := gormDb.Table(fmt.Sprintf(dbModel.TableNamePrefix_Transaction+"_%02d", tableNum))

	if blockHeight > 0 {
		queryDb = queryDb.Where("block_height = ?", blockHeight)
	}

	err = queryDb.Limit(int(pageSize)).Offset(int(offset)).Order("timestamp desc").
		Find(&list).Error

	if err != nil {
		return list, err
	}

	return list, nil
}

func GetLatestTxListByContractName(genHash string, contractName string, limit int,
	gormDb *gorm.DB) ([]*dbModel.Transaction, error) {

	var list []*dbModel.Transaction

	tableNum, err := getChainTableNum(genHash, gormDb)
	if err != nil {
		return list, err
	}

	queryDb := gormDb.Table(fmt.Sprintf(dbModel.TableNamePrefix_Transaction+"_%02d", tableNum))

	if len(contractName) != 0 {
		queryDb = queryDb.Where("contract_name = ?", contractName)
	}

	err = queryDb.Limit(limit).Order("timestamp desc").Find(&list).Error
	if err != nil {
		return list, err
	}

	return list, nil
}

func GetTxInfo(genHash string, txId string,
	gormDb *gorm.DB) (*dbModel.Transaction, int, error) {

	var tx dbModel.Transaction

	tableNum, err := getChainTableNum(genHash, gormDb)
	if err != nil {
		return nil, 0, err
	}

	err = gormDb.Table(fmt.Sprintf(dbModel.TableNamePrefix_Transaction+"_%02d", tableNum)).
		Where("tx_id = ?", txId).First(&tx).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, 0, nil
		}

		return nil, 0, err
	}

	return &tx, tableNum, nil
}
