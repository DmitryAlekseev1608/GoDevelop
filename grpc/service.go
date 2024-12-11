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
	"google.golang.org/grpc/codes"
	"fmt"
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
			grpc.StreamInterceptor(streamInterceptor),
		)
		RegisterAdminServer(server, NewAdmin(accesses))
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
	accesses map[string][]string
}

func NewAdmin(accesses map[string][]string) *admin {
	return &admin{
		accesses: accesses,
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

func (a *admin) Logging(in *Nothing, adminServer Admin_LoggingServer) (err error) {
	out := &Event{}
	err = adminServer.Send(out)
	return
}

func (a *admin) Statistics(*StatInterval, Admin_StatisticsServer) (err error) {
	return
}

func (a *admin) mustEmbedUnimplementedAdminServer() {
}

func (b *biz) Logging() {
}

func (b *biz) Add(context.Context, *Nothing) (out *Nothing, err error) {
	return
}

func (b *biz) Test(ctx context.Context, in *Nothing) (out *Nothing, err error) {
	out = &Nothing{}
	return 
}

func (b *biz) Check(ctx context.Context, in *Nothing) (out *Nothing, err error) {
	out = &Nothing{}
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
	currentNameOfUser, exist := getCurrentNameOfUser(metadataOfUser)
	if !exist {
		return nil, status.Errorf(codes.Unauthenticated, "authentication failed")
	}
	methodWhichCall, exist := getMethodWhichCall(ctx)
	if !exist {
		return nil, status.Errorf(codes.Unauthenticated, "method is out in call")
	}
	accesses := make(map[string][]string)
	switch  info.Server.(type) {
		case (*biz):
			accesses = info.Server.(*biz).accesses
		case (*admin):
			accesses = info.Server.(*admin).accesses
	}
	if cannotProceed(*currentNameOfUser, methodWhichCall, accesses) {
		return nil, status.Errorf(codes.Unauthenticated, "authentication failed")
	}
	return handler(ctx, req)
}

func streamInterceptor(
	srv interface{},
	ss grpc.ServerStream,
	info *grpc.StreamServerInfo,
	handler grpc.StreamHandler,
) error {
	metadataOfUser, ok := metadata.FromIncomingContext(ss.Context())
	if !ok {
		return status.Errorf(16, "NOT_FOUND")
	}
	currentNameOfUser, exist := getCurrentNameOfUser(metadataOfUser)
	if !exist {
		return status.Errorf(codes.Unauthenticated, "authentication failed")
	}
	methodWhichCall, exist := getMethodWhichCall(ss.Context())
	if !exist {
		return status.Errorf(codes.Unauthenticated, "method is out in call")
	}
	accesses := make(map[string][]string)
	switch  srv.(type) {
		case (*biz):
			accesses = srv.(*biz).accesses
		case (*admin):
			accesses = srv.(*admin).accesses
	}
	if cannotProceed(*currentNameOfUser, methodWhichCall, accesses) {
		return status.Errorf(codes.Unauthenticated, "authentication failed")
	}
	return err
}

func cannotProceed(currentNameOfUser string, methodWhichCall string, accesses map[string][]string) bool {
	methodAcceptable, exist := accesses[currentNameOfUser]
	if exist {
		for _, val := range methodAcceptable {
			if val == methodWhichCall || (val[len(val)-1] == byte('*') && strings.Contains(methodWhichCall, val[:len(val)-2])) {
				test := val[:len(val)-2]
				fmt.Println(test)
				return !exist
			}
		}
		exist = false
	}
	return !exist
}

func getCurrentNameOfUser(metadataOfUser metadata.MD) (currentNameOfUser *string, exist bool) {
	_, exist = metadataOfUser["consumer"]
	if !exist {
		currentNameOfUser = nil
	} else {
		currentNameOfUser = &metadataOfUser["consumer"][0]
	}
	return
}

func createMapFromACLData(ACLData string) (accesses map[string][]string, err error) {
	accesses = make(map[string][]string)
	err = json.Unmarshal([]byte(ACLData), &accesses)
	if err != nil {
		log.Printf("Ошибка при парсинге ACLData: %v", err)
	}
	return
}

func getMethodWhichCall(ctx context.Context) (methodWhichCall string, exist bool) {
	methodWhichCall, exist = grpc.Method(ctx)
	return
}