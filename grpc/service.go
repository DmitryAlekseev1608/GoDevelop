package main

import (
	"context"
	"net"
	"log"
	"google.golang.org/grpc"
)

func StartMyMicroservice(ctx context.Context, listenAddr string, ACLData string) (err error) {
	lis, err := net.Listen("tcp", ":8081")
	if err != nil {
		log.Fatalln("can't listet port", err)
	}
	server := grpc.NewServer()
	RegisterAdminServer(server, NewAdmin())
	RegisterBizServer(server, NewBiz())
	go func() {
		defer server.GracefulStop()
		go server.Serve(lis)
		<- ctx.Done()
		}()
	return
}

type admin struct {
}

func NewAdmin() *admin {
	return &admin{
	}
}

type biz struct {
}

func NewBiz() *biz {
	return &biz{
	}
}

func (a *admin) Logging(*Nothing, Admin_LoggingServer) (err error) {
	return
}

func (a *admin) Statistics(*StatInterval, Admin_StatisticsServer) (err error) {
	return
}

func (a *admin) mustEmbedUnimplementedAdminServer() {
}

func (b *biz) Logging() {
}

func (b *biz) Add(context.Context, *Nothing) (anyThing *Nothing, err error) {
	return
}

func (b *biz) Test(context.Context, *Nothing) (anyThing *Nothing, err error) {
	return
}

func (b *biz) Check(context.Context, *Nothing) (anyThing *Nothing, err error) {
	return
}

func (b *biz) mustEmbedUnimplementedBizServer() {
}
