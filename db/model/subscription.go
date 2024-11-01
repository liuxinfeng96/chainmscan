package model

import "chainmscan/db"

const TableName_Subscription = "subscription"

type Subscription struct {
	db.CommonField
	GenHash          string
	ChainId          string
	OrgId            string
	NodeAddr         string
	NodeCaCertPem    string
	NodeTlsHostName  string
	NodeUseTls       bool
	SignCertPem      string
	SignKeyPem       string
	TlsCertPem       string
	TlsKeyPem        string
	ArchiveCenterUrl string
}

func (t Subscription) TableName() string {
	return TableName_Subscription
}
