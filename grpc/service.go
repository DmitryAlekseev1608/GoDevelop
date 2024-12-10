package main

import (
	"context"
	"net"
	"log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"strings"
	"encoding/json"
)

func StartMyMicroservice(ctx context.Context, listenAddr string, ACLData string) (err error) {
	accesses, err := createMapFromACLData(ACLData)
	if err == nil {
		port := strings.Split(listenAddr, ":")[1]
		lis, err := net.Listen("tcp", ":" + port)
		if err != nil {
			log.Fatalln("can't listet port", err)
		}
		server := grpc.NewServer(
			grpc.UnaryInterceptor(authInterceptor),
		)
		RegisterAdminServer(server, NewAdmin())
		RegisterBizServer(server, NewBiz(accesses))
		go func() {
			defer server.GracefulStop()
			go server.Serve(lis)
			<- ctx.Done()
			}()
		}
	return
}

type admin struct {
}

func NewAdmin() *admin {
	return &admin{
	}
}

type biz struct {
	accesses map[string][]string
}

func NewBiz(accesses map[string][]string) *biz {
	return &biz{
		accesses: accesses,
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

func (b *biz) Test(ctx context.Context, in *Nothing) (out *Nothing, err error) {
	return
}

func (b *biz) Check(context.Context, *Nothing) (anyThing *Nothing, err error) {
	return
}

func (b *biz) mustEmbedUnimplementedBizServer() {
}

func authInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	metadataOfUser, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(16, "NOT_FOUND")
	}
	currentNameOfUser := metadataOfUser["consumer"][0]
	accesses := info.Server.(*biz).accesses
	if cannotProceed(currentNameOfUser, accesses) {
		return nil, status.Errorf(7, "PERMISSION_DENIED")
	}
	return handler(ctx, req)
}

func cannotProceed(currentNameOfUser string, accesses map[string][]string) bool {
	_, exist := accesses[currentNameOfUser]
	return exist
}

func createMapFromACLData(ACLData string) (accesses map[string][]string, err error) {
	accesses = make(map[string][]string)
	err = json.Unmarshal([]byte(ACLData), &accesses)
	if err != nil {
		log.Printf("Ошибка при парсинге ACLData: %v", err)
	}
	return
}
