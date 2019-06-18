package withdrawer

import (
	"context"
	"github.com/tokend/stellar-withdraw-svc/internal/horizon/getters"
	"github.com/tokend/stellar-withdraw-svc/internal/horizon/submit"
	"github.com/tokend/stellar-withdraw-svc/internal/services/oracle"
	"github.com/tokend/stellar-withdraw-svc/internal/services/request"
	"github.com/tokend/stellar-withdraw-svc/internal/services/watchlist"
	"gitlab.com/distributed_lab/logan/v3"
)

func (s *Service) Run(ctx context.Context) {
	go s.assetWatcher.Run(ctx)

	s.Add(2)
	go s.spawner(ctx)
	go s.cancellor(ctx)
	s.Wait()

}

func (s *Service) spawner(ctx context.Context) {
	defer s.Done()
	for asset := range s.assetsToAdd {
		if _, ok := s.spawned.Load(asset.ID); !ok {
			s.spawn(ctx, asset)
		}
	}
}

func (s *Service) cancellor(ctx context.Context) {
	defer s.Done()
	for asset := range s.assetsToRemove {
		if raw, ok := s.spawned.Load(asset); ok {
			cancelFunc := raw.(context.CancelFunc)
			cancelFunc()
			s.spawned.Delete(asset)
		}
	}
}

func (s *Service) spawn(ctx context.Context, details watchlist.Details) {

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

	innerCtx, cancelFunc := context.WithCancel(ctx)
	s.spawned.Store(details.Asset.ID, cancelFunc)

	go oracleService.Run(innerCtx)
	go withdrawStreamer.Run(innerCtx)

	s.log.WithFields(logan.F{
		"asset_code": details.Stellar.Code,
		"asset_type": details.Stellar.AssetType,
	}).Info("Started listening for withdrawals")
}
