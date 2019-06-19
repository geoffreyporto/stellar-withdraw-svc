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
	streamer  getters.AssetHandler
	log       *logan.Entry
	owner     string
	watchlist map[string]bool
	toAdd     chan Details
	toRemove  chan string
}

type Opts struct {
	Streamer   getters.AssetHandler
	Log        *logan.Entry
	AssetOwner string
}

func New(opts Opts) *Service {
	ch := make(chan Details)
	return &Service{
		streamer:  opts.Streamer,
		owner:     opts.AssetOwner,
		log:       opts.Log.WithField("service", "watchlist"),
		toAdd:     ch,
		watchlist: make(map[string]bool),
	}
}

func (s *Service) GetToAdd() <-chan Details {
	return s.toAdd
}

func (s *Service) GetToRemove() <-chan string {
	return s.toRemove
}

func (s *Service) Run(ctx context.Context) {
	defer close(s.toAdd)
	defer close(s.toRemove)

	// TODO: It is better to use addrstate here later
	running.WithBackOff(
		ctx,
		s.log,
		"asset-watcher",
		s.processAllAssetsOnce,
		time.Minute,
		time.Minute,
		time.Hour,
	)
}

func (s *Service) processAllAssetsOnce(ctx context.Context) error {
	active := make(map[string]bool)
	assetsToWatch, err := s.getWatchList()
	if err != nil {
		return errors.Wrap(err, "failed to get asset watch list")
	}
	for _, asset := range assetsToWatch {
		s.toAdd <- asset
		active[asset.ID] = true
	}

	for asset := range s.watchlist {
		if _, ok := active[asset]; !ok {
			s.toRemove <- asset
		}
	}

	s.watchlist = active
	return nil
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

		if !assetDetails.Stellar.Withdraw {
			continue
		}
		if err = assetDetails.Validate(); err != nil {
			s.log.WithError(err).Warn("incorrect asset details")
			continue
		}

		result = append(result, Details{
			Asset:        asset,
			AssetDetails: assetDetails,
		})
	}

	return result, nil
}
