// Use of this source code is governed by a Apache-2.0 license that can be found
// at https://github.com/etcd-io/etcd/blob/main/client/v3/mock/mockserver/mockserver.go

package etcdmock

import (
	"context"
	"fmt"
	pb "github.com/coreos/etcd/etcdserver/etcdserverpb"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"google.golang.org/grpc"
	"net"
	"sync"
)

// For reference only, add, delete, query
func init() {
	var err error

	_, err = startMockServer()
	if err != nil {
		panic(err)
	}
}

type server struct {
	ln         net.Listener
	GrpcServer *grpc.Server
}

func startMockServer() (ms *server, err error) {
	ln, err := net.Listen("tcp", "localhost:2379")
	if err != nil {
		return nil, fmt.Errorf("failed to listen %v", err)
	}

	ms = &server{ln: ln}

	ms.start()
	return ms, nil
}

func (ms *server) start() {
	svr := grpc.NewServer()
	pb.RegisterKVServer(svr, &mockKVServer{data: make(map[string]string)})
	pb.RegisterLeaseServer(svr, &mockLeaseServer{})
	ms.GrpcServer = svr

	go func(svr *grpc.Server, l net.Listener) {
		svr.Serve(l)
	}(ms.GrpcServer, ms.ln)
}

func (ms *server) stop() {
	ms.GrpcServer.Stop()
}

type mockKVServer struct {
	sync.RWMutex
	data map[string]string
}

// Range one to return
func (m *mockKVServer) Range(c context.Context, req *pb.RangeRequest) (*pb.RangeResponse, error) {
	m.RLock()
	defer m.RUnlock()

	var kv []*mvccpb.KeyValue
	v, ok := m.data[string(req.GetKey())]
	if ok {
		tmp := new(mvccpb.KeyValue)
		tmp.Key = req.Key
		tmp.Value = []byte(v)
		kv = append(kv, tmp)
	}
	return &pb.RangeResponse{
		Kvs:   kv,
		Count: 1,
	}, nil
}

// Put one to cache data
func (m *mockKVServer) Put(c context.Context, req *pb.PutRequest) (*pb.PutResponse, error) {
	m.Lock()
	defer m.Unlock()

	pre := new(mvccpb.KeyValue)
	v, ok := m.data[string(req.Key)]
	if ok {
		pre.Key = req.Key
		pre.Value = []byte(v)
		pre.Lease = req.Lease
	}

	m.data[string(req.Key)] = string(req.Value)
	return &pb.PutResponse{PrevKv: pre}, nil
}

// DeleteRange No need to implement for mock
func (m *mockKVServer) DeleteRange(c context.Context, req *pb.DeleteRangeRequest) (*pb.DeleteRangeResponse, error) {
	m.RLock()
	defer m.RUnlock()

	_, ok := m.data[string(req.GetKey())]
	if ok {
		delete(m.data, string(req.GetKey()))
	}
	return &pb.DeleteRangeResponse{}, nil
}

// Txn No need to implement
func (m *mockKVServer) Txn(c context.Context, req *pb.TxnRequest) (*pb.TxnResponse, error) {
	return &pb.TxnResponse{}, nil
}

// Compact No need to implement
func (m *mockKVServer) Compact(c context.Context, req *pb.CompactionRequest) (*pb.CompactionResponse, error) {
	return &pb.CompactionResponse{}, nil
}

func (m *mockKVServer) Lease(context.Context, *pb.LeaseGrantRequest) (*pb.LeaseGrantResponse, error) {
	return &pb.LeaseGrantResponse{}, nil
}

type mockLeaseServer struct{}

func (s *mockLeaseServer) LeaseGrant(context.Context, *pb.LeaseGrantRequest) (*pb.LeaseGrantResponse, error) {
	return &pb.LeaseGrantResponse{}, nil
}

func (s *mockLeaseServer) LeaseRevoke(context.Context, *pb.LeaseRevokeRequest) (*pb.LeaseRevokeResponse, error) {
	return &pb.LeaseRevokeResponse{}, nil
}

func (s *mockLeaseServer) LeaseKeepAlive(pb.Lease_LeaseKeepAliveServer) error {
	return nil
}

func (s *mockLeaseServer) LeaseTimeToLive(context.Context, *pb.LeaseTimeToLiveRequest) (*pb.LeaseTimeToLiveResponse, error) {
	return &pb.LeaseTimeToLiveResponse{}, nil
}

func (s *mockLeaseServer) LeaseLeases(context.Context, *pb.LeaseLeasesRequest) (*pb.LeaseLeasesResponse, error) {
	return &pb.LeaseLeasesResponse{}, nil
}
