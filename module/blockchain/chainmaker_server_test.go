/*
Copyright (C) BABEC. All rights reserved.
Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/

package blockchain

import (
	"errors"
	"reflect"
	"sync"
	"testing"
	"time"

	"chainmaker.org/chainmaker/localconf/v2"
	"github.com/stretchr/testify/assert"

	"chainmaker.org/chainmaker-go/module/subscriber"
	commonPb "chainmaker.org/chainmaker/pb-go/v2/common"
	"chainmaker.org/chainmaker/protocol/v2"
	"chainmaker.org/chainmaker/protocol/v2/mock"

	"github.com/golang/mock/gomock"
)

func TestInitNet(t *testing.T) {
	t.Log("TestInitNet")
	chainmakerServerList := makeChainServer(t)

	for _, server := range chainmakerServerList {
		err := server.initNet()

		if err != nil {
			t.Log(err)
		}
	}

}

func TestInitBlockchains(t *testing.T) {
	t.Log("TestInitBlockchains")

	chainmakerServerList := makeChainServer(t)

	for _, server := range chainmakerServerList {
		err := server.initBlockchains()

		if err != nil {
			t.Log(err)
		}
	}
}

func TestDeleteBlockchainTaskListener(t *testing.T) {
	t.Log("TestDeleteBlockchainTaskListener")

	defer func() {
		err := recover()
		assert.Nil(t, err)
	}()

	chainmakerServerList := makeChainServer(t)
	chainmakerServerList[0].blockchains.Store("chain1", &Blockchain{
		chainId: "chain1",
	})
	chainmakerServerList[0].blockchains.Store("chain1", &Blockchain{
		chainId: "chain2",
	})
	go chainmakerServerList[0].deleteBlockchainTaskListener()

	localconf.FindDeleteBlockChainNotifyC <- "chain1"
	time.Sleep(time.Second)
}

func TestNewBlockchainTaskListener(t *testing.T) {
	t.Log("TestDeleteBlockchainTaskListener")

	defer func() {
		err := recover()
		assert.Nil(t, err)
	}()

	chainmakerServerList := makeChainServer(t)
	chainmakerServerList[0].blockchains.Store("chain1", &Blockchain{
		chainId: "chain1",
	})
	go chainmakerServerList[0].newBlockchainTaskListener()

	localconf.FindNewBlockChainNotifyC <- "chain2"
	time.Sleep(time.Second)
}

//func TestStart(t *testing.T) {
//	t.Log("TestStart")
//
//	chainmakerServerList := makeChainServer(t)
//
//	for _, server := range chainmakerServerList {
//		err := server.Start()
//
//		if err != nil {
//			t.Log(err)
//		}
//	}
//}

//func TestStop(t *testing.T) {
//	t.Log("TestStop")
//
//	chainmakerServerList := makeChainServer(t)
//
//	for _, server := range chainmakerServerList {
//		err := server.Start()
//
//		if err != nil {
//			t.Log(err)
//		}
//
//		server.Stop()
//	}
//}

//func TestAddTx(t *testing.T) {
//	t.Log("TestAddTx")
//
//	chainmakerServerList := makeChainServer(t)
//
//	for _, server := range chainmakerServerList {
//		err := server.Start()
//
//		if err != nil {
//			t.Log(err)
//		}
//
//		server.AddTx("123456", &common.Transaction{
//			//Payload:   "Payload",
//			//Sender:    req.Sender,
//			//Endorsers: req.Endorsers,
//			Result: nil}, protocol.RPC)
//	}
//}

//func TestGetStore(t *testing.T) {
//	t.Log("TestGetStore")
//
//	chainmakerServerList := makeChainServer(t)
//
//	for _, server := range chainmakerServerList {
//		err := server.Start()
//
//		if err != nil {
//			t.Log(err)
//		}
//
//		bkStore, err := server.GetStore("123456")
//
//		if err != nil {
//			t.Log(err)
//		}
//
//		t.Log(bkStore)
//	}
//
//}

//func TestGetChainConf(t *testing.T) {
//	t.Log("TestGetChainConf")
//
//	chainmakerServerList := makeChainServer(t)
//
//	for _, server := range chainmakerServerList {
//		err := server.Start()
//
//		if err != nil {
//			t.Log(err)
//		}
//
//		chainConf, err := server.GetChainConf("123456")
//
//		if err != nil {
//			t.Log(err)
//		}
//
//		t.Log(chainConf)
//	}
//
//}

//func TestGetAllChainConf(t *testing.T) {
//	t.Log("TestGetAllChainConf")
//
//	chainmakerServerList := makeChainServer(t)
//
//	for _, server := range chainmakerServerList {
//		err := server.Start()
//
//		if err != nil {
//			t.Log(err)
//		}
//
//		chainAllConf, err := server.GetAllChainConf()
//
//		if err != nil {
//			t.Log(err)
//		}
//
//		t.Log(chainAllConf)
//	}
//
//}

//func TestGetVmManager(t *testing.T) {
//	t.Log("TestGetVmManager")
//
//	chainmakerServerList := makeChainServer(t)
//
//	for _, server := range chainmakerServerList {
//		err := server.Start()
//
//		if err != nil {
//			t.Log(err)
//		}
//
//		vmMgr, err := server.GetVmManager("123456")
//
//		if err != nil {
//			t.Log(err)
//		}
//
//		t.Log(vmMgr)
//	}
//
//}

//func TestGetEventSubscribe(t *testing.T) {
//	t.Log("TestGetEventSubscribe")
//
//	chainmakerServerList := makeChainServer(t)
//
//	for _, server := range chainmakerServerList {
//		err := server.Start()
//
//		if err != nil {
//			t.Log(err)
//		}
//
//		sub, err := server.GetEventSubscribe("123456")
//
//		if err != nil {
//			t.Log(err)
//		}
//
//		t.Log(sub)
//	}
//
//}

//func TestGetNetService(t *testing.T) {
//	t.Log("TestGetNetService")
//
//	chainmakerServerList := makeChainServer(t)
//
//	for _, server := range chainmakerServerList {
//		err := server.Start()
//
//		if err != nil {
//			t.Log(err)
//		}
//
//		netService, err := server.GetNetService("123456")
//
//		if err != nil {
//			t.Log(err)
//		}
//
//		t.Log(netService)
//	}
//}

//

/*func TestInitBlockchain(t *testing.T)  {
	t.Log("TestInitBlockchain")

	chainmakerServer := ChainMakerServer{}
	err := chainmakerServer.initBlockchain("12344", "666")

	if err != nil {
		t.Log(err)
	}
}*/

// TODO
/*func TestNewBlockchainTaskListener(t *testing.T)  {
	t.Log("TestInitBlockchains")

	chainmakerServer := ChainMakerServer{}
	chainmakerServer.newBlockchainTaskListener()
	timer := time.NewTimer(3 * time.Second)
	<-timer.C
}*/

//func TestChainMakerServer_initBlockchain(t *testing.T) {
//	type fields struct {
//		net         protocol.Net
//		blockchains sync.Map
//		readyC      chan struct{}
//	}
//	type args struct {
//		chainId string
//		genesis string
//	}
//	tests := []struct {
//		name    string
//		fields  fields
//		args    args
//		wantErr bool
//	}{
//		{
//			name: "test0",
//			fields: fields{
//				net:         nil,
//				blockchains: sync.Map{},
//				readyC:      nil,
//			},
//			args: args{
//				chainId: "chain1",
//				genesis: "chain1",
//			},
//			wantErr: false,
//		},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			server := &ChainMakerServer{
//				net:         tt.fields.net,
//				blockchains: tt.fields.blockchains,
//				readyC:      tt.fields.readyC,
//			}
//
//			server.blockchains.Store("chain1", &Blockchain{
//				chainId: "chain1",
//				net:     newMockNet(t),
//				store: func() protocol.BlockchainStore {
//					store := newMockBlockchainStore(t)
//					store.EXPECT().GetLastChainConfig().AnyTimes()
//					return store
//				}(),
//				initModules: map[string]struct{}{
//					moduleNameSubscriber: {},
//					//moduleNameStore:         {},
//					moduleNameOldStore:      {},
//					moduleNameLedger:        {},
//					moduleNameChainConf:     {},
//					moduleNameAccessControl: {},
//					moduleNameNetService:    {},
//					moduleNameVM:            {},
//					moduleNameTxPool:        {},
//					moduleNameCore:          {},
//					moduleNameConsensus:     {},
//					moduleNameSync:          {},
//				},
//			})
//
//			if err := server.initBlockchain(tt.args.chainId, tt.args.genesis); (err != nil) != tt.wantErr {
//				t.Errorf("initBlockchain() error = %v, wantErr %v", err, tt.wantErr)
//			}
//		})
//	}
//}

func TestChainMakerServer_StartForRebuildDbs(t *testing.T) {
	type fields struct {
		net    protocol.Net
		readyC chan struct{}
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "test0",
			fields: fields{
				net:    nil,
				readyC: make(chan struct{}),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := &ChainMakerServer{
				net:    tt.fields.net,
				readyC: tt.fields.readyC,
			}
			if err := server.StartForRebuildDbs(false); (err != nil) != tt.wantErr {
				t.Errorf("StartForRebuildDbs() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestChainMakerServer_Start(t *testing.T) {
	type fields struct {
		net    protocol.Net
		readyC chan struct{}
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "test0",
			fields: fields{
				net: func() protocol.Net {
					net := newMockNet(t)
					net.EXPECT().Start().AnyTimes()
					return net
				}(),
				readyC: make(chan struct{}),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := &ChainMakerServer{
				net:    tt.fields.net,
				readyC: tt.fields.readyC,
			}
			if err := server.Start(); (err != nil) != tt.wantErr {
				t.Errorf("Start() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestChainMakerServer_Stop(t *testing.T) {
	type fields struct {
		net    protocol.Net
		readyC chan struct{}
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name: "test0",
			fields: fields{
				net: func() protocol.Net {
					net := newMockNet(t)
					net.EXPECT().Stop().AnyTimes()
					return net
				}(),
				readyC: nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := &ChainMakerServer{
				net:    tt.fields.net,
				readyC: tt.fields.readyC,
			}
			server.Stop()
		})
	}
}

func TestChainMakerServer_AddTx(t *testing.T) {
	type fields struct {
		net    protocol.Net
		readyC chan struct{}
	}
	type args struct {
		chainId string
		tx      *commonPb.Transaction
		source  protocol.TxSource
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "test0",
			fields: fields{
				net:    nil,
				readyC: nil,
			},
			args:    args{},
			wantErr: true,
		},
		{
			name: "test1",
			fields: fields{
				net:    nil,
				readyC: nil,
			},
			args: args{
				chainId: "chain1",
			},
			wantErr: false,
		},
	}
	for k, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := &ChainMakerServer{
				net:    tt.fields.net,
				readyC: tt.fields.readyC,
			}
			if k == 1 {
				server.blockchains.Store("chain1", &Blockchain{
					chainId: "chain1",
					txPool: func() protocol.TxPool {
						txPool := newMockTxPool(t)
						txPool.EXPECT().AddTx(gomock.Any(), gomock.Any()).AnyTimes()
						return txPool
					}(),
				})
			}
			if err := server.AddTx(tt.args.chainId, tt.args.tx, tt.args.source); (err != nil) != tt.wantErr {
				t.Errorf("AddTx() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestChainMakerServer_GetStore(t *testing.T) {
	type fields struct {
		net    protocol.Net
		readyC chan struct{}
	}
	type args struct {
		chainId string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    protocol.BlockchainStore
		wantErr bool
	}{
		{
			name: "test0",
			fields: fields{
				net:    nil,
				readyC: nil,
			},
			args:    args{},
			want:    nil,
			wantErr: true,
		},
		{
			name: "test1",
			fields: fields{
				net:    nil,
				readyC: nil,
			},
			args: args{
				chainId: "chain1",
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "test2",
			fields: fields{
				net:    nil,
				readyC: nil,
			},
			args: args{
				chainId: "chain1",
			},
			want:    nil,
			wantErr: false,
		},
	}
	for k, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := &ChainMakerServer{
				net:    tt.fields.net,
				readyC: tt.fields.readyC,
			}
			if k == 1 {
				server.blockchains.Store("chain1", &Blockchain{
					chainId: "chain1",
				})
			}
			if k == 2 {
				server.blockchains.Store("chain1", &Blockchain{})
			}
			got, err := server.GetStore(tt.args.chainId)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetStore() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetStore() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestChainMakerServer_GetChainConf(t *testing.T) {
	type fields struct {
		net    protocol.Net
		readyC chan struct{}
	}
	type args struct {
		chainId string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    protocol.ChainConf
		wantErr bool
	}{
		{
			name: "test0",
			fields: fields{
				net:    nil,
				readyC: nil,
			},
			args:    args{},
			want:    nil,
			wantErr: true,
		},
		{
			name: "test1",
			fields: fields{
				net:    nil,
				readyC: nil,
			},
			args: args{
				chainId: "chain1",
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "test2",
			fields: fields{
				net:    nil,
				readyC: nil,
			},
			args: args{
				chainId: "chain1",
			},
			want:    nil,
			wantErr: false,
		},
	}
	for k, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := &ChainMakerServer{
				net:    tt.fields.net,
				readyC: tt.fields.readyC,
			}

			if k == 1 {
				server.blockchains.Store("chain1", &Blockchain{
					chainId: "chain1",
				})
			}
			if k == 2 {
				server.blockchains.Store("chain1", &Blockchain{})
			}
			got, err := server.GetChainConf(tt.args.chainId)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetChainConf() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetChainConf() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestChainMakerServer_GetAllChainConf(t *testing.T) {
	type fields struct {
		net    protocol.Net
		readyC chan struct{}
	}

	chainConf := newMockChainConf(t)
	tests := []struct {
		name    string
		fields  fields
		want    []protocol.ChainConf
		wantErr bool
	}{
		{
			name: "test0",
			fields: fields{
				net:    nil,
				readyC: nil,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "test1",
			fields: fields{
				net:    nil,
				readyC: nil,
			},
			want:    []protocol.ChainConf{nil},
			wantErr: false,
		},
		{
			name: "test2",
			fields: fields{
				net:    nil,
				readyC: nil,
			},
			want:    []protocol.ChainConf{chainConf},
			wantErr: false,
		},
	}
	for k, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := &ChainMakerServer{
				net:    tt.fields.net,
				readyC: tt.fields.readyC,
			}

			if k == 1 {
				server.blockchains.Store("chain1", &Blockchain{
					chainId:   "chain1",
					chainConf: nil,
				})
			}
			if k == 2 {
				server.blockchains.Store("chain1", &Blockchain{
					chainId:   "chain1",
					chainConf: chainConf,
				})
			}

			got, err := server.GetAllChainConf()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAllChainConf() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetAllChainConf() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestChainMakerServer_GetVmManager(t *testing.T) {
	type fields struct {
		net    protocol.Net
		readyC chan struct{}
	}
	type args struct {
		chainId string
	}

	vmMgr := newMockVmManager(t)
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    protocol.VmManager
		wantErr bool
	}{
		{
			name: "test0",
			fields: fields{
				net:    nil,
				readyC: nil,
			},
			args: args{
				chainId: "chain1",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "test1",
			fields: fields{
				net:    nil,
				readyC: nil,
			},
			args: args{
				chainId: "chain1",
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "test2",
			fields: fields{
				net:    nil,
				readyC: nil,
			},
			args: args{
				chainId: "chain1",
			},
			want:    vmMgr,
			wantErr: false,
		},
	}
	for k, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := &ChainMakerServer{
				net:    tt.fields.net,
				readyC: tt.fields.readyC,
			}

			if k == 1 {
				server.blockchains.Store("chain1", &Blockchain{
					chainId: "chain1",
				})
			}
			if k == 2 {
				server.blockchains.Store("chain1", &Blockchain{
					chainId: "chain1",
					vmMgr:   vmMgr,
				})
			}
			got, err := server.GetVmManager(tt.args.chainId)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetVmManager() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetVmManager() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestChainMakerServer_GetEventSubscribe(t *testing.T) {
	type fields struct {
		net    protocol.Net
		readyC chan struct{}
	}
	type args struct {
		chainId string
	}

	msgBus := newMockMessageBus(t)
	msgBus.EXPECT().Register(gomock.Any(), gomock.Any()).AnyTimes()
	eventSubscriber := subscriber.NewSubscriber(msgBus)
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *subscriber.EventSubscriber
		wantErr bool
	}{
		{
			name: "test0",
			fields: fields{
				net:    nil,
				readyC: nil,
			},
			args:    args{},
			want:    nil,
			wantErr: true,
		},
		{
			name: "test1",
			fields: fields{
				net:    nil,
				readyC: nil,
			},
			args: args{
				chainId: "chain1",
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "test2",
			fields: fields{
				net:    nil,
				readyC: nil,
			},
			args: args{
				chainId: "chain1",
			},
			want:    eventSubscriber,
			wantErr: false,
		},
	}
	for k, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := &ChainMakerServer{
				net:    tt.fields.net,
				readyC: tt.fields.readyC,
			}

			if k == 1 {
				server.blockchains.Store("chain1", &Blockchain{
					chainId: "chain1",
				})
			}
			if k == 2 {
				server.blockchains.Store("chain1", &Blockchain{
					chainId:         "chain1",
					eventSubscriber: eventSubscriber,
				})
			}
			got, err := server.GetEventSubscribe(tt.args.chainId)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetEventSubscribe() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetEventSubscribe() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestChainMakerServer_GetNetService(t *testing.T) {
	type fields struct {
		net    protocol.Net
		readyC chan struct{}
	}
	type args struct {
		chainId string
	}

	netService := newMockNetService(t)
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    protocol.NetService
		wantErr bool
	}{
		{
			name: "test0",
			fields: fields{
				net:    newMockNet(t),
				readyC: nil,
			},
			args:    args{},
			want:    nil,
			wantErr: true,
		},
		{
			name: "test1",
			fields: fields{
				net:    newMockNet(t),
				readyC: nil,
			},
			args: args{
				chainId: "chain1",
			},
			want:    netService,
			wantErr: false,
		},
		{
			name: "test2",
			fields: fields{
				net:    newMockNet(t),
				readyC: nil,
			},
			args: args{
				chainId: "chain1",
			},
			want:    nil,
			wantErr: false,
		},
	}
	for k, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := &ChainMakerServer{
				net:    tt.fields.net,
				readyC: tt.fields.readyC,
			}

			if k == 1 {
				server.blockchains.Store("chain1", &Blockchain{
					netService: netService,
				})
			}
			if k == 2 {
				server.blockchains.Store("chain1", &Blockchain{})
			}
			got, err := server.GetNetService(tt.args.chainId)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetNetService() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetNetService() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestChainMakerServer_GetBlockchain(t *testing.T) {
	type fields struct {
		net    protocol.Net
		readyC chan struct{}
	}
	type args struct {
		chainId string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *Blockchain
		wantErr bool
	}{
		{
			name: "test0",
			fields: fields{
				net:    newMockNet(t),
				readyC: nil,
			},
			args:    args{},
			want:    nil,
			wantErr: true,
		},
		{
			name: "test1",
			fields: fields{
				net:    newMockNet(t),
				readyC: nil,
			},
			args: args{
				chainId: "chain1",
			},
			want:    &Blockchain{},
			wantErr: false,
		},
	}

	for k, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := &ChainMakerServer{
				net:    tt.fields.net,
				readyC: tt.fields.readyC,
			}

			if k == 1 {
				server.blockchains.Store("chain1", &Blockchain{})
			}

			got, err := server.GetBlockchain(tt.args.chainId)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetBlockchain() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetBlockchain() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestChainMakerServer_GetAllAC(t *testing.T) {
	type fields struct {
		net    protocol.Net
		readyC chan struct{}
	}
	tests := []struct {
		name    string
		fields  fields
		want    []protocol.AccessControlProvider
		wantErr bool
	}{
		{
			name: "test0",
			fields: fields{
				net:    nil,
				readyC: nil,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "test1",
			fields: fields{
				net:    newMockNet(t),
				readyC: nil,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "test2",
			fields: fields{
				net:    newMockNet(t),
				readyC: nil,
			},
			want: []protocol.AccessControlProvider{
				nil,
			},
			wantErr: false,
		},
	}
	for k, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := &ChainMakerServer{
				net:    tt.fields.net,
				readyC: tt.fields.readyC,
			}

			if k == 2 {
				server.blockchains.Store("chain1", &Blockchain{})
			}

			got, err := server.GetAllAC()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAllAC() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetAllAC() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestChainMakerServer_Version(t *testing.T) {
	type fields struct{}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name:   "test0",
			fields: fields{},
			want:   CurrentVersion,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := &ChainMakerServer{}
			if got := server.Version(); got != tt.want {
				t.Errorf("Version() = %v, want %v", got, tt.want)
			}
		})
	}
}

func makeChainServer(t *testing.T) []*ChainMakerServer {

	serverList := make([]*ChainMakerServer, 0)
	for i := 0; i < 4; i++ {
		var (
			ctrl = gomock.NewController(t)
			net  = mock.NewMockNet(ctrl)
			//blockChain = mock.NewMockSyncService(ctrl)
		)

		if i == 0 {
			net.EXPECT().Start().AnyTimes().Return(errors.New("server start test err msg"))
			net.EXPECT().Stop().AnyTimes().Return(nil)
		} else if i == 1 {
			net.EXPECT().Start().AnyTimes().Return(errors.New("server start test err msg"))
			net.EXPECT().Stop().AnyTimes().Return(errors.New("server stop test err msg"))
		} else if i == 2 {
			net.EXPECT().Start().AnyTimes().Return(nil)
			net.EXPECT().Stop().AnyTimes().Return(errors.New("server stop test err msg"))
		} else {
			net.EXPECT().Start().AnyTimes().Return(nil)
			net.EXPECT().Stop().AnyTimes().Return(nil)
		}

		//net.EXPECT().Start().AnyTimes().Return(nil)
		//net.EXPECT().Stop().AnyTimes().Return(nil)
		chainMaker := &ChainMakerServer{
			net:         net,
			blockchains: sync.Map{},
			//readyC: struct {}<-,
		}

		err := chainMaker.Init()

		if err != nil {
			t.Log(err)
		}

		serverList = append(serverList, chainMaker)
	}

	return serverList
}
