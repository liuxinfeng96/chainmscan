package model

import "chainmscan/db"

const TableName_Subscription = "subscription"

type Subscription struct {
	db.CommonField
	GenHash          string `gorm:"uniqueIndex:gen_hash_index"`
	ChainName        string
	ChainId          string
	OrgId            string
	NodeAddr         string
	NodeCaCertPem    string `gorm:"type:longtext"`
	NodeTlsHostName  string
	NodeUseTls       bool
	SignCertPem      string `gorm:"type:longtext"`
	SignKeyPem       string `gorm:"type:longtext"`
	TlsCertPem       string `gorm:"type:longtext"`
	TlsKeyPem        string `gorm:"type:longtext"`
	ArchiveCenterUrl string
}

func (t Subscription) TableName() string {
	return TableName_Subscription
}

func init() {
	t := new(Subscription)
	db.TableSlice = append(db.TableSlice, t)
}
