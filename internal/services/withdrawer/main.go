package withdrawer

import (
	"context"
	"github.com/stellar/go/clients/horizonclient"
	hProtocol "github.com/stellar/go/protocols/horizon"
	"github.com/tokend/stellar-withdraw-svc/internal/config"
	"github.com/tokend/stellar-withdraw-svc/internal/horizon"
	"github.com/tokend/stellar-withdraw-svc/internal/horizon/getters"
	"github.com/tokend/stellar-withdraw-svc/internal/horizon/submit"
	"github.com/tokend/stellar-withdraw-svc/internal/services/oracle"
	"github.com/tokend/stellar-withdraw-svc/internal/services/request"
	"github.com/tokend/stellar-withdraw-svc/internal/services/watchlist"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/tokend/go/xdrbuild"
)

type Service struct {
	assetWatcher  *watchlist.Service
	log           *logan.Entry
	config        config.Config
	stellarSource hProtocol.Account
	stellarRoot   hProtocol.Root
	spawned       map[string]bool
	assets        <-chan watchlist.Details
	builder       *xdrbuild.Builder
}

func New(cfg config.Config) *Service {
	assetWatcher := watchlist.New(watchlist.Opts{
		AssetOwner: cfg.WithdrawConfig().Owner.Address(),
		Streamer:   getters.NewDefaultAssetHandler(cfg.Horizon()),
		Log:        cfg.Log(),
	})
	builder, err := horizon.NewConnector(cfg.Horizon()).Builder()
	if err != nil {
		cfg.Log().WithError(err).Fatal("failed to make builder")
	}

	stellarSource, err := cfg.Stellar().AccountDetail(horizonclient.AccountRequest{
		AccountID: cfg.PaymentConfig().SourceAddress.Address(),
	})
	if err != nil {
		cfg.Log().WithError(err).Fatal("failed to get stellar source account")
	}

	root, err := cfg.Stellar().Root()
	if err != nil {
		cfg.Log().WithError(err).Fatal("failed to get root info for stellar network")
	}

	return &Service{
		log:           cfg.Log(),
		config:        cfg,
		assetWatcher:  assetWatcher,
		assets:        assetWatcher.GetChan(),
		spawned:       make(map[string]bool),
		builder:       builder,
		stellarSource: stellarSource,
		stellarRoot:   root,
	}
}

func (s *Service) Run(ctx context.Context) {
	go s.assetWatcher.Run(ctx)

	for asset := range s.assets {
		s.spawn(ctx, asset)
	}
}

func (s *Service) spawn(ctx context.Context, details watchlist.Details) {
	if s.spawned[details.Asset.ID] {
		return
	}
	withdrawStreamer := request.New(request.Opts{
		Reviewer:           s.config.WithdrawConfig().Owner.Address(),
		AssetDetails:       details,
		WithdrawalStreamer: getters.NewDefaultCreateWithdrawRequestHandler(s.config.Horizon()),
		Log:                s.log,
	})

	ch := withdrawStreamer.GetCh()
	oracleService := oracle.New(oracle.Opts{
		StellarSource:  s.stellarSource,
		StellarClient:  s.config.Stellar(),
		Log:            s.log,
		AssetDetails:   details,
		PaymentConfig:  s.config.PaymentConfig(),
		WithdrawConfig: s.config.WithdrawConfig(),
		TXSubmitter:    submit.New(s.config.Horizon()),
		Builder:        s.builder,
		StellarRoot:    s.stellarRoot,
		Withdrawals:    ch,
	})
	s.spawned[details.Asset.ID] = true

	go oracleService.Run(ctx)
	go withdrawStreamer.Run(ctx)

	s.log.WithFields(logan.F{
		"asset_code": details.Stellar.Code,
		"asset_type": details.Stellar.AssetType,
	}).Info("Started listening for withdrawals")
}
