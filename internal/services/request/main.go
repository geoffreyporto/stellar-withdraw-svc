package request

import (
	"context"
	"fmt"
	"github.com/tokend/stellar-withdraw-svc/internal/horizon/getters"
	"github.com/tokend/stellar-withdraw-svc/internal/horizon/page"
	"github.com/tokend/stellar-withdraw-svc/internal/horizon/query"
	"github.com/tokend/stellar-withdraw-svc/internal/services/watchlist"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
	"gitlab.com/distributed_lab/running"
	regources "gitlab.com/tokend/regources/generated"
	"time"
)

type Details struct {
	regources.ReviewableRequest
	*regources.CreateWithdrawRequest
}

type StellarWithdrawDetails struct {
	TargetAddress string `json:"address"`
}

type Service struct {
	asset            watchlist.Details
	withdrawStreamer getters.CreateWithdrawRequestHandler
	log              *logan.Entry
	reviewer         string
	ch               chan Details
}

func New(opts Opts) *Service {
	return &Service{
		asset:            opts.AssetDetails,
		withdrawStreamer: opts.WithdrawalStreamer,
		log:              opts.Log,
		reviewer:         opts.Reviewer,
		ch:               make(chan Details),
	}
}

type Opts struct {
	AssetDetails       watchlist.Details
	WithdrawalStreamer getters.CreateWithdrawRequestHandler
	Reviewer           string
	Log                *logan.Entry
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

func (s *Service) GetCh() <-chan Details {
	return s.ch
}

func (s *Service) Run(ctx context.Context) {
	defer close(s.ch)

	s.prepare()

	withdrawPage, err := s.withdrawStreamer.List()
	if err != nil {
		s.log.WithError(err).Fatal("error occured while withdrawal request fetching")
	}

	running.WithBackOff(ctx, s.log, "withdraw-processor", func(ctx context.Context) error {
		for _, data := range withdrawPage.Data {
			details := withdrawPage.Included.MustCreateWithdrawRequest(data.Relationships.RequestDetails.Data.GetKey())

			s.ch <- Details{
				ReviewableRequest:     data,
				CreateWithdrawRequest: details,
			}

		}
		if len(withdrawPage.Data) < requestPageSizeLimit {
			withdrawPage, err = s.withdrawStreamer.List()
		} else {
			withdrawPage, err = s.withdrawStreamer.Next()
		}
		if err != nil {
			return errors.Wrap(err, "error occurred while withdrawal request page fetching")
		}
		return nil
	}, 10*time.Second, 10*time.Second, 5*time.Minute)
}
