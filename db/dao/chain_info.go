package dao

import (
	dbModel "chainmscan/db/model"
	"database/sql"

	"gorm.io/gorm"
)

func GetChainInfo(genHash string, gormDb *gorm.DB) (*dbModel.ChainInfo, error) {

	var chainInfo dbModel.ChainInfo

	err := gormDb.Table(dbModel.TableName_ChainInfo).
		Where("gen_hash = ?", genHash).First(&chainInfo).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}

		return nil, err
	}

	return &chainInfo, nil
}

func getChainTableNum(genHash string, gormDb *gorm.DB) (int, error) {

	var tableNum int

	err := gormDb.Table(dbModel.TableName_ChainInfo).
		Select("table_num").
		Where("gen_hash = ?", genHash).
		Scan(&tableNum).Error
	if err != nil {
		return tableNum, err
	}

	return tableNum, nil
}

func GetMaxTableNumOfChainInfo(gormDb *gorm.DB) (int, error) {

	var maxTableNum int

	err := gormDb.Table(dbModel.TableName_ChainInfo).
		Select("table_num").
		Order("table_num desc").
		Limit(1).Scan(&maxTableNum).Error
	if err != nil {
		return maxTableNum, err
	}

	return maxTableNum, nil
}

func SaveChainInfo(chainInfo *dbModel.ChainInfo, gormDb *gorm.DB) error {
	return gormDb.Save(chainInfo).Error
}

func UpdateChainTxAndBlockAmount(genHash string,
	txAmount, blockAmount int, gormDb *gorm.DB) error {
	return gormDb.Table(dbModel.TableName_ChainInfo).
		Where("gen_hash = ?", genHash).
		Updates(dbModel.ChainInfo{TxAmount: txAmount, BlockAmount: blockAmount}).Error
}

func GetChainTxAmount(genHash string, gormDb *gorm.DB) (int64, error) {

	var txAmount sql.NullInt64

	err := gormDb.Table(dbModel.TableName_ChainInfo).
		Select("tx_amount").
		Where("gen_hash = ?", genHash).Scan(&txAmount).Error
	if err != nil {
		return 0, err
	}

	return txAmount.Int64, nil
}
