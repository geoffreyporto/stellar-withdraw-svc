package watchlist

import (
	"context"
	"encoding/json"
	"github.com/tokend/stellar-withdraw-svc/internal/horizon/getters"
	"github.com/tokend/stellar-withdraw-svc/internal/horizon/query"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
	"gitlab.com/distributed_lab/running"
	"gitlab.com/tokend/regources/generated"
	"time"
)


type Service struct {
	streamer getters.AssetHandler
	log      *logan.Entry
	owner    string
	ch       chan Details
}

type Opts struct {
	Streamer   getters.AssetHandler
	Log        *logan.Entry
	AssetOwner string
}

func New(opts Opts) *Service {
	ch := make(chan Details)
	return &Service{
		streamer: opts.Streamer,
		owner:    opts.AssetOwner,
		log:      opts.Log.WithField("service", "watchlist"),
		ch:       ch,
	}
}

func (s *Service) GetChan() <-chan Details {
	return s.ch
}

func (s *Service) Run(ctx context.Context) {
	defer close(s.ch)

	running.WithBackOff(ctx, s.log, "asset-watcher", func(ctx context.Context) error {
		assetsToWatch, err := s.getWatchList()
		if err != nil {
			return errors.Wrap(err, "failed to get asset watch list")
		}
		for _, asset := range assetsToWatch {
			s.ch <- asset
		}
		return nil
	}, 10*time.Second, 20*time.Second, 5*time.Minute)
}

func (s *Service) getWatchList() ([]Details, error) {
	s.streamer.SetFilters(query.AssetFilters{Owner: &s.owner})

	assetsResponse, err := s.streamer.List()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get asset list for owner", logan.F{
			"owner_address": s.owner,
		})
	}

	watchList, err := s.filter(assetsResponse.Data)
	if err != nil {
		return nil, errors.Wrap(err, "failed to filter asset list")
	}

	links := assetsResponse.Links
	for len(assetsResponse.Data) > 0 {
		assetsResponse, err = s.streamer.Next()
		if err != nil {
			return nil, errors.Wrap(err, "failed to get next page of assetsResponse", logan.F{
				"links": links,
			})
		}

		links = assetsResponse.Links
		filtered, err := s.filter(assetsResponse.Data)
		if err != nil {
			return nil, errors.Wrap(err, "failed to filter asset list")
		}
		watchList = append(watchList, filtered...)
	}

	return watchList, nil
}

func (s *Service) filter(assets []regources.Asset) ([]Details, error) {
	result := make([]Details, 0, len(assets))
	for _, asset := range assets {
		details := asset.Attributes.Details
		assetDetails := AssetDetails{}
		err := json.Unmarshal([]byte(details), &assetDetails)
		if err != nil {
			return nil, errors.Wrap(err, "failed to unmarshal asset details", logan.F{
				"asset_code":    asset.ID,
				"asset_details": details,
			})
		}
		if err = assetDetails.Validate(); err != nil {
			return nil, errors.Wrap(err, "incorrect asset details")
		}

		if !assetDetails.Stellar.Withdraw {
			continue
		}

		result = append(result, Details{
			Asset:        asset,
			AssetDetails: assetDetails,
		})
	}

	return result, nil
}
