package dao

import (
	"chainmscan/db"

	"gorm.io/gorm"
)

// InsertObjectsToDBInTransaction 多模型/多条数据入库，符合事务逻辑
func InsertObjectsToDBInTransaction(gormDb *gorm.DB, objects []db.DbModel) error {
	tx := gormDb.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Error; err != nil {
		return err
	}

	for i := range objects {
		if err := tx.Create(objects[i]).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}

// InsertOneObjectToDB 单条数据入库
func InsertOneObjectToDB(object db.DbModel, gormDb *gorm.DB) error {
	if err := gormDb.Create(object).Error; err != nil {
		return err
	}
	return nil
}

// InsertOneObjectToDBByTableName 单条数据入库
func InsertOneObjectToDBByTableName(object db.DbModel,
	tableName string, gormDb *gorm.DB) error {
	if err := gormDb.Table(tableName).Create(object).Error; err != nil {
		return err
	}
	return nil
}
