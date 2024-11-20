package application

import (
	"log"
	"sync/atomic"

	storagepb "github.com/Bronn1/s3-test/proto"
	"golang.org/x/xerrors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type StorageServerConnections interface {
	GetFreeConnection() Connection
	GetConnectionByAdrr(addr string) *Connection
	AddNewConnection(addr string) error
	Close()
}

type Connection struct {
	grpcClient storagepb.ObjectStorageClient
	serverAddr string
}

type SimpleBalancerRoundR struct {
	servers  []Connection
	next     int32
	grpcConn []*grpc.ClientConn
}

func NewSimpleBalancerRoundR(storageAdrrs []string) (*SimpleBalancerRoundR, error) {
	balancer := SimpleBalancerRoundR{}
	for _, addr := range storageAdrrs {
		conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Printf("Failed to connect to storage server %v: %v", addr, err)
			continue
		}

		client := Connection{grpcClient: storagepb.NewObjectStorageClient(conn), serverAddr: addr}
		balancer.servers = append(balancer.servers, client)
		balancer.grpcConn = append(balancer.grpcConn, conn)
	}
	if len(balancer.servers) == 0 {
		return &SimpleBalancerRoundR{}, xerrors.Errorf("Cannot connect to any storage server")
	}

	return &balancer, nil
}

func (b *SimpleBalancerRoundR) GetFreeConnection() Connection {
	n := atomic.AddInt32(&b.next, 1)
	return b.servers[(int(n)-1)%len(b.servers)]
}

func (b *SimpleBalancerRoundR) GetConnectionByAdrr(addr string) *Connection {
	for _, conn := range b.servers {
		if conn.serverAddr != addr {
			continue
		}

		return &conn
	}

	return nil
}

func (b *SimpleBalancerRoundR) AddNewConnection(addr string) error {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}

	client := Connection{grpcClient: storagepb.NewObjectStorageClient(conn), serverAddr: addr}
	b.servers = append(b.servers, client)
	b.grpcConn = append(b.grpcConn, conn)
	return nil
}

func (b *SimpleBalancerRoundR) Close() {
	for _, conn := range b.grpcConn {
		conn.Close()
	}
}
