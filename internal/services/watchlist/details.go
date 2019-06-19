package watchlist

import (
	. "github.com/go-ozzo/ozzo-validation"
	"github.com/stellar/go/clients/horizonclient"
	regources "gitlab.com/tokend/regources/generated"
)

var assetTypes = []interface{}{
	string(horizonclient.AssetTypeNative),
	string(horizonclient.AssetType4),
	string(horizonclient.AssetType12),
}

type Details struct {
	regources.Asset
	AssetDetails
}

type AssetDetails struct {
	Stellar struct {
		Withdraw  bool   `json:"withdraw"`
		AssetType string `json:"asset_type"`
		Code      string `json:"asset_code"`
		Issuer    string `json:"issuer"`
	} `json:"stellar"`
}

func (s AssetDetails) Validate() error {
	errs := Errors{
		"Withdraw":  Validate(&s.Stellar.Withdraw, Required),
		"AssetType": Validate(&s.Stellar.AssetType, Required, In(assetTypes...)),
	}
	if s.Stellar.AssetType == string(horizonclient.AssetType4) {
		errs["Code"] = Validate(&s.Stellar.Code, Required, Length(1, 4))
		errs["Issuer"] = Validate(&s.Stellar.Issuer, Required)
	}

	if s.Stellar.AssetType == string(horizonclient.AssetType12) {
		errs["Code"] = Validate(&s.Stellar.Code, Required, Length(5, 12))
		errs["Issuer"] = Validate(&s.Stellar.Issuer, Required)
	}

	return errs.Filter()
}
