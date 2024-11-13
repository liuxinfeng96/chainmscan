package server

import (
	"chainmscan/blockchain"

	"gorm.io/gorm"
)

func (s *Server) Subscribe(c *blockchain.BlockChainClient, chainName string, db *gorm.DB) error {
	return s.subscriber.Subscribe(c, chainName, db)
}

func (s *Server) UnSubscribe(genHash string, db *gorm.DB) error {
	return s.subscriber.Unsubscribe(genHash, db)
}

func (s *Server) GetChainList() []string {
	return s.subscriber.GetChainList()
}
