package blockchain

import (
	"encoding/hex"
	"errors"
	"strings"

	cmsdk "chainmaker.org/chainmaker/sdk-go/v2"
	"go.uber.org/zap"
)

type NodeConnConfig struct {
	Addr        string
	ConnCount   int
	CaCertPem   []string
	TlsHostName string
	UseTls      bool
}

type ClientConfig struct {
	ChainId                                                     string
	OrgId                                                       string
	SignKeyBytes                                                []byte
	SignCertBytes                                               []byte
	HashAlgorithm                                               string
	NodeConfs                                                   []*NodeConnConfig
	Logger                                                      *zap.SugaredLogger
	TlsKeyBytes                                                 []byte
	TlsCertBytes                                                []byte
	RpcClientMaxReceiveMessageSize, RpcClientMaxSendMessageSize int
	ArchiveCenterUrl                                            string
}

type BlockChainClient struct {
	client       *cmsdk.ChainClient
	config       *ClientConfig
	chainGenHash string
}

func (c *BlockChainClient) GetConfig() *ClientConfig {
	return c.config
}

func (c *BlockChainClient) GetChainMakerClient() *cmsdk.ChainClient {
	return c.client
}

func (c *BlockChainClient) GetChainGenHash() string {
	return c.chainGenHash
}

func NewChainmakerClient(config *ClientConfig) (*BlockChainClient, error) {

	optionList := make([]cmsdk.ChainClientOption, 0)

	if len(config.ChainId) == 0 {
		return nil, errors.New("the chain id cannot be empty")
	}

	if len(config.SignKeyBytes) == 0 {
		return nil, errors.New("the sign key bytes cannot be empty")
	}

	if config.Logger == nil {
		return nil, errors.New("the logger cannot be nil")
	}

	if len(config.SignCertBytes) == 0 {
		err := checkTheHashAlgo(config.HashAlgorithm)
		if err != nil {
			return nil, err
		}

		optionList = append(optionList,
			cmsdk.WithCryptoConfig(
				cmsdk.NewCryptoConfig(cmsdk.WithHashAlgo(config.HashAlgorithm))),
			cmsdk.WithAuthType("public"))

	} else {
		optionList = append(optionList, cmsdk.WithAuthType("permissionedwithcert"))
	}

	optionList = append(optionList,
		cmsdk.WithChainClientChainId(config.ChainId),
		cmsdk.WithChainClientOrgId(config.OrgId),
		cmsdk.WithChainClientLogger(config.Logger),
		cmsdk.WithUserSignKeyBytes(config.SignKeyBytes),
		cmsdk.WithUserSignCrtBytes(config.SignCertBytes),
		cmsdk.WithUserKeyBytes(config.TlsKeyBytes),
		cmsdk.WithUserCrtBytes(config.TlsCertBytes),
	)

	nodeOptionList := make([]cmsdk.NodeOption, 0)

	for _, nodeConf := range config.NodeConfs {

		if nodeConf.UseTls {

			nodeOptionList = append(nodeOptionList, cmsdk.WithNodeUseTLS(true))

			if len(nodeConf.CaCertPem) == 0 {
				return nil, errors.New("the tls ca cert cannot be empty")
			}

			nodeOptionList = append(nodeOptionList, cmsdk.WithNodeCACerts(nodeConf.CaCertPem))

			if len(nodeConf.TlsHostName) == 0 {
				return nil, errors.New("the tls host name cannot be empty")
			}

			nodeOptionList = append(nodeOptionList, cmsdk.WithNodeTLSHostName(nodeConf.TlsHostName))

		}

		if len(nodeConf.Addr) == 0 {
			return nil, errors.New("the node address cannot be empty")
		}

		nodeOptionList = append(nodeOptionList, cmsdk.WithNodeAddr(nodeConf.Addr))

		if nodeConf.ConnCount == 0 {
			nodeOptionList = append(nodeOptionList, cmsdk.WithNodeConnCnt(10))
		} else {
			nodeOptionList = append(nodeOptionList, cmsdk.WithNodeConnCnt(nodeConf.ConnCount))
		}

	}

	optionList = append(optionList, cmsdk.AddChainClientNodeConfig(cmsdk.NewNodeConfig(nodeOptionList...)))

	rpcOptionList := make([]cmsdk.RPCClientOption, 0)

	if config.RpcClientMaxReceiveMessageSize == 0 {
		rpcOptionList = append(rpcOptionList, cmsdk.WithRPCClientMaxReceiveMessageSize(512))
	} else {
		rpcOptionList = append(rpcOptionList, cmsdk.WithRPCClientMaxReceiveMessageSize(config.RpcClientMaxReceiveMessageSize))
	}

	if config.RpcClientMaxSendMessageSize == 0 {
		rpcOptionList = append(rpcOptionList, cmsdk.WithRPCClientMaxReceiveMessageSize(512))
	} else {
		rpcOptionList = append(rpcOptionList, cmsdk.WithRPCClientMaxReceiveMessageSize(config.RpcClientMaxSendMessageSize))
	}

	optionList = append(optionList, cmsdk.WithRPCClientConfig(cmsdk.NewRPCClientConfig(rpcOptionList...)))

	client, err := cmsdk.NewChainClient(optionList...)
	if err != nil {
		return nil, err
	}

	block, err := client.GetBlockByHeight(0, false)
	if err != nil {
		return nil, err
	}

	chainGenHash := hex.EncodeToString(block.Block.Hash())

	if len(config.ArchiveCenterUrl) != 0 {

		optionList = append(optionList, cmsdk.WithArchiveCenterQueryFirst(true),
			cmsdk.WithArchiveConfig(&cmsdk.ArchiveConfig{}),
			cmsdk.WithArchiveCenterHttpConfig(&cmsdk.ArchiveCenterConfig{
				ChainGenesisHash:     chainGenHash,
				ArchiveCenterHttpUrl: config.ArchiveCenterUrl,
				ReqeustSecondLimit:   20,
			}))

		client.Stop()

		client, err = cmsdk.NewChainClient(optionList...)
		if err != nil {
			return nil, err
		}
	}

	return &BlockChainClient{
		client:       client,
		config:       config,
		chainGenHash: chainGenHash,
	}, nil
}

func checkTheHashAlgo(hashAlgo string) error {
	switch strings.ToLower(hashAlgo) {
	case strings.ToLower("SHA256"):
		return nil
	case strings.ToLower("SHA3_256"):
		return nil
	case strings.ToLower("SM3"):
		return nil
	}

	return errors.New("the hash algorithm is unknown")
}
