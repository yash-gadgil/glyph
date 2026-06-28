package handlers

import (
	"errors"
	"os"

	"github.com/yash-gadgil/glyph/pkg/logger"
	"github.com/yash-gadgil/glyph/services/gateway/server/utils"
	advisorpb "github.com/yash-gadgil/glyph/services/gen/golang/advisor"
	authpb "github.com/yash-gadgil/glyph/services/gen/golang/auth"
	mrktpb "github.com/yash-gadgil/glyph/services/gen/golang/mrktdata"
	ordrpb "github.com/yash-gadgil/glyph/services/gen/golang/order"
	strategypb "github.com/yash-gadgil/glyph/services/gen/golang/strategy"
	userpb "github.com/yash-gadgil/glyph/services/gen/golang/user"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type Config struct {
	GatewayServiceAddr string

	userConn        *grpc.ClientConn
	accountClient   userpb.AccountServiceClient
	watchlistClient userpb.WatchlistServiceClient
	portfolioClient userpb.PortfolioServiceClient

	strategyConn   *grpc.ClientConn
	strategyClient strategypb.StrategyServiceClient

	authConn   *grpc.ClientConn
	AuthClient authpb.AuthServiceClient
	jwks       *jwksCache

	mrktdataConn   *grpc.ClientConn
	mrktdataClient mrktpb.MrktdataServiceClient

	orderConn   *grpc.ClientConn
	orderClient ordrpb.OrderServiceClient

	advisorConn   *grpc.ClientConn
	advisorClient advisorpb.AdvisorServiceClient

	log *zap.Logger
}

func NewFromEnv() *Config {
	cfg := &Config{}
	cfg.log = logger.New("gateway-service")

	if v := os.Getenv("GATEWAY_SVC_PORT"); v != "" {
		cfg.GatewayServiceAddr = v
	}

	if v := os.Getenv("USER_SVC_PORT"); v != "" {
		conn, err := utils.GetGrpcClient(v)
		if err != nil {
			cfg.log.Error("user_service_dial_failed", zap.String("addr", v), zap.Error(err))
		} else {
			cfg.userConn = conn
			cfg.accountClient = userpb.NewAccountServiceClient(conn)
			cfg.watchlistClient = userpb.NewWatchlistServiceClient(conn)
			cfg.portfolioClient = userpb.NewPortfolioServiceClient(conn)
		}
	}

	if v := os.Getenv("AUTH_SVC_PORT"); v != "" {
		conn, err := utils.GetGrpcClient(v)
		if err != nil {
			cfg.log.Error("auth_service_dial_failed", zap.String("addr", v), zap.Error(err))
		} else {
			cfg.authConn = conn
			cfg.AuthClient = authpb.NewAuthServiceClient(conn)
			cfg.jwks = newJWKSCache(cfg.AuthClient)
		}
	}

	if v := os.Getenv("MRKTDATA_SVC_PORT"); v != "" {
		conn, err := utils.GetGrpcClient(v)
		if err != nil {
			cfg.log.Error("mrktdata_service_dial_failed", zap.String("addr", v), zap.Error(err))
		} else {
			cfg.mrktdataConn = conn
			cfg.mrktdataClient = mrktpb.NewMrktdataServiceClient(conn)
		}
	}

	if v := os.Getenv("ORDER_SVC_PORT"); v != "" {
		conn, err := utils.GetGrpcClient(v)
		if err != nil {
			cfg.log.Error("order_service_dial_failed", zap.String("addr", v), zap.Error(err))
		} else {
			cfg.orderConn = conn
			cfg.orderClient = ordrpb.NewOrderServiceClient(conn)
		}
	} else {
		cfg.log.Warn("order_service_unconfigured", zap.String("reason", "ORDER_SVC_PORT not set"))
	}

	if v := os.Getenv("ADVISOR_SVC_PORT"); v != "" {
		conn, err := utils.GetGrpcClient(v)
		if err != nil {
			cfg.log.Error("advisor_service_dial_failed", zap.String("addr", v), zap.Error(err))
		} else {
			cfg.advisorConn = conn
			cfg.advisorClient = advisorpb.NewAdvisorServiceClient(conn)
		}
	} else {
		cfg.log.Warn("advisor_service_unconfigured", zap.String("reason", "ADVISOR_SVC_PORT not set"))
	}

	if v := os.Getenv("STRATEGY_SVC_PORT"); v != "" {
		conn, err := utils.GetGrpcClient(v)
		if err != nil {
			cfg.log.Error("strategy_service_dial_failed", zap.String("addr", v), zap.Error(err))
		} else {
			cfg.strategyConn = conn
			cfg.strategyClient = strategypb.NewStrategyServiceClient(conn)
		}
	} else {
		cfg.log.Warn("strategy_service_unconfigured", zap.String("reason", "STRATEGY_SVC_PORT not set"))
	}

	return cfg
}

func NewTestConfig(authClient authpb.AuthServiceClient) *Config {
	return &Config{
		AuthClient: authClient,
		jwks:       newJWKSCache(authClient),
		log:        zap.NewNop(),
	}
}

func (cfg *Config) ExpireJWKSCacheForTest() {
	if cfg.jwks != nil {
		cfg.jwks.expireForTest()
	}
}

func (cfg *Config) WithAccountClient(c userpb.AccountServiceClient) *Config {
	cfg.accountClient = c
	return cfg
}

func (cfg *Config) WithPortfolioClient(c userpb.PortfolioServiceClient) *Config {
	cfg.portfolioClient = c
	return cfg
}

func (cfg *Config) WithOrderClient(c ordrpb.OrderServiceClient) *Config {
	cfg.orderClient = c
	return cfg
}

func (cfg *Config) WithWatchlistClient(c userpb.WatchlistServiceClient) *Config {
	cfg.watchlistClient = c
	return cfg
}

func (cfg *Config) WithMrktdataClient(c mrktpb.MrktdataServiceClient) *Config {
	cfg.mrktdataClient = c
	return cfg
}

func (cfg *Config) WithStrategyClient(c strategypb.StrategyServiceClient) *Config {
	cfg.strategyClient = c
	return cfg
}

func (cfg *Config) WithAdvisorClient(c advisorpb.AdvisorServiceClient) *Config {
	cfg.advisorClient = c
	return cfg
}

func (cfg *Config) Close() error {
	var errs []error
	for _, conn := range []*grpc.ClientConn{cfg.userConn, cfg.authConn, cfg.mrktdataConn, cfg.orderConn, cfg.advisorConn, cfg.strategyConn} {
		if conn != nil {
			if err := conn.Close(); err != nil {
				errs = append(errs, err)
			}
		}
	}
	return errors.Join(errs...)
}
