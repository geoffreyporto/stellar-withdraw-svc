package oracle

import (
	"context"
	"github.com/stellar/go/clients/horizonclient"
	hProtocol "github.com/stellar/go/protocols/horizon"
	"github.com/tokend/stellar-withdraw-svc/internal/config"
	"github.com/tokend/stellar-withdraw-svc/internal/horizon/getters"
	"github.com/tokend/stellar-withdraw-svc/internal/horizon/submit"
	"github.com/tokend/stellar-withdraw-svc/internal/services/watchlist"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
	"gitlab.com/distributed_lab/running"
	"gitlab.com/tokend/go/xdrbuild"
	regources "gitlab.com/tokend/regources/generated"
	"time"
)

type StellarWithdrawDetails struct {
	TargetAddress string `json:"address"`
}

type Service struct {
	stellarRoot hProtocol.Root
	builder     xdrbuild.Builder
	asset       watchlist.Details

	stellarClient    horizonclient.ClientInterface
	txSubmitter      submit.Interface
	log              *logan.Entry
	stellarSource    hProtocol.Account
	paymentCfg       config.PaymentConfig
	withdrawCfg      config.WithdrawConfig
	withdrawStreamer getters.CreateWithdrawRequestHandler
}

func New(opts Opts) *Service {
	return &Service{
		log:              opts.Log,
		stellarSource:    opts.StellarSource,
		stellarClient:    opts.StellarClient,
		paymentCfg:       opts.PaymentConfig,
		withdrawCfg:      opts.WithdrawConfig,
		builder:          opts.Builder,
		txSubmitter:      opts.TXSubmitter,
		stellarRoot:      opts.StellarRoot,
		asset:            opts.AssetDetails,
		withdrawStreamer: opts.Streamer,
	}
}

type Opts struct {
	StellarSource  hProtocol.Account
	StellarRoot    hProtocol.Root
	PaymentConfig  config.PaymentConfig
	AssetDetails   watchlist.Details
	WithdrawConfig config.WithdrawConfig
	StellarClient  horizonclient.ClientInterface
	Builder        xdrbuild.Builder
	TXSubmitter    submit.Interface
	Log            *logan.Entry
	Streamer       getters.CreateWithdrawRequestHandler
}

func (s *Service) Run(ctx context.Context) {
	s.prepare()

	withdrawPage := &regources.ReviewableRequestListResponse{}
	var err error
	running.WithBackOff(ctx, s.log, "withdraw-processor", func(ctx context.Context) error {
		if len(withdrawPage.Data) < requestPageSizeLimit {
			withdrawPage, err = s.withdrawStreamer.List()
		} else {
			withdrawPage, err = s.withdrawStreamer.Next()
		}
		if err != nil {
			return errors.Wrap(err, "error occurred while withdrawal request page fetching")
		}
		for _, data := range withdrawPage.Data {
			details := withdrawPage.Included.MustCreateWithdrawRequest(data.Relationships.RequestDetails.Data.GetKey())
			err := s.processWithdraw(ctx, data, details)
			if err != nil {
				s.log.
					WithError(err).
					WithField("details", details).
					Warn("failed to process withdraw request")
				continue
			}
			s.log.WithField("request_id", details.ID).Debug("Successfully processed withdraw")
		}
		return nil
	}, 10*time.Second, 10*time.Second, 5*time.Minute)
}
