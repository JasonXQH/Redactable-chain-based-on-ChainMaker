package common

import (
	//"chainmaker.org/chainmaker-go/module/common/modify"
	msgbusMock "chainmaker.org/chainmaker/common/v2/msgbus/mock"
	commonpb "chainmaker.org/chainmaker/pb-go/v2/common"
	"chainmaker.org/chainmaker/protocol/v2"
	"chainmaker.org/chainmaker/protocol/v2/mock"
	"fmt"
	"github.com/golang/mock/gomock"
	"testing"
)

func TestCommitBlock_ReplaceBlock(t *testing.T) {
	var (
		blockchainStore = newMockBlockchainStore(t)
		snapshotManager = newMockSnapshotManager(t)
		ledgerCache     = newMockLedgerCache(t)
		chainConf       = newMockChainConf(t)
		//msgBus          = msgbus.NewMessageBus()
		//storeHelper     = newMockStoreHelper(t)
		txFilter = newMockTxFilter(t)
		log      = newMockLogger(t)
	)
	replaceBlockInstance := &ReplaceBlock{
		store:           blockchainStore,
		log:             log,
		snapshotManager: snapshotManager,
		ledgerCache:     ledgerCache,
		chainConf:       chainConf,
		txFilter:        txFilter,
	}
	block, _ := modify.ModifyBlockByHeight(2)
	fmt.Println("new Block :", block.String())
	txRWSetMap := make(map[string]*commonpb.TxRWSet)
	//rwSet := utils.RearrangeRWSet(block, txRWSetMap)

	_, _, _, _, _, _, _, _ = replaceBlockInstance.ReplaceBlock(block, txRWSetMap)
}

//func newMockChainConf(t *testing.T) *mock.MockChainConf {
//	ctrl := gomock.NewController(t)
//	chainConf := mock.NewMockChainConf(ctrl)
//	return chainConf
//}

func newMockBlockchainStore(t *testing.T) *mock.MockBlockchainStore {
	ctrl := gomock.NewController(t)
	blockchainStore := mock.NewMockBlockchainStore(ctrl)
	return blockchainStore
}

func newMockStoreHelper(t *testing.T) *mock.MockStoreHelper {
	ctrl := gomock.NewController(t)
	storeHelper := mock.NewMockStoreHelper(ctrl)
	return storeHelper
}

func newMockLogger(t *testing.T) *mock.MockLogger {
	ctrl := gomock.NewController(t)
	logger := mock.NewMockLogger(ctrl)
	logger.EXPECT().Debugf(gomock.Any(), gomock.Any()).AnyTimes()
	logger.EXPECT().Infof(gomock.Any(), gomock.Any()).AnyTimes()
	logger.EXPECT().Error(gomock.Any()).AnyTimes()

	return logger
}

func newMockVmManager(t *testing.T) *mock.MockVmManager {
	ctrl := gomock.NewController(t)
	vmManager := mock.NewMockVmManager(ctrl)
	vmManager.EXPECT().RunContract(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(&commonpb.ContractResult{
		Code: 0,
	}, protocol.ExecOrderTxTypeNormal, commonpb.TxStatusCode_SUCCESS).AnyTimes()
	return vmManager
}

func newMockTxPool(t *testing.T) *mock.MockTxPool {
	ctrl := gomock.NewController(t)
	txPool := mock.NewMockTxPool(ctrl)
	return txPool
}

func newMockSnapshotManager(t *testing.T) *mock.MockSnapshotManager {
	ctrl := gomock.NewController(t)
	snapshotManager := mock.NewMockSnapshotManager(ctrl)
	return snapshotManager
}

func newMockLedgerCache(t *testing.T) *mock.MockLedgerCache {
	ctrl := gomock.NewController(t)
	newMockLedgerCache := mock.NewMockLedgerCache(ctrl)
	return newMockLedgerCache
}

//func newMockProposalCache(t *testing.T) *mock.MockProposalCache {
//	ctrl := gomock.NewController(t)
//	proposalCache := mock.NewMockProposalCache(ctrl)
//	return proposalCache
//}

func newMockBlockVerifier(t *testing.T) *mock.MockBlockVerifier {
	ctrl := gomock.NewController(t)
	blockVerifier := mock.NewMockBlockVerifier(ctrl)
	return blockVerifier
}

func newMockBlockCommitter(t *testing.T) *mock.MockBlockCommitter {
	ctrl := gomock.NewController(t)
	blockCommitter := mock.NewMockBlockCommitter(ctrl)
	return blockCommitter
}

func newMockSigningMember(t *testing.T) *mock.MockSigningMember {
	ctrl := gomock.NewController(t)
	signingMember := mock.NewMockSigningMember(ctrl)
	return signingMember
}

func newMockAccessControlProvider(t *testing.T) *mock.MockAccessControlProvider {
	ctrl := gomock.NewController(t)
	ac := mock.NewMockAccessControlProvider(ctrl)
	return ac
}

func newMockTxScheduler(t *testing.T) *mock.MockTxScheduler {
	ctrl := gomock.NewController(t)
	txScheduler := mock.NewMockTxScheduler(ctrl)
	return txScheduler
}

func newMockMessageBus(t *testing.T) *msgbusMock.MockMessageBus {
	ctrl := gomock.NewController(t)
	messageBus := msgbusMock.NewMockMessageBus(ctrl)
	return messageBus
}

func newMockBlockProposer(t *testing.T) *mock.MockBlockProposer {
	ctrl := gomock.NewController(t)
	blockProposer := mock.NewMockBlockProposer(ctrl)
	return blockProposer
}
func newMockTxFilter(t *testing.T) *mock.MockTxFilter {
	ctrl := gomock.NewController(t)
	TxFilter := mock.NewMockTxFilter(ctrl)
	return TxFilter
}
