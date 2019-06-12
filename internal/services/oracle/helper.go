package oracle

import (
	"context"
	"encoding/json"
	"fmt"
	hProtocol "github.com/stellar/go/protocols/horizon"
	"github.com/stellar/go/txnbuild"
	"github.com/tokend/stellar-withdraw-svc/internal/horizon/query"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
	regources "gitlab.com/tokend/regources/generated"
)

func (s *Service) getFilters() query.CreateWithdrawRequestFilters {
	state := ReviewableRequestStatePending
	reviewer := s.withdrawCfg.Owner.Address()
	pendingTasks := fmt.Sprintf("%d", TaskWithdrawReadyToSendPayment)
	pendingTasksNotSet := fmt.Sprintf("%d", TaskWithdrawSending)
	return query.CreateWithdrawRequestFilters{
		Asset: &s.asset.ID,
		ReviewableRequestFilters: query.ReviewableRequestFilters{
			State:              &state,
			Reviewer:           &reviewer,
			PendingTasks:       &pendingTasks,
			PendingTasksNotSet: &pendingTasksNotSet,
		},
	}
}

func (s *Service) processWithdraw(ctx context.Context, request regources.ReviewableRequest, details *regources.CreateWithdrawRequest) error {
	detailsbb := []byte(details.Attributes.CreatorDetails)
	withdrawDetails := StellarWithdrawDetails{}
	err := json.Unmarshal(detailsbb, &withdrawDetails)
	if err != nil {
		s.log.WithField("request_id", request.ID).WithError(err).Warn("Unable to unmarshal creator details")
		return nil
	}
	if withdrawDetails.TargetAddress == "" {
		s.log.
			WithField("creator_details", details.Attributes.CreatorDetails).
			WithError(err).
			Warn("address missing")
		return nil
	}

	err = s.approveRequest(ctx, request, TaskWithdrawSending, TaskWithdrawReadyToSendPayment, map[string]interface{}{})
	if err != nil {
		return errors.Wrap(err, "failed to review request first time", logan.F{"request_id": request.ID})
	}

	txSuccess, err := s.submitPayment(request.ID, withdrawDetails.TargetAddress, details.Attributes.Amount.String())
	if err != nil {
		return errors.Wrap(err, "payment failed")
	}

	err = s.approveRequest(ctx, request, 0, TaskWithdrawSending, map[string]interface{}{
		"stellar_tx_hash": txSuccess.Hash,
	})

	if err != nil {
		return errors.Wrap(err, "failed to review request second time", logan.F{"request_id": request.ID})
	}

	return nil
}

func (s *Service) getAsset() txnbuild.Asset {
	if s.asset.StellarDetails.AssetType == "native" {
		return txnbuild.NativeAsset{}
	}

	return txnbuild.CreditAsset{
		Issuer: s.asset.StellarDetails.Issuer,
		Code:   s.asset.StellarDetails.Code,
	}
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
		Operations: []txnbuild.Operation{
			&txnbuild.Payment{
				Destination: targetAddress,
				Asset:       asset,
				Amount:      amount,
			}},
		Timebounds: txnbuild. NewTimeout(60),
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
