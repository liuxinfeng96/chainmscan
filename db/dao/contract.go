package dao

import (
	dbModel "chainmscan/db/model"
	"fmt"

	"gorm.io/gorm"
)

func GetContractList(genHash string, page, pageSize int32,
	gormDb *gorm.DB) ([]*dbModel.Contract, int64, error) {

	var list []*dbModel.Contract

	tableNum, err := getChainTableNum(genHash, gormDb)
	if err != nil {
		return list, 0, err
	}

	var total int64

	sql1 := gormDb.Table(fmt.Sprintf(dbModel.TableNamePrefix_Contract+"_%02d", tableNum)).
		Select("name,MAX(tx_timestamp) AS time").Group("name")

	err = gormDb.Table(fmt.Sprintf(dbModel.TableNamePrefix_Contract+"_%02d AS c1", tableNum)).
		Joins("INNER JOIN (?) AS c2 ON c1.tx_timestamp = c2.time", sql1).Count(&total).Error
	if err != nil {
		return list, 0, err
	}

	offset := (page - 1) * pageSize
	err = gormDb.Table(fmt.Sprintf(dbModel.TableNamePrefix_Contract+"_%02d AS c1", tableNum)).
		Joins("INNER JOIN (?) AS c2 ON c1.tx_timestamp = c2.time", sql1).
		Limit(int(pageSize)).Offset(int(offset)).Find(&list).Error
	if err != nil {
		return list, 0, err
	}

	return list, total, nil
}

func GetContractInfo(genHash string, id uint,
	gormDb *gorm.DB) (*dbModel.Contract, error) {

	var contract dbModel.Contract

	tableNum, err := getChainTableNum(genHash, gormDb)
	if err != nil {
		return nil, err
	}

	err = gormDb.Table(fmt.Sprintf(dbModel.TableNamePrefix_Contract+"_%02d", tableNum)).
		Where("id = ?", id).First(&contract).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}

		return nil, err
	}

	return &contract, nil
}
