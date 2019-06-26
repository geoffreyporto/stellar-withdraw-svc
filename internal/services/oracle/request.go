package oracle

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/tokend/stellar-withdraw-svc/internal/horizon/query"
	"gitlab.com/distributed_lab/logan/v3/errors"
	"gitlab.com/tokend/go/xdr"
	"gitlab.com/tokend/go/xdrbuild"
	regources "gitlab.com/tokend/regources/generated"
	"strconv"
)

const (
	taskTrySendToStellar        uint32 = 2048
	taskApproveSuccessfulTxSend uint32 = 4096

	//Request state
	reviewableRequestStatePending = 1
	//page size
	requestPageSizeLimit = 10

	invalidDetails       = "Invalid creator details"
	invalidTargetAddress = "Invalid target address"
	noExtAccount         = "External account does not exist"
)

func (s *Service) approveRequest(
	ctx context.Context,
	request regources.ReviewableRequest,
	toAdd, toRemove uint32,
	extDetails map[string]interface{}) error {
	id, err := strconv.ParseUint(request.ID, 10, 64)
	if err != nil {
		return errors.Wrap(err, "failed to parse request id")
	}
	bb, err := json.Marshal(extDetails)
	if err != nil {
		return errors.Wrap(err, "failed to marshal external bb map")
	}
	envelope, err := s.builder.Transaction(s.withdrawCfg.Owner).Op(xdrbuild.ReviewRequest{
		ID:     id,
		Hash:   &request.Attributes.Hash,
		Action: xdr.ReviewRequestOpActionApprove,
		Details: xdrbuild.WithdrawalDetails{
			ExternalDetails: string(bb),
		},
		ReviewDetails: xdrbuild.ReviewDetails{
			TasksToAdd:      toAdd,
			TasksToRemove:   toRemove,
			ExternalDetails: string(bb),
		},
	}).Sign(s.withdrawCfg.Signer).Marshal()
	if err != nil {
		return errors.Wrap(err, "failed to prepare transaction envelope")
	}
	_, err = s.txSubmitter.Submit(ctx, envelope, true)
	if err != nil {
		return errors.Wrap(err, "failed to approve withdraw request")
	}

	return nil
}

func (s *Service) permanentReject(
	ctx context.Context,
	request regources.ReviewableRequest, reason string) error {
	id, err := strconv.ParseUint(request.ID, 10, 64)
	if err != nil {
		return errors.Wrap(err, "failed to parse request id")
	}
	envelope, err := s.builder.Transaction(s.withdrawCfg.Owner).Op(xdrbuild.ReviewRequest{
		ID:     id,
		Hash:   &request.Attributes.Hash,
		Action: xdr.ReviewRequestOpActionPermanentReject,
		Details: xdrbuild.WithdrawalDetails{},
	}).Sign(s.withdrawCfg.Signer).Marshal()
	if err != nil {
		return errors.Wrap(err, "failed to prepare transaction envelope")
	}
	_, err = s.txSubmitter.Submit(ctx, envelope, true)
	if err != nil {
		return errors.Wrap(err, "failed to approve withdraw request")
	}

	return nil
}


func (s *Service) getFilters() query.CreateWithdrawRequestFilters {
	state := reviewableRequestStatePending
	reviewer := s.withdrawCfg.Owner.Address()
	pendingTasks := fmt.Sprintf("%d", taskTrySendToStellar)
	pendingTasksNotSet := fmt.Sprintf("%d", taskApproveSuccessfulTxSend)
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
