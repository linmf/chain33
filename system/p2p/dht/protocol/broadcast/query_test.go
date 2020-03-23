// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package broadcast

import (
	"encoding/hex"
	"testing"

	"github.com/33cn/chain33/common/merkle"
	"github.com/33cn/chain33/queue"
	"github.com/33cn/chain33/types"
	"github.com/stretchr/testify/assert"
)

func Test_sendQueryData(t *testing.T) {

	proto := newTestProtocol()
	_, ok := proto.handleSend(&types.P2PQueryData{}, testPid, testAddr)
	assert.True(t, ok)
}

func Test_sendQueryReply(t *testing.T) {

	proto := newTestProtocol()
	_, ok := proto.handleSend(&types.P2PBlockTxReply{}, testPid, testAddr)
	assert.True(t, ok)
}

func Test_recvQueryData(t *testing.T) {

	q := queue.New("test")
	go q.Start()
	defer q.Close()
	proto := newTestProtocolWithQueue(q)
	query := &types.P2PQueryData{
		Value: &types.P2PQueryData_TxReq{
			TxReq: &types.P2PTxReq{TxHash: tx.Hash()}}}
	sendData, _ := proto.handleSend(query, testPid, testAddr)
	memTxs := []*types.Transaction{nil}
	go testHandleMempool(q, &memTxs)
	err := proto.handleReceive(sendData, testPid, testAddr)
	assert.Equal(t, errRecvMempool, err)
	memTxs = []*types.Transaction{tx}
	err = proto.handleReceive(sendData, testPid, testAddr)
	assert.Equal(t, errSendStream, err)

	blockHash := hex.EncodeToString(testBlock.Hash(proto.ChainCfg))
	proto.totalBlockCache.Add(blockHash, testBlock, int64(testBlock.Size()))
	req := &types.P2PBlockTxReq{
		BlockHash: blockHash,
		TxIndices: []int32{0, 1, 2},
	}
	query = &types.P2PQueryData{
		Value: &types.P2PQueryData_BlockTxReq{
			BlockTxReq: req,
		},
	}
	sendData, _ = proto.handleSend(query, testPid, testAddr)
	err = proto.handleReceive(sendData, testPid, testAddr)
	assert.Equal(t, errSendStream, err)
	req.TxIndices = nil
	err = proto.handleReceive(sendData, testPid, testAddr)
	assert.Equal(t, errSendStream, err)
}

func Test_recvQueryReply(t *testing.T) {

	q := queue.New("test")
	go q.Start()
	defer q.Close()

	proto := newTestProtocolWithQueue(q)
	block := &types.Block{TxHash: []byte("test"), Txs: txList, Height: 10}
	blockHash := hex.EncodeToString(block.Hash(proto.ChainCfg))
	reply := &types.P2PBlockTxReply{
		BlockHash: blockHash,
	}
	sendData, _ := proto.handleSend(reply, testPid, testAddr)
	err := proto.handleReceive(sendData, testPid, testAddr)
	assert.Equal(t, types.ErrInvalidParam, err)
	proto.ltBlockCache.Add(blockHash, nil, 1)
	err = proto.handleReceive(sendData, testPid, testAddr)
	assert.Equal(t, types.ErrInvalidParam, err)
	proto.ltBlockCache.Add(blockHash, block, 1)
	//block组装失败,重新请求
	reply.Txs = []*types.Transaction{tx}
	reply.TxIndices = []int32{2}
	err = proto.handleReceive(sendData, testPid, testAddr)
	assert.Equal(t, errSendStream, err)

	//block组装失败,不再请求
	proto.ltBlockCache.Add(blockHash, block, 1)
	reply.TxIndices = nil
	err = proto.handleReceive(sendData, testPid, testAddr)
	assert.Equal(t, errBuildBlockFailed, err)
	//block组装成功
	newCli := q.Client()
	newCli.Sub("blockchain")
	reply.Txs = block.Txs
	block.TxHash = merkle.CalcMerkleRoot(proto.ChainCfg, block.Height, block.GetTxs())
	proto.ltBlockCache.Add(blockHash, block, 1)
	err = proto.handleReceive(sendData, testPid, testAddr)
	assert.Nil(t, err)
	msg := <-newCli.Recv()
	assert.Equal(t, types.EventBroadcastAddBlock, int(msg.Ty))
	blc, ok := msg.Data.(*types.BlockPid)
	assert.True(t, ok)
	assert.Equal(t, testPid, blc.Pid)
	assert.Equal(t, block.Hash(proto.ChainCfg), blc.Block.Hash(proto.ChainCfg))

}