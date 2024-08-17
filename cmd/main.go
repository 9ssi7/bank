package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/9ssi7/bank/api/rest"
	"github.com/9ssi7/bank/api/rpc"
	"github.com/9ssi7/bank/config"
	"github.com/9ssi7/bank/internal/infra/db"
	"github.com/9ssi7/bank/internal/infra/db/migration"
	"github.com/9ssi7/bank/internal/infra/eventer"
	"github.com/9ssi7/bank/internal/infra/keyval"
	"github.com/9ssi7/bank/internal/infra/observer"
	"github.com/9ssi7/bank/internal/repository"
	"github.com/9ssi7/bank/internal/usecase"
	"github.com/9ssi7/bank/pkg/cancel"
	"github.com/9ssi7/bank/pkg/retry"
	"github.com/9ssi7/bank/pkg/token"
	"github.com/9ssi7/bank/pkg/validation"
	"github.com/redis/go-redis/v9"
)

var once sync.Once
var a app

type app struct {
	db       *sql.DB
	rdb      *redis.Client
	eventSrv *eventer.Srv
	obsrvr   *observer.Srv
	valSrv   *validation.Srv
	tokenSrv *token.Service
	cnf      *config.App

	authUseCase    *usecase.AuthUseCase
	accountUseCase *usecase.AccountUseCase
}

func init() {
	once.Do(func() {
		ctx := context.Background()
		if err := a.initialize(ctx); err != nil {
			log.Fatalf("failed to initialize app: %v", err)
		}
		a.valSrv = validation.New()
		userRepo := repository.NewUserSqlRepo(a.db)
		accountRepo := repository.NewAccountSqlRepo(a.db)
		transactionRepo := repository.NewTransactionSqlRepo(a.db)
		verifyRepo := repository.NewVerifyRedisRepo(a.rdb)
		sessionRepo := repository.NewSessionRedisRepo(a.rdb)
		a.authUseCase = &usecase.AuthUseCase{
			TokenSrv:    a.tokenSrv,
			EventSrv:    a.eventSrv,
			VerifyRepo:  verifyRepo,
			UserRepo:    userRepo,
			SessionRepo: sessionRepo,
		}
		a.accountUseCase = &usecase.AccountUseCase{
			EventSrv:        a.eventSrv,
			AccountRepo:     accountRepo,
			TransactionRepo: transactionRepo,
			UserRepo:        userRepo,
		}
	})
}

func main() {
	tracer := a.obsrvr.GetTracer()
	meter := a.obsrvr.GetMeter()

	restSrv := rest.New(rest.Config{
		Tracer:           tracer,
		Meter:            meter,
		ValidationSrv:    a.valSrv,
		AuthUseCase:      a.authUseCase,
		AccountUseCase:   a.accountUseCase,
		Host:             a.cnf.Rest.Host,
		Port:             a.cnf.Rest.Port,
		Domain:           a.cnf.Rest.Domain,
		AllowedMethods:   a.cnf.Rest.AllowMethods,
		AllowedHeaders:   a.cnf.Rest.AllowHeaders,
		AllowedOrigins:   a.cnf.Rest.AllowOrigins,
		ExposeHeaders:    a.cnf.Rest.ExposeHeader,
		AllowCredentials: a.cnf.Rest.AllowCred,
		Locales:          a.cnf.I18n.Locales,
		TurnstileSecret:  a.cnf.Turnstile.Secret,
		TurnstileSkip:    a.cnf.Turnstile.Skip,
	})

	rpcSrv := rpc.New(rpc.Config{
		Tracer:          tracer,
		Meter:           meter,
		ValidationSrv:   a.valSrv,
		AuthUseCase:     a.authUseCase,
		AccountUseCasee: a.accountUseCase,
		Domain:          a.cnf.Rpc.Domain,
		Port:            a.cnf.Rpc.Port,
	})

	var wg sync.WaitGroup
	wg.Add(3)
	go func() {
		defer wg.Done()
		if err := restSrv.Listen(); err != nil {
			log.Fatalf("failed to start rest server: %v", err)
		}
	}()
	go func() {
		defer wg.Done()
		if err := rpcSrv.Listen(); err != nil {
			log.Fatalf("failed to start rpc server: %v", err)
		}
	}()
	shutdownCh := make(chan os.Signal, 1)
	signal.Notify(shutdownCh, os.Interrupt)
	go func() {
		defer wg.Done()
		<-shutdownCh
		log.Println("application is shutting down...")
		if err := a.disconnect(context.Background(), restSrv.Shutdown, rpcSrv.Shutdown); err != nil {
			log.Fatalf("failed to disconnect: %v", err)
		}
	}()

	wg.Wait()
	fmt.Println("All servers are stopped.")
}

func (a *app) loadConfig() error {
	var configs config.App
	if err := config.Bind(&configs); err != nil {
		return err
	}
	a.cnf = &configs
	return nil
}

func (a *app) initialize(ctx context.Context) error {
	return retry.Run(func() error {
		return cancel.NewWithTimeout(ctx, 10*time.Second, func(ctx context.Context) error {
			if err := a.loadConfig(); err != nil {
				return err
			}
			obsrvr := observer.New(observer.Config{
				Name:     a.cnf.Observer.Name,
				Endpoint: a.cnf.Observer.Endpoint,
				UseSSL:   a.cnf.Observer.UseSSL,
			})
			if err := obsrvr.Init(ctx); err != nil {
				return err
			}
			db, err := db.New(ctx, db.Config{
				Host:     a.cnf.Database.Host,
				Port:     a.cnf.Database.Port,
				User:     a.cnf.Database.User,
				Password: a.cnf.Database.Pass,
				DBName:   a.cnf.Database.Name,
				SSLMode:  a.cnf.Database.SslMode,
			})
			if err != nil {
				return err
			}
			if a.cnf.Database.Migrate {
				if err := migration.Run(ctx, db); err != nil {
					return err
				}
			}
			rdb, err := keyval.New(ctx, keyval.Config{
				Host: a.cnf.Keyval.Host,
				Port: a.cnf.Keyval.Port,
				Pw:   a.cnf.Keyval.Pw,
				Db:   a.cnf.Keyval.Db,
			})
			if err != nil {
				return err
			}
			if err := rdb.Ping(ctx).Err(); err != nil {
				return err
			}
			tknSrv, err := token.New(token.Config{
				PublicKeyFile:  a.cnf.Token.PublicKeyFile,
				PrivateKeyFile: a.cnf.Token.PrivateKeyFile,
				Project:        a.cnf.Token.Project,
				SignMethod:     a.cnf.Token.SignMethod,
			})
			if err != nil {
				return err
			}
			a.eventSrv = eventer.New(a.cnf.Event.StreamUrl)
			if err := a.eventSrv.Connect(ctx); err != nil {
				return err
			}
			a.tokenSrv = tknSrv
			a.rdb = rdb
			a.db = db
			a.obsrvr = obsrvr
			return nil
		})
	}, retry.DefaultConfig)
}

type disconFunc func(context.Context) error

func (a *app) disconnectAll(ctx context.Context, fns ...disconFunc) error {
	for _, fn := range fns {
		if err := fn(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (a *app) disconnect(ctx context.Context, fns ...disconFunc) error {
	return cancel.NewWithTimeout(ctx, 5*time.Second, func(ctx context.Context) error {
		fns = append(fns, a.obsrvr.Shutdown, a.closeDB, a.closeKeyval, a.eventSrv.Disconnect)
		return a.disconnectAll(ctx, fns...)
	})
}

func (a *app) closeDB(ctx context.Context) error {
	return a.db.Close()
}

func (a *app) closeKeyval(ctx context.Context) error {
	return a.rdb.Close()
}
