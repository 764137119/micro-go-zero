package scheduler

import (
	"context"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
)

// LeaderElection etcd 分布式选主
type LeaderElection struct {
	etcdClient  *clientv3.Client
	electionKey string
	ttlSec      int
	scheduler   *Scheduler
	session     *concurrency.Session
	election    *concurrency.Election
	isLeader    bool
	nodeID      string
}

func NewLeaderElection(etcdClient *clientv3.Client, serviceName string, ttlSec int, scheduler *Scheduler) *LeaderElection {
	return &LeaderElection{
		etcdClient:  etcdClient,
		electionKey: fmt.Sprintf("/cronjob/leader/%s", serviceName),
		ttlSec:      ttlSec,
		scheduler:   scheduler,
		nodeID:      fmt.Sprintf("node-%d", time.Now().UnixNano()),
	}
}

// Elect 参与 Leader 选举（阻塞直到成为 Leader 并启动调度）
func (e *LeaderElection) Elect(ctx context.Context) {
	logx.Infof("Starting leader election for key: %s, node: %s", e.electionKey, e.nodeID)

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		err := e.tryElect(ctx)
		if err != nil {
			logx.Errorf("Leader election error: %v, retrying in 3s...", err)
			time.Sleep(3 * time.Second)
			continue
		}

		// 成为 Leader，启动调度器
		logx.Infof("Elected as leader! node: %s, key: %s", e.nodeID, e.electionKey)
		e.isLeader = true
		e.scheduler.Start(ctx)

		// 观察 Leader 变化（如果 session 断开会自动退出）
		e.watchLeadership(ctx)

		// Leader 身份丢失，停止调度器
		e.isLeader = false
		e.scheduler.Stop()
		logx.Infof("Leader role lost, stepping down...")
	}
}

func (e *LeaderElection) tryElect(ctx context.Context) error {
	var err error

	// 创建 Session
	e.session, err = concurrency.NewSession(e.etcdClient,
		concurrency.WithTTL(e.ttlSec),
		concurrency.WithContext(ctx),
	)
	if err != nil {
		return fmt.Errorf("create session: %w", err)
	}

	// 创建 Election
	e.election = concurrency.NewElection(e.session, e.electionKey)

	// 竞选 Leader
	if err = e.election.Campaign(ctx, e.nodeID); err != nil {
		e.session.Close()
		return fmt.Errorf("campaign: %w", err)
	}

	return nil
}

func (e *LeaderElection) watchLeadership(ctx context.Context) {
	// 等待 Session 结束（lease 过期或主动关闭）
	<-e.session.Done()
	logx.Infof("Leader session ended")
}

// IsLeader 检查当前节点是否为 Leader
func (e *LeaderElection) IsLeader() bool {
	return e.isLeader
}

// Resign 主动放弃 Leader 身份
func (e *LeaderElection) Resign() {
	if e.election != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if err := e.election.Resign(ctx); err != nil {
			logx.Errorf("Resign error: %v", err)
		}
	}
	if e.session != nil {
		_ = e.session.Close()
	}
	e.isLeader = false
}

// GetLeaderNodeID 获取当前 Leader 节点 ID
func (e *LeaderElection) GetLeaderNodeID() string {
	return e.nodeID
}
