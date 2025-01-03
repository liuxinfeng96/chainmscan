package server

import (
	"chainmscan/blockchain"
	"chainmscan/db/dao"
	"context"
	"errors"
	"fmt"

	"chainmaker.org/chainmaker/pb-go/v2/common"
	"gorm.io/gorm"

	dbModel "chainmscan/db/model"
)

func (s *Server) Subscribe(c *blockchain.BlockChainClient, chainName string) error {
	// 判断是否已经订阅
	chainGenHash := c.GetChainGenHash()

	s.chainListMapMutex.Lock()
	defer s.chainListMapMutex.Unlock()

	_, ok := s.chainList[chainGenHash]
	if ok {
		return nil
	}

	// 查询是否是第一次订阅
	chainInfo, err := dao.GetChainInfo(chainGenHash, s.gormDb)
	if err != nil {
		return errors.New("query chain info err, " + err.Error())
	}

	var tableNum int

	var blockC <-chan interface{}

	closedC := make(chan string)

	if chainInfo == nil {
		// 第一次订阅表后缀序号递增（需要提前数据库分好表）
		maxTableNum, err := dao.GetMaxTableNumOfChainInfo(s.gormDb)
		if err != nil {
			return errors.New("query the sub info err, " + err.Error())
		}

		tableNum = maxTableNum + 1

		// 从0号区块开始订阅
		blockC, err = c.GetChainMakerClient().SubscribeBlock(s.ctx, 0, -1, false, false)
		if err != nil {
			return errors.New("fail to subscribe block, " + err.Error())
		}

		chainInfo = new(dbModel.ChainInfo)

		chainInfo.TableNum = tableNum
		chainInfo.ChainId = c.GetConfig().ChainId
		chainInfo.GenHash = c.GetChainGenHash()

		err = dao.InsertOneObjectToDB(chainInfo, s.gormDb)
		if err != nil {
			return errors.New("insert chaininfo to db err, " + err.Error())
		}

	} else {

		// 不是初次订阅需要索引到该链的分表，查询当前库内区块高度

		tableNum = chainInfo.TableNum

		maxHeight, err := dao.MaxBlockHeightInDb(chainGenHash, s.gormDb)
		if err != nil {
			return errors.New("fail to get max block height in db, " + err.Error())
		}

		blockC, err = c.GetChainMakerClient().SubscribeBlock(s.ctx, int64(maxHeight+1), -1, false, false)
		if err != nil {
			return errors.New("fail to subscribe block, " + err.Error())
		}

	}

	// 开启订阅监听
	s.workerPool.Submit(s.listen(closedC))

	// 开启区块监听
	s.workerPool.Submit(s.startProcess(chainGenHash, tableNum, closedC, blockC, s.gormDb))

	// 新增订阅列表
	s.chainList[chainGenHash] = closedC

	// 数据库更新订阅配置
	sub, err := dao.GetInfoOfSubscription(chainGenHash, s.gormDb)
	if err != nil {
		return errors.New("query the sub info err, " + err.Error())
	}

	if sub == nil {
		sub = &dbModel.Subscription{}
	}

	sub.ChainName = chainName
	sub.GenHash = chainGenHash
	sub.ChainId = c.GetConfig().ChainId
	sub.OrgId = c.GetConfig().OrgId
	sub.NodeAddr = c.GetConfig().NodeConfs[0].Addr
	sub.NodeUseTls = c.GetConfig().NodeConfs[0].UseTls
	if sub.NodeUseTls {
		sub.NodeCaCertPem = c.GetConfig().NodeConfs[0].CaCertPem[0]
		sub.NodeTlsHostName = c.GetConfig().NodeConfs[0].TlsHostName
	}
	sub.SignCertPem = string(c.GetConfig().SignCertBytes)
	sub.SignKeyPem = string(c.GetConfig().SignKeyBytes)
	sub.TlsCertPem = string(c.GetConfig().TlsCertBytes)
	sub.TlsKeyPem = string(c.GetConfig().TlsKeyBytes)
	sub.ArchiveCenterUrl = c.GetConfig().ArchiveCenterUrl

	return dao.SaveInfoOfSubscription(sub, s.gormDb)
}

func (s *Server) UnSubscribe(genHash string, db *gorm.DB) error {
	s.chainListMapMutex.Lock()
	defer s.chainListMapMutex.Unlock()

	delete(s.chainList, genHash)

	// 主动删除订阅配置
	return dao.DeleteSubscription(genHash, db)
}

func (s *Server) GetChainList() []string {
	s.chainListMapMutex.Lock()
	defer s.chainListMapMutex.Unlock()

	var list []string

	for k := range s.chainList {
		list = append(list, k)
	}

	return list
}

func (s *Server) startProcess(genHash string, tableNum int,
	closedSignal chan<- string, c <-chan interface{},
	gormDb *gorm.DB) func(ctx context.Context) error {

	return func(ctx context.Context) error {
		for {
			select {
			case block, ok := <-c:
				if !ok {
					closedSignal <- genHash
					err := errors.New("the chan of subscriber is closed")
					s.SysLog().Error(err.Error())
					return err
				}

				blockInfo, ok := block.(*common.BlockInfo)
				if !ok {
					closedSignal <- genHash
					err := errors.New("the block info type error")
					s.SysLog().Error(err.Error())
					return err
				}

				err := blockchain.StorageBlock(blockInfo, genHash, tableNum, gormDb)
				if err != nil {
					closedSignal <- genHash
					err := fmt.Errorf("fail to storage block, err: [%s]", err.Error())
					s.SysLog().Error(err.Error())
					return err
				}

			case <-ctx.Done():
				s.SysLog().Infof("the chain subscriber has been closed, genHash: [%s]\n", genHash)
				return nil
			}
		}
	}
}

func (s *Server) listen(closedSignal <-chan string) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		for {
			select {
			case genHash := <-closedSignal:
				// 将链状态更新为未订阅状态，走重新订阅逻辑。

				s.chainListMapMutex.Lock()
				delete(s.chainList, genHash)
				s.chainListMapMutex.Unlock()

				s.SysLog().Warnf("the subscriber is closed, genHash: [%s]\n", genHash)

				return nil

			case <-ctx.Done():
				s.SysLog().Info("the subscriber's listening has been closed ...")
				return nil
			}
		}
	}
}

func (s *Server) SubscriberStart() error {
	s.chainListMapMutex.Lock()
	defer s.chainListMapMutex.Unlock()

	// 查询订阅配置列表
	list, err := dao.GetAllSubscription(s.gormDb)
	if err != nil {
		return errors.New("fail to get all subscription, " + err.Error())
	}

	sdkLog, err := s.GetZapLogger("BCSDK")
	if err != nil {
		return err
	}

	// 依次重新开启订阅
	for _, v := range list {
		// 建立客户端
		config := &blockchain.ClientConfig{
			ChainId:       v.ChainId,
			OrgId:         v.OrgId,
			SignKeyBytes:  []byte(v.SignKeyPem),
			SignCertBytes: []byte(v.SignCertPem),
			NodeConfs: []*blockchain.NodeConnConfig{
				{
					Addr:        v.NodeAddr,
					CaCertPem:   []string{v.NodeCaCertPem},
					TlsHostName: v.NodeTlsHostName,
					UseTls:      v.NodeUseTls,
				},
			},
			TlsKeyBytes:      []byte(v.TlsKeyPem),
			TlsCertBytes:     []byte(v.TlsCertPem),
			ArchiveCenterUrl: v.ArchiveCenterUrl,
			Logger:           sdkLog,
		}

		c, err := blockchain.NewChainmakerClient(config)
		if err != nil {
			return errors.New("create chainmaker client err, " + err.Error())
		}

		maxHeight, err := dao.MaxBlockHeightInDb(c.GetChainGenHash(), s.gormDb)
		if err != nil {
			return errors.New("fail to get max block height in db, " + err.Error())
		}

		blockC, err := c.GetChainMakerClient().SubscribeBlock(s.ctx, int64(maxHeight+1), -1, false, false)
		if err != nil {
			return errors.New("fail to subscribe block, " + err.Error())
		}

		chainInfo, err := dao.GetChainInfo(c.GetChainGenHash(), s.gormDb)
		if err != nil {
			return errors.New("query chain info err, " + err.Error())
		}

		if chainInfo == nil {
			return errors.New("the chain info does not exist")
		}

		closedC := make(chan string)
		// 开启订阅监听
		s.workerPool.Submit(s.listen(closedC))

		// 开启区块监听
		s.workerPool.Submit(s.startProcess(c.GetChainGenHash(), chainInfo.TableNum, closedC, blockC, s.gormDb))

		// 新增订阅列表
		s.chainList[c.GetChainGenHash()] = closedC
	}

	return nil
}
