package dao

import (
	"chainmscan/db"
	"database/sql"

	"gorm.io/gorm"
)

// QueryCondition 通用查询条件
type QueryCondition struct {
	// 查询条件列名称
	Column string
	// 查询条件输入
	Input interface{}
}

// QueryObjectList 列表查询带分页
// 返回sqlRows,根据业务解析
func QueryObjectList(gormDb *gorm.DB, object db.DbModel, page, pageSize int32,
	qc ...*QueryCondition) (sqlRow *sql.Rows, err error) {

	offset := (page - 1) * pageSize

	querySub := gormDb.Model(object).Select("id").
		Limit(int(pageSize)).Offset(int(offset)).Order("id desc")

	if qc != nil {
		for i := 0; i < len(qc); i++ {
			if qc[i] != nil {
				querySub = querySub.Where(qc[i].Column+" = ?", qc[i].Input)
			}
		}
	}

	return gormDb.Model(object).Order("id desc").Joins("INNER JOIN (?) AS t2 USING (id)", querySub).Rows()
}

// QueryObjectListTotal 列表查询分页求总数
// 传入通道，为了效率并发操作
func QueryObjectListTotal(gormDb *gorm.DB, object db.DbModel,
	totalChan chan int64, qc ...*QueryCondition) {

	var total int64

	totalSub := gormDb.Model(object)

	if qc != nil {
		for i := 0; i < len(qc); i++ {
			if qc[i] != nil {
				totalSub = totalSub.Where(qc[i].Column+" = ?", qc[i].Input)
			}
		}
	}

	totalSub.Count(&total)

	totalChan <- total

}

// QueryObjectListByCondition 列表查询不带分页
// 返回sqlRows,根据业务解析
func QueryObjectListByCondition(gormDb *gorm.DB, object db.DbModel,
	conditions ...*QueryCondition) (*sql.Rows, error) {

	db := gormDb.Model(object)

	for _, c := range conditions {
		db = db.Where(c.Column+" = ?", c.Input)
	}

	return db.Rows()
}

// QueryObjectByCondition 单条查询，无数据报错
func QueryObjectByCondition(gormDb *gorm.DB, object db.DbModel,
	conditions ...*QueryCondition) error {

	db := gormDb.Model(object)

	for _, c := range conditions {
		db = db.Where(c.Column+" = ?", c.Input)
	}

	if err := db.First(object).Error; err != nil {
		return err
	}

	return nil
}
