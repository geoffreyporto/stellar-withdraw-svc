package oracle

import (
	"context"
	"fmt"
	"github.com/stellar/go/clients/horizonclient"
	hProtocol "github.com/stellar/go/protocols/horizon"
	"github.com/tokend/stellar-withdraw-svc/internal/config"
	"github.com/tokend/stellar-withdraw-svc/internal/horizon/getters"
	"github.com/tokend/stellar-withdraw-svc/internal/horizon/page"
	"github.com/tokend/stellar-withdraw-svc/internal/horizon/query"
	"github.com/tokend/stellar-withdraw-svc/internal/horizon/submit"
	"github.com/tokend/stellar-withdraw-svc/internal/services/watchlist"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
	"gitlab.com/distributed_lab/running"
	"gitlab.com/tokend/go/xdrbuild"
	"sync"
	"time"
)

type StellarWithdrawDetails struct {
	TargetAddress string `json:"address"`
}

type Service struct {
	stellarRoot      hProtocol.Root
	asset            watchlist.Details
	withdrawStreamer getters.CreateWithdrawRequestHandler
	builder          *xdrbuild.Builder
	stellarClient    horizonclient.ClientInterface
	txSubmitter      submit.Interface
	log              *logan.Entry
	stellarSource    hProtocol.Account
	withdrawCfg      config.WithdrawConfig
	paymentCfg       config.PaymentConfig
	wg               *sync.WaitGroup
}

func New(opts Opts) *Service {
	return &Service{
		asset:            opts.AssetDetails,
		withdrawStreamer: opts.WithdrawalStreamer,
		log:              opts.Log,
		stellarSource:    opts.StellarSource,
		withdrawCfg:      opts.WithdrawConfig,
		wg:               opts.WG,
		stellarClient:    opts.StellarClient,
		paymentCfg:       opts.PaymentConfig,
		builder:          opts.Builder,
		txSubmitter:      opts.TXSubmitter,
		stellarRoot:      opts.StellarRoot,
	}
}

type Opts struct {
	StellarSource      hProtocol.Account
	StellarRoot        hProtocol.Root
	AssetDetails       watchlist.Details
	WithdrawalStreamer getters.CreateWithdrawRequestHandler
	WithdrawConfig     config.WithdrawConfig
	PaymentConfig      config.PaymentConfig
	StellarClient      horizonclient.ClientInterface
	Builder            *xdrbuild.Builder
	TXSubmitter        submit.Interface
	Log                *logan.Entry
	WG                 *sync.WaitGroup
}

func (s *Service) prepare() {
	filters := s.getFilters()
	s.withdrawStreamer.SetFilters(filters)
	s.withdrawStreamer.SetIncludes(query.CreateWithdrawRequestIncludes{
		ReviewableRequestIncludes: query.ReviewableRequestIncludes{
			RequestDetails: true,
		},
	})
	limit := fmt.Sprintf("%d", requestPageSizeLimit)
	s.withdrawStreamer.SetPageParams(page.Params{Limit: &limit})
}

func (s *Service) Run(ctx context.Context) {
	defer s.wg.Done()

	s.prepare()

	withdrawPage, err := s.withdrawStreamer.List()
	if err != nil {
		s.log.WithError(err).Fatal("error occured while withdrawal request fetching")
	}

	running.WithBackOff(ctx, s.log, "withdraw-processor", func(ctx context.Context) error {
		for _, data := range withdrawPage.Data {
			details := withdrawPage.Included.MustCreateWithdrawRequest(data.Relationships.RequestDetails.Data.GetKey())
			err := s.processWithdraw(ctx, data, details)
			if err != nil {
				return errors.Wrap(err, "failed to process withdraw request", logan.F{
					"request_id": data.ID,
				})
			}
		}
		if len(withdrawPage.Data) < requestPageSizeLimit {
			withdrawPage, err = s.withdrawStreamer.List()
		} else {
			withdrawPage, err = s.withdrawStreamer.Next()
		}
		if err != nil {
			return errors.Wrap(err, "error occured while withdrawal request page fetching")
		}
		return nil
	}, 10*time.Second, 10*time.Second, 5*time.Minute)
}
