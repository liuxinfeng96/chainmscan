package server

import (
	"chainmscan/blockchain"

	"gorm.io/gorm"
)

func (s *Server) Subscribe(c *blockchain.BlockChainClient, db *gorm.DB) error {
	return s.subscriber.Subscribe(c, db)
}
