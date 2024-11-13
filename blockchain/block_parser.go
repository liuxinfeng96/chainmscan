package blockchain

import (
	"chainmscan/db/dao"
	dbModel "chainmscan/db/model"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"

	"chainmaker.org/chainmaker/pb-go/v2/common"
	"chainmaker.org/chainmaker/pb-go/v2/syscontract"
	"github.com/gogo/protobuf/proto"
	"gorm.io/gorm"
)

type BlockData struct {
	Block              *dbModel.Block
	BlockDetails       *dbModel.BlockDetails
	Transactions       []*dbModel.Transaction
	TransactionDetails []*dbModel.TxDetails
	Contracts          []*dbModel.Contract
}

func ParseBlock(blockInfo *common.BlockInfo) (*BlockData, error) {

	blockData := new(BlockData)

	blockHeader := new(common.BlockHeader)
	blockHeader = blockInfo.Block.Header

	dbBlock := &dbModel.Block{
		BlockHeight:    blockHeader.BlockHeight,
		BlockHash:      hex.EncodeToString(blockHeader.BlockHash),
		ChainId:        blockHeader.ChainId,
		PreBlockHash:   hex.EncodeToString(blockHeader.PreBlockHash),
		BlockType:      blockHeader.BlockType.String(),
		BlockVersion:   blockHeader.BlockVersion,
		PreConfHeight:  blockHeader.PreConfHeight,
		TxCount:        blockHeader.TxCount,
		TxRoot:         hex.EncodeToString(blockHeader.TxRoot),
		DagHash:        hex.EncodeToString(blockHeader.DagHash),
		RwSetRoot:      hex.EncodeToString(blockHeader.RwSetRoot),
		BlockTimestamp: blockHeader.BlockTimestamp,
		ConsensusArgs:  string(blockHeader.ConsensusArgs),
	}

	dbBlockDetails := &dbModel.BlockDetails{
		BlockHash: dbBlock.BlockHash,
		Dag:       blockInfo.Block.Dag.String(),
	}

	if blockHeader.Proposer != nil {
		proposerBytes, err := blockHeader.Proposer.Marshal()
		if err != nil {
			return nil, err
		}

		dbBlock.ProposerOrgId = blockHeader.Proposer.OrgId
		dbBlockDetails.ProposerBytes = proposerBytes
		dbBlockDetails.ProposerSignature = base64.StdEncoding.EncodeToString(blockHeader.Signature)
	}

	blockData.Block = dbBlock
	blockData.BlockDetails = dbBlockDetails

	transactions := make([]*dbModel.Transaction, 0)
	transactionDetails := make([]*dbModel.TxDetails, 0)
	contracts := make([]*dbModel.Contract, 0)

	for _, t := range blockInfo.Block.Txs {
		tx := &dbModel.Transaction{
			ChainId:     blockHeader.ChainId,
			BlockHeight: blockHeader.BlockHeight,
		}

		txDetails := &dbModel.TxDetails{}

		if t.Payload != nil {
			tx.TxId = t.Payload.TxId
			tx.ContractName = t.Payload.ContractName
			tx.Method = t.Payload.Method
			tx.TxType = t.Payload.TxType.String()
			tx.Timestamp = t.Payload.Timestamp
			tx.ExpirationTime = t.Payload.ExpirationTime
			tx.Sequence = t.Payload.Sequence

			txDetails.TxId = t.Payload.TxId
		}

		if t.Payload.Limit != nil {
			tx.GasLimit = t.Payload.Limit.GasLimit
		}

		if t.Result != nil {
			tx.TxStatusCode = t.Result.Code.String()
			txDetails.TxStatusCode = t.Result.Code.String()
			txDetails.RwSetHash = hex.EncodeToString(t.Result.RwSetHash)
			txDetails.TxMessage = t.Result.Message

			if t.Result.ContractResult != nil {
				txDetails.ContractResultCode = t.Result.ContractResult.Code
				txDetails.ContractResult = t.Result.ContractResult.Result
				txDetails.ContractResultMessage = t.Result.ContractResult.Message
				txDetails.GasUsed = t.Result.ContractResult.GasUsed

				if len(t.Result.ContractResult.ContractEvent) != 0 {
					eventJson, err := json.Marshal(t.Result.ContractResult.ContractEvent)
					if err != nil {
						return nil, err
					}
					txDetails.ContractEventBytes = eventJson
				}

				if t.Payload.ContractName == syscontract.SystemContract_CONTRACT_MANAGE.String() {
					pbContract := new(common.Contract)
					err := proto.Unmarshal(t.Result.ContractResult.Result, pbContract)
					if err != nil {
						err := json.Unmarshal(t.Result.ContractResult.Result, pbContract)
						if err != nil {
							return nil, err
						}
					}

					if t.Result.Code == common.TxStatusCode_SUCCESS &&
						t.Result.ContractResult.Code == 0 {
						c := &dbModel.Contract{
							Name:         pbContract.Name,
							Version:      pbContract.Version,
							ChainId:      blockHeader.ChainId,
							State:        pbContract.Status.String(),
							RuntimeType:  pbContract.RuntimeType.String(),
							CreatorOrgId: pbContract.Creator.OrgId,
							Address:      pbContract.Address,
							TxId:         t.Payload.TxId,
							Height:       blockHeader.BlockHeight,
							TxTimestamp:  t.Payload.Timestamp,
						}
						creatorBytes, err := pbContract.Creator.Marshal()
						if err != nil {
							return nil, err
						}

						c.CreatorBytes = creatorBytes

						contracts = append(contracts, c)
					}
				}
			}
		}

		if len(t.Payload.Parameters) != 0 {
			paramsJson, err := json.Marshal(t.Payload.Parameters)
			if err != nil {
				return nil, err
			}
			txDetails.TxParameters = paramsJson
		}

		if t.Sender != nil && t.Sender.Signer != nil {
			senderBytes, err := t.Sender.Marshal()
			if err != nil {
				return nil, err
			}
			tx.SenderOrgId = t.Sender.Signer.OrgId
			txDetails.SenderBytes = senderBytes
		}

		if len(t.Endorsers) != 0 {
			endorserJson, err := json.Marshal(t.Endorsers)
			if err != nil {
				return nil, err
			}
			txDetails.EndorsersBytes = endorserJson
		}

		transactions = append(transactions, tx)
		transactionDetails = append(transactionDetails, txDetails)
	}

	blockData.Transactions = transactions
	blockData.TransactionDetails = transactionDetails
	blockData.Contracts = contracts

	return blockData, nil
}

func StorageBlock(blockInfo *common.BlockInfo, genHash string, tableNum int,
	gormDb *gorm.DB) error {
	blockData, err := ParseBlock(blockInfo)
	if err != nil {
		return errors.New("fail to parse block, " + err.Error())
	}

	err = gormDb.Transaction(func(tx *gorm.DB) error {
		// 查询链统计信息
		chainInfo, err := dao.GetChainInfo(genHash, tx)
		if err != nil {
			return err
		}

		if chainInfo == nil {
			chainInfo = &dbModel.ChainInfo{}
		}

		// 区块为单位解析，区块+1
		blockAmount := chainInfo.BlockAmount + 1
		txAmount := chainInfo.TxAmount + int(blockData.Block.TxCount)

		err = dao.InsertOneObjectToDBByTableName(blockData.Block,
			fmt.Sprintf(dbModel.TableNamePrefix_Block+"_%02d", tableNum), tx)
		if err != nil {
			return err
		}

		err = dao.InsertOneObjectToDBByTableName(blockData.BlockDetails,
			fmt.Sprintf(dbModel.TableNamePrefix_BlockDetails+"_%02d", tableNum), tx)
		if err != nil {
			return err
		}

		for _, t := range blockData.Transactions {
			err = dao.InsertOneObjectToDBByTableName(t,
				fmt.Sprintf(dbModel.TableNamePrefix_Transaction+"_%02d", tableNum), tx)
			if err != nil {
				return err
			}
		}

		for _, td := range blockData.TransactionDetails {
			err = dao.InsertOneObjectToDBByTableName(td,
				fmt.Sprintf(dbModel.TableNamePrefix_TxDetails+"_%02d", tableNum), tx)
			if err != nil {
				return err
			}
		}

		for _, c := range blockData.Contracts {
			err = dao.InsertOneObjectToDBByTableName(c,
				fmt.Sprintf(dbModel.TableNamePrefix_Contract+"_%02d", tableNum), tx)
			if err != nil {
				return err
			}
		}

		return dao.UpdateChainTxAndBlockAmount(genHash, txAmount, blockAmount, tx)
	})
	if err != nil {
		return errors.New("fail to insert block to db, " + err.Error())
	}

	return nil
}
