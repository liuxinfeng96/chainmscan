package dao

import (
	dbModel "chainmscan/db/model"
	"fmt"

	"gorm.io/gorm"
)

func GetTxDetails(txId string, tableNum int,
	gormDb *gorm.DB) (*dbModel.TxDetails, error) {

	var txDetails dbModel.TxDetails

	err := gormDb.Table(fmt.Sprintf(dbModel.TableNamePrefix_TxDetails+"_%02d", tableNum)).
		Where("tx_id = ?", txId).First(&txDetails).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}

		return nil, err
	}

	return &txDetails, nil
}
