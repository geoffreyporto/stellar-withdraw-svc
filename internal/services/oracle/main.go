package oracle

import (
	"context"
	"github.com/stellar/go/clients/horizonclient"
	hProtocol "github.com/stellar/go/protocols/horizon"
	"github.com/tokend/stellar-withdraw-svc/internal/config"
	"github.com/tokend/stellar-withdraw-svc/internal/horizon/submit"
	"github.com/tokend/stellar-withdraw-svc/internal/services/request"
	"github.com/tokend/stellar-withdraw-svc/internal/services/watchlist"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/tokend/go/xdrbuild"
)

type StellarWithdrawDetails struct {
	TargetAddress string `json:"address"`
}

type Service struct {
	stellarRoot hProtocol.Root
	builder     *xdrbuild.Builder
	asset       watchlist.Details

	stellarClient horizonclient.ClientInterface
	txSubmitter   submit.Interface
	log           *logan.Entry
	stellarSource hProtocol.Account
	paymentCfg    config.PaymentConfig
	withdrawCfg   config.WithdrawConfig
	withdrawals   <-chan request.Details
}

func New(opts Opts) *Service {
	return &Service{
		log:           opts.Log,
		stellarSource: opts.StellarSource,
		stellarClient: opts.StellarClient,
		paymentCfg:    opts.PaymentConfig,
		withdrawCfg:   opts.WithdrawConfig,
		builder:       opts.Builder,
		txSubmitter:   opts.TXSubmitter,
		stellarRoot:   opts.StellarRoot,
		asset:         opts.AssetDetails,
		withdrawals:   opts.Withdrawals,
	}
}

type Opts struct {
	StellarSource  hProtocol.Account
	StellarRoot    hProtocol.Root
	PaymentConfig  config.PaymentConfig
	AssetDetails   watchlist.Details
	WithdrawConfig config.WithdrawConfig
	StellarClient  horizonclient.ClientInterface
	Builder        *xdrbuild.Builder
	TXSubmitter    submit.Interface
	Log            *logan.Entry
	Withdrawals    <-chan request.Details
}

func (s *Service) Run(ctx context.Context) {
	for details := range s.withdrawals {
		err := s.processWithdraw(ctx, details.ReviewableRequest, details.CreateWithdrawRequest)
		if err != nil {
			s.log.WithField("details", details).Warn("failed to process withdraw request")
		}
	}
}
