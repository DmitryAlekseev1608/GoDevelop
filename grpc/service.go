package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func StartMyMicroservice(ctx context.Context, listenAddr string, ACLData string) (err error) {
	accesses, err := createMapFromACLData(ACLData)
	if err == nil {
		port := strings.Split(listenAddr, ":")[1]
		lis, err := net.Listen("tcp", ":"+port)
		if err != nil {
			log.Fatalln("can't listet port", err)
		}
		server := grpc.NewServer(
			grpc.UnaryInterceptor(authInterceptor),
			grpc.StreamInterceptor(streamInterceptor),
		)
		event := &[]*Event{}
		RegisterAdminServer(server, NewAdmin(event, accesses, listenAddr))
		RegisterBizServer(server, NewBiz(event, accesses, listenAddr))
		go func() {
			defer server.GracefulStop()
			go server.Serve(lis)
			<-ctx.Done()
		}()
	}
	return
}

type admin struct {
	log      *[]*Event
	host     string
	accesses map[string][]string
}

func NewAdmin(event *[]*Event, accesses map[string][]string, host string) *admin {
	return &admin{
		log:      event,
		host:     host,
		accesses: accesses,
	}
}

type biz struct {
	log      *[]*Event
	host     string
	accesses map[string][]string
}

func NewBiz(event *[]*Event, accesses map[string][]string, host string) *biz {
	return &biz{
		log:      event,
		host:     host,
		accesses: accesses,
	}
}

func (a *admin) Logging(in *Nothing, adminServer Admin_LoggingServer) (err error) {
	logs := []*Event{}
	counter := len(*a.log)
	ctx := adminServer.Context()
	metadata, _ := metadata.FromIncomingContext(ctx)
	currentNameOfUser := metadata["consumer"][0]
	currentNameOfMethod := "/main.Admin/Logging"
	addr := strings.Split(a.host, ":")[0] + ":"
	event := &Event{Consumer: currentNameOfUser, Method: currentNameOfMethod, Host: addr}
	*a.log = append(*a.log, event)
	for {
		logs = append(logs, (*a.log)[counter:]...)
		counter += len(logs)
		for idx, v := range logs {
			if v.Consumer == currentNameOfUser && v.Method == "/main.Admin/Logging" {
				logs = append(logs[:idx], logs[idx+1:]...)
			}
		}
		for len(logs) > 0 {
			out := (logs)[0]
			logs = append((logs)[:0], (logs)[(len(logs)):]...)
			out = &Event{Consumer: out.Consumer, Method: out.Method, Host: out.Host}
			err = adminServer.Send(out)
			if err != nil {
				return
			}
			select {
			case <-adminServer.Context().Done():
				return nil
			default:
				continue
			}
		}
	}
}

func (a *admin) Statistics(statInterval *StatInterval, adminStatisticsServer Admin_StatisticsServer) (err error) {
	byMethod := map[string]uint64{}
	byConsumer := map[string]uint64{}
	logs := []*Event{}
	counter := len(*a.log)
	ctx := adminStatisticsServer.Context()
	metadata, _ := metadata.FromIncomingContext(ctx)
	currentNameOfUser := metadata["consumer"][0]
	currentNameOfMethod := "/main.Admin/Statistics"
	addr := strings.Split(a.host, ":")[0] + ":"
	event := &Event{Consumer: currentNameOfUser, Method: currentNameOfMethod, Host: addr}
	*a.log = append(*a.log, event)
	for {
		logs = append(logs, (*a.log)[counter:]...)
		counter += len(logs)
		for idx, v := range logs {
			if v.Consumer == currentNameOfUser && v.Method == "/main.Admin/Statistics" {
				logs = append(logs[:idx], logs[idx+1:]...)
			}
		}
		for _, v := range logs {
			byMethod[v.Method] += 1
			byConsumer[v.Consumer] += 1
		}
		logs = []*Event{}
		stat := &Stat{
			ByMethod: byMethod,
			ByConsumer: byConsumer,
		}
		err = adminStatisticsServer.Send(stat)
		if err != nil {
			return
		}
		select {
		case <-adminStatisticsServer.Context().Done():
			return nil
		default:
			continue
		}
		time.Sleep(time.Duration(statInterval.IntervalSeconds) * time.Second)
	}
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
	metadata, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return
	}
	currentNameOfUser := metadata["consumer"][0]
	currentNameOfMethod, _ := grpc.Method(ctx)
	addr := strings.Split(b.host, ":")[0] + ":"
	event := &Event{Consumer: currentNameOfUser, Method: currentNameOfMethod, Host: addr}
	*b.log = append(*b.log, event)
	return
}

func (b *biz) Check(ctx context.Context, in *Nothing) (out *Nothing, err error) {
	out = &Nothing{}
	metadata, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return
	}
	currentNameOfUser := metadata["consumer"][0]
	currentNameOfMethod, _ := grpc.Method(ctx)
	addr := strings.Split(b.host, ":")[0] + ":"
	event := &Event{Consumer: currentNameOfUser, Method: currentNameOfMethod, Host: addr}
	*b.log = append(*b.log, event)
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
	switch info.Server.(type) {
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
	accesses := srv.(*admin).accesses
	if cannotProceed(*currentNameOfUser, methodWhichCall, accesses) {
		return status.Errorf(codes.Unauthenticated, "authentication failed")
	}
	return handler(srv, ss)
}

func cannotProceed(currentNameOfUser string, methodWhichCall string, accesses map[string][]string) bool {
	methodAcceptable, exist := accesses[currentNameOfUser]
	if exist {
		for _, val := range methodAcceptable {
			if val == methodWhichCall || (val[len(val)-1] == byte('*') && strings.Contains(methodWhichCall, val[:len(val)-2])) {
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
	fmt.Println()
	return
}

func getMethodWhichCall(ctx context.Context) (methodWhichCall string, exist bool) {
	methodWhichCall, exist = grpc.Method(ctx)
	return
}
