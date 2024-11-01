package blockchain

import (
	"chainmscan/db/dao"
	dbModel "chainmscan/db/model"
	"context"
	"errors"
	"sync"

	"chainmaker.org/chainmaker/pb-go/v2/common"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Subscriber struct {
	ctx               context.Context
	chainList         map[string]chan<- string
	chainListMapMutex sync.Mutex
	log               *zap.SugaredLogger
}

func NewSubscriber(ctx context.Context, log *zap.SugaredLogger) (*Subscriber, error) {
	s := new(Subscriber)
	s.ctx = ctx
	s.chainList = make(map[string]chan<- string)
	s.chainListMapMutex = sync.Mutex{}
	s.log = log

	return s, nil
}

func (s *Subscriber) Subscribe(c *BlockChainClient, db *gorm.DB) error {
	// 判断是否已经订阅
	chainGenHash := c.GetChainGenHash()

	s.chainListMapMutex.Lock()
	defer s.chainListMapMutex.Unlock()

	_, ok := s.chainList[chainGenHash]
	if ok {
		return nil
	}

	// 查询是否是第一次订阅
	chainInfo, err := dao.GetChainInfo(chainGenHash, db)
	if err != nil {
		return errors.New("query chain info err, " + err.Error())
	}

	var tableNum int

	var blockC <-chan interface{}

	closedC := make(chan string)

	if chainInfo == nil {
		// 第一次订阅表后缀序号递增（需要提前数据库分好表）
		maxTableNum, err := dao.GetMaxTableNumOfChainInfo(db)
		if err != nil {
			return errors.New("query the sub info err, " + err.Error())
		}

		tableNum = maxTableNum + 1

		// 从0号区块开始订阅
		blockC, err = c.GetChainMakerClient().SubscribeBlock(s.ctx, 0, -1, false, false)
		if err != nil {
			return errors.New("failed to subscribe block, " + err.Error())
		}

		chainInfo = new(dbModel.ChainInfo)

		chainInfo.TableNum = tableNum

		err = dao.InsertOneObjectToDB(chainInfo, db)
		if err != nil {
			return errors.New("insert chaininfo to db err, " + err.Error())
		}

	} else {

		// 不是初次订阅需要索引到该链的分表，查询当前库内区块高度

		tableNum = chainInfo.TableNum

		maxHeight, err := dao.MaxBlockHeightInDb(chainGenHash, db)
		if err != nil {
			return errors.New("failed to get max block height in db, " + err.Error())
		}

		blockC, err = c.GetChainMakerClient().SubscribeBlock(s.ctx, int64(maxHeight+1), -1, false, false)
		if err != nil {
			return errors.New("failed to subscribe block, " + err.Error())
		}

	}

	// 开启订阅监听
	s.listen(closedC)

	// 开启区块监听
	s.startProcess(chainGenHash, tableNum, closedC, blockC, db)

	// 新增订阅列表
	s.chainList[chainGenHash] = closedC

	// 数据库更新订阅配置
	sub, err := dao.GetInfoOfSubscription(chainGenHash, db)
	if err != nil {
		return errors.New("query the sub info err, " + err.Error())
	}

	if sub == nil {
		sub = &dbModel.Subscription{}
	}

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

	return dao.SaveInfoOfSubscription(sub, db)
}

func (s *Subscriber) Unsubscribe(genHash string, gormDb *gorm.DB) error {
	s.chainListMapMutex.Lock()
	defer s.chainListMapMutex.Unlock()

	delete(s.chainList, genHash)

	// 主动删除订阅配置
	return dao.DeleteSubscription(genHash, gormDb)
}

func (s *Subscriber) Start(gormDb *gorm.DB) error {
	s.chainListMapMutex.Lock()
	defer s.chainListMapMutex.Unlock()

	// 查询订阅配置列表
	list, err := dao.GetAllSubscription(gormDb)
	if err != nil {
		return errors.New("failed to get all subscription, " + err.Error())
	}
	// 依次重新开启订阅
	for _, v := range list {
		// 建立客户端
		config := &ClientConfig{
			ChainId:       v.ChainId,
			OrgId:         v.OrgId,
			SignKeyBytes:  []byte(v.SignKeyPem),
			SignCertBytes: []byte(v.SignCertPem),
			NodeConfs: []*NodeConnConfig{
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
		}

		c, err := NewChainmakerClient(config)
		if err != nil {
			return errors.New("create chainmaker client err, " + err.Error())
		}

		maxHeight, err := dao.MaxBlockHeightInDb(c.GetChainGenHash(), gormDb)
		if err != nil {
			return errors.New("failed to get max block height in db, " + err.Error())
		}

		blockC, err := c.GetChainMakerClient().SubscribeBlock(s.ctx, int64(maxHeight+1), -1, false, false)
		if err != nil {
			return errors.New("failed to subscribe block, " + err.Error())
		}

		chainInfo, err := dao.GetChainInfo(c.GetChainGenHash(), gormDb)
		if err != nil {
			return errors.New("query chain info err, " + err.Error())
		}

		if chainInfo == nil {
			return errors.New("the chain info does not exist")
		}

		closedC := make(chan string)
		// 开启订阅监听
		s.listen(closedC)

		// 开启区块监听
		s.startProcess(c.GetChainGenHash(), chainInfo.TableNum, closedC, blockC, gormDb)

		// 新增订阅列表
		s.chainList[c.GetChainGenHash()] = closedC
	}

	return nil
}

func (s *Subscriber) startProcess(genHash string, tableNum int,
	closedSignal chan<- string, c <-chan interface{}, gormDb *gorm.DB) {
	go func() {
		for {
			select {
			case block, ok := <-c:
				if !ok {
					closedSignal <- genHash
					s.log.Errorln("the chan of subscriber is closed...")
					return
				}

				blockInfo, ok := block.(*common.BlockInfo)
				if !ok {
					closedSignal <- genHash
					s.log.Errorln("the block info type error")
					return
				}

				// 协程池并发存储区块
				err := StorageBlock(blockInfo, genHash, tableNum, gormDb)
				if err != nil {
					closedSignal <- genHash
					s.log.Errorf("failed to storage block, err: [%s]\n", err.Error())
					return
				}

			case <-s.ctx.Done():
				return
			}
		}
	}()
}

func (s *Subscriber) listen(closedSignal <-chan string) {
	go func() {
		for {
			select {
			case genHash := <-closedSignal:
				// 将链状态更新为未订阅状态，走重新订阅逻辑。

				s.chainListMapMutex.Lock()
				delete(s.chainList, genHash)
				s.chainListMapMutex.Unlock()

				s.log.Warnf("the subscriber is closed, genHash: [%s]\n", genHash)
				return

			case <-s.ctx.Done():
				return
			}
		}
	}()
}
