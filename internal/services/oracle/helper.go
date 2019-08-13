package oracle

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/stellar/go/clients/horizonclient"
	hProtocol "github.com/stellar/go/protocols/horizon"
	"github.com/stellar/go/txnbuild"
	"github.com/tokend/stellar-withdraw-svc/internal/horizon/page"
	"github.com/tokend/stellar-withdraw-svc/internal/horizon/query"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
	regources "gitlab.com/tokend/regources/generated"
)

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

func (s *Service) processWithdraw(ctx context.Context, request regources.ReviewableRequest, details *regources.CreateWithdrawRequest) error {
	detailsbb := []byte(details.Attributes.CreatorDetails)
	withdrawDetails := StellarWithdrawDetails{}
	err := json.Unmarshal(detailsbb, &withdrawDetails)
	if err != nil {
		s.log.WithField("request_id", request.ID).WithError(err).Warn("Unable to unmarshal creator details")
		return s.permanentReject(ctx, request, invalidDetails)
	}

	if withdrawDetails.TargetAddress == "" {
		s.log.
			WithField("creator_details", details.Attributes.CreatorDetails).
			WithError(err).
			Warn("address missing")
		return s.permanentReject(ctx, request, invalidTargetAddress)
	}

	if !s.proveExternalAccountExists(withdrawDetails.TargetAddress) {
		return s.permanentReject(ctx, request, noExtAccount)
	}

	err = s.approveRequest(ctx, request, taskApproveSuccessfulTxSend, taskTrySendToStellar, map[string]interface{}{})
	if err != nil {
		return errors.Wrap(err, "failed to review request first time", logan.F{"request_id": request.ID})
	}

	txSuccess, err := s.submitPayment(request.ID, withdrawDetails.TargetAddress, details.Attributes.Amount.String())
	if err != nil {
		return s.permanentReject(ctx, request, stellarTxFailed)
	}

	err = s.approveRequest(ctx, request, 0, taskApproveSuccessfulTxSend, map[string]interface{}{
		"stellar_tx_hash": txSuccess.Hash,
	})

	if err != nil {
		return errors.Wrap(err, "failed to review request second time", logan.F{"request_id": request.ID})
	}

	return nil
}

func (s *Service) getAsset() txnbuild.Asset {
	if s.asset.Stellar.AssetType == string(horizonclient.AssetTypeNative) {
		return txnbuild.NativeAsset{}
	}

	return txnbuild.CreditAsset{
		Issuer: s.asset.Stellar.Issuer,
		Code:   s.asset.Stellar.Code,
	}
}

func (s *Service) proveExternalAccountExists(accountID string) bool {
	_, err := s.stellarClient.AccountDetail(horizonclient.AccountRequest{
		AccountID: accountID,
	})
	if err != nil {
		s.log.WithField("account_id", accountID).Debug("failed to prove existence of target account")
		return false
	}
	return true

}

func (s *Service) submitPayment(
	id string,
	targetAddress string,
	amount string,
) (*hProtocol.TransactionSuccess, error) {

	asset := s.getAsset()

	tx := txnbuild.Transaction{
		SourceAccount: &s.stellarSource,
		Memo:          txnbuild.MemoText(id),
		BaseFee:       s.paymentCfg.MaxBaseFee,
		Operations: []txnbuild.Operation{
			&txnbuild.Payment{
				Destination: targetAddress,
				Asset:       asset,
				Amount:      amount,
			}},
		Timebounds: txnbuild.NewTimeout(60),
		Network:    s.stellarRoot.NetworkPassphrase,
	}
	envelope, err := tx.BuildSignEncode(s.paymentCfg.SourceSigner)
	if err != nil {
		return nil, errors.Wrap(err, "failed to build and sign stellar tx envelope")
	}
	success, err := s.stellarClient.SubmitTransactionXDR(envelope)
	if err != nil {
		return nil, errors.Wrap(err, "failed ot submit transaction to stellar network")
	}

	return &success, nil
}
