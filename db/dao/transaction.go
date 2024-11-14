package dao

import (
	"chainmscan/db/model"
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

	if tableNum == 0 {
		return nil, nil
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

func GetLatestTxListByContractName(genHash string, contractName string, page, pageSize int32,
	gormDb *gorm.DB) ([]*dbModel.Transaction, error) {

	var list []*dbModel.Transaction

	tableNum, err := getChainTableNum(genHash, gormDb)
	if err != nil {
		return list, err
	}

	if tableNum == 0 {
		return nil, nil
	}

	offset := (page - 1) * pageSize

	queryDb := gormDb.Table(fmt.Sprintf(dbModel.TableNamePrefix_Transaction+"_%02d", tableNum))

	if len(contractName) != 0 {
		queryDb = queryDb.Where("contract_name = ?", contractName)
	}

	err = queryDb.Limit(int(pageSize)).Offset(int(offset)).Order("timestamp desc").Find(&list).Error
	if err != nil {
		return list, err
	}

	return list, nil
}

func GetLatestTxCountByContractName(genHash string, contractName string, limit int,
	gormDb *gorm.DB) (int, error) {

	var txList []*model.Transaction

	tableNum, err := getChainTableNum(genHash, gormDb)
	if err != nil {
		return 0, err
	}

	queryDb := gormDb.Table(fmt.Sprintf(dbModel.TableNamePrefix_Transaction+"_%02d", tableNum))

	if len(contractName) != 0 {
		queryDb = queryDb.Where("contract_name = ?", contractName)
	}

	err = queryDb.Limit(limit).Find(&txList).Error
	if err != nil {
		return 0, err
	}

	if len(txList) < limit {
		return len(txList), nil
	}

	return limit, nil
}

func GetTxInfo(genHash string, txId string, id int,
	gormDb *gorm.DB) (*dbModel.Transaction, int, error) {

	var tx dbModel.Transaction

	tableNum, err := getChainTableNum(genHash, gormDb)
	if err != nil {
		return nil, 0, err
	}

	if tableNum == 0 {
		return nil, 0, nil
	}

	queryDb := gormDb.Table(fmt.Sprintf(dbModel.TableNamePrefix_Transaction+"_%02d", tableNum))

	if len(txId) != 0 {
		queryDb = queryDb.Where("tx_id = ?", txId)
	}

	if id != 0 {
		queryDb = queryDb.Where("id = ?", id)
	}

	err = queryDb.First(&tx).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, 0, nil
		}

		return nil, 0, err
	}

	return &tx, tableNum, nil
}

func GetTxAmountByTime(genHash string,
	startTime, endTime int64, gormDb *gorm.DB) (int64, error) {

	tableNum, err := getChainTableNum(genHash, gormDb)
	if err != nil {
		return 0, err
	}

	if tableNum == 0 {
		return 0, nil
	}

	var txAmount int64

	err = gormDb.Table(fmt.Sprintf(dbModel.TableNamePrefix_Transaction+"_%02d", tableNum)).
		Where("timestamp >= ? AND timestamp < ?", startTime, endTime).
		Count(&txAmount).Error
	if err != nil {
		return 0, err
	}

	return txAmount, nil
}
