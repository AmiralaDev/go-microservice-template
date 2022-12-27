package infraContainer

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/infranyx/go-grpc-template/pkg/config"
	"github.com/infranyx/go-grpc-template/pkg/grpc"
	echoHttp "github.com/infranyx/go-grpc-template/pkg/http/echo"
	kafkaConsumer "github.com/infranyx/go-grpc-template/pkg/kafka/consumer"
	kafkaProducer "github.com/infranyx/go-grpc-template/pkg/kafka/producer"
	"github.com/infranyx/go-grpc-template/pkg/logger"
	"github.com/infranyx/go-grpc-template/pkg/postgres"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

type IContainer struct {
	Config         *config.Config
	Logger         *zap.Logger
	Postgres       *postgres.Postgres
	GrpcServer     grpc.Server
	EchoHttpServer echoHttp.ServerInterface
	KafkaWriter    *kafkaProducer.Writer
	KafkaReader    *kafkaConsumer.Reader
}

func NewIC(ctx context.Context) (*IContainer, func(), error) {
	var downFns []func()
	down := func() {
		for _, df := range downFns {
			df()
		}
	}

	se := sentry.Init(sentry.ClientOptions{
		Dsn:              config.BaseConfig.Sentry.Dsn,
		TracesSampleRate: 1.0,
		EnableTracing:    config.IsDevEnv(),
	})
	if se != nil {
		log.Fatalf("can not initialize sentry with error:  %s", se)
	}
	downFns = append(downFns, func() {
		sentry.Flush(2 * time.Second)
	})

	grpcServerConfig := &grpc.Config{
		Port:        config.BaseConfig.Grpc.Port,
		Host:        config.BaseConfig.Grpc.Host,
		Development: config.IsDevEnv(),
	}
	grpcServer := grpc.NewGrpcServer(grpcServerConfig)
	downFns = append(downFns, func() {
		grpcServer.GracefulShutdown()
	})

	echoServerConfig := &echoHttp.ServerConfig{
		Port:     config.BaseConfig.Http.Port,
		BasePath: "/api/v1",
		IsDev:    config.IsDevEnv(),
	}
	echoServer := echoHttp.NewServer(echoServerConfig)
	echoServer.SetupDefaultMiddlewares()
	downFns = append(downFns, func() {
		echoServer.GracefulShutdown(ctx)
	})

	pg, err := postgres.NewConnection(ctx, &postgres.PgConfig{
		Host:    config.BaseConfig.Postgres.Host,
		Port:    config.BaseConfig.Postgres.Port,
		User:    config.BaseConfig.Postgres.User,
		Pass:    config.BaseConfig.Postgres.Pass,
		DBName:  config.BaseConfig.Postgres.DBName,
		SslMode: config.BaseConfig.Postgres.SslMode,
	})
	if err != nil {
		return nil, down, fmt.Errorf("can not connect to database using sqlx with error: %s", err)
	}
	downFns = append(downFns, func() {
		pg.Close()
	})

	kwc := &kafkaProducer.WriterConfig{
		Brokers:      config.BaseConfig.Kafka.ClientBrokers,
		Topic:        config.BaseConfig.Kafka.Topic,
		RequiredAcks: kafka.RequireAll,
	}
	kw := kafkaProducer.NewKafkaWriter(kwc)
	downFns = append(downFns, func() {
		kw.Client.Close()
	})

	krc := &kafkaConsumer.ReaderConfig{
		Brokers: config.BaseConfig.Kafka.ClientBrokers,
		Topic:   config.BaseConfig.Kafka.Topic,
		GroupID: config.BaseConfig.Kafka.ClientGroupId,
	}
	kr := kafkaConsumer.NewKafkaReader(krc)
	downFns = append(downFns, func() {
		kr.Client.Close()
	})

	ic := &IContainer{
		Config:         config.BaseConfig,
		Logger:         logger.Zap,
		Postgres:       pg,
		GrpcServer:     grpcServer,
		EchoHttpServer: echoServer,
		KafkaWriter:    kw, KafkaReader: kr}

	return ic, down, nil
}
