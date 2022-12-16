package infraContainer

import (
	"context"
	"fmt"

	"github.com/infranyx/go-grpc-template/pkg/config"
	"github.com/infranyx/go-grpc-template/pkg/grpc"
	httpEcho "github.com/infranyx/go-grpc-template/pkg/http/echo"
	"github.com/infranyx/go-grpc-template/pkg/kafka"
	"github.com/infranyx/go-grpc-template/pkg/logger"
	"github.com/infranyx/go-grpc-template/pkg/postgres"
	kk "github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

type IContainer struct {
	GrpcServer  grpc.GrpcServer // grpc.GrpcServer : Interface
	EchoServer  httpEcho.EchoHttpServer
	Logger      *zap.Logger
	Cfg         *config.Config
	Pg          *postgres.Postgres
	KafkaWriter *kafka.Writer
	KafkaReader *kafka.Reader
}

func NewIC(ctx context.Context) (*IContainer, func(), error) {
	var downFns []func()
	down := func() {
		for _, df := range downFns {
			df()
		}
	}

	grpcServerConfig := &grpc.GrpcConfig{
		Port:        config.Conf.Grpc.Port,
		Host:        config.Conf.Grpc.Host,
		Development: config.IsDevEnv(),
	}
	grpcServer := grpc.NewGrpcServer(grpcServerConfig)
	downFns = append(downFns, func() {
		grpcServer.GracefulShutdown()
	})

	echoServerConfig := &httpEcho.EchoHttpConfig{
		Port:        config.Conf.Http.Port,
		Development: config.IsDevEnv(),
		BasePath:    "/api/v1",
	}
	echoServer := httpEcho.NewEchoHttpServer(echoServerConfig)
	echoServer.SetupDefaultMiddlewares()
	downFns = append(downFns, func() {
		echoServer.GracefulShutdown(ctx)
	})

	pg, err := postgres.NewPgConn(ctx, &postgres.PgConf{
		Host:    config.Conf.Postgres.Host,
		Port:    config.Conf.Postgres.Port,
		User:    config.Conf.Postgres.User,
		Pass:    config.Conf.Postgres.Pass,
		DBName:  config.Conf.Postgres.DBName,
		SslMode: config.Conf.Postgres.SslMode,
	})
	if err != nil {
		return nil, down, fmt.Errorf("could not initialize database connection using sqlx %s", err)
	}
	downFns = append(downFns, func() {
		pg.Close()
	})

	kwc := &kafka.WriterConf{
		Brokers: []string{"localhost:9092"},
		Topic:   "test-topic",
	}
	kw := kafka.NewKafkaWriter(kwc)

	downFns = append(downFns, func() {
		kw.Client.Close()
	})

	// TODO : rm after test
	errr := kw.Client.WriteMessages(context.Background(),
		kk.Message{
			Key:   []byte("Key-A"),
			Value: []byte("Hello World!"),
		},
		kk.Message{
			Key:   []byte("Key-B"),
			Value: []byte("One!"),
		},
		kk.Message{
			Key:   []byte("Key-C"),
			Value: []byte("Two!"),
		},
	)
	if errr != nil {
		fmt.Println("failed to write messages:", errr)
	}

	krc := &kafka.ReaderConf{
		Brokers: []string{"localhost:9092"},
		Topic:   "test-topic",
		GroupID: "test-id",
	}
	kr := kafka.NewKafkaReader(krc)
	downFns = append(downFns, func() {
		kr.Client.Close()
	})

	// TODO : rm after test
	kr.Client.SetOffset(42)

	for {
		m, err := kr.Client.ReadMessage(context.Background())
		if err != nil {
			break
		}
		fmt.Printf("message at offset %d: %s = %s\n", m.Offset, string(m.Key), string(m.Value))
	}

	ic := &IContainer{Cfg: config.Conf, Logger: logger.Zap, GrpcServer: grpcServer, EchoServer: echoServer, Pg: pg, KafkaWriter: kw, KafkaReader: kr}

	return ic, down, nil
}
