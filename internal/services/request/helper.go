package request

import (
	"fmt"
	"github.com/tokend/stellar-withdraw-svc/internal/horizon/query"
)

func (s *Service) getFilters() query.CreateWithdrawRequestFilters {
	state := reviewableRequestStatePending
	reviewer := s.reviewer
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
