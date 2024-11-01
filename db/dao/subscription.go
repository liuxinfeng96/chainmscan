package dao

import (
	dbModel "chainmscan/db/model"

	"gorm.io/gorm"
)

func GetSubscription(genHash string, gormDb *gorm.DB) (*dbModel.Subscription, error) {

	var sub dbModel.Subscription

	err := gormDb.Table(dbModel.TableName_ChainInfo).
		Where("gen_hash = ?", genHash).First(&sub).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}

		return nil, err
	}

	return &sub, nil
}

func SaveInfoOfSubscription(sub *dbModel.Subscription, gormDb *gorm.DB) error {
	return gormDb.Save(sub).Error
}

func GetInfoOfSubscription(genHash string, gormDb *gorm.DB) (*dbModel.Subscription, error) {

	var sub dbModel.Subscription

	err := gormDb.Table(dbModel.TableName_Subscription).
		Where("gen_hash = ?", genHash).First(&sub).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}

		return nil, err
	}

	return &sub, nil
}

func DeleteSubscription(genHash string, gormDb *gorm.DB) error {
	return gormDb.Where("gen_hash = ?", genHash).
		Delete(&dbModel.Subscription{}).Error
}

func GetAllSubscription(gormDb *gorm.DB) ([]*dbModel.Subscription, error) {
	var res []*dbModel.Subscription
	err := gormDb.Table(dbModel.TableName_Subscription).Find(&res).Error
	if err != nil {
		return nil, err
	}

	return res, nil
}
