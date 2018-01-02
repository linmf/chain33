package consensus

import (
	"code.aliyun.com/chain33/chain33/consensus/solo"
	"code.aliyun.com/chain33/chain33/queue"
	"code.aliyun.com/chain33/chain33/types"
	log "github.com/inconshreveable/log15"
)

func New(cfg *types.Consensus) Consumer {
	consensusType := cfg.Name
	if consensusType == "solo" {
		con := solo.NewSolo(cfg)
		return con
	} else if consensusType == "raft" {
		// TODO:
	} else if consensusType == "pbft" {
		// TODO:
	}
	panic("Unsupported consensus type")
}

type nopMiner struct{}

func (m *nopMiner) SetQueue(q *queue.Queue) {
	client := q.GetClient()
	client.Sub("consensus")
	go func() {
		for msg := range client.Recv() {
			log.Info("consensus recv", "msg", msg)
		}
	}()
}

func (m *nopMiner) Close() {
	log.Info("consensus module closed")
}

type Consumer interface {
	SetQueue(q *queue.Queue)
	Close()
}
