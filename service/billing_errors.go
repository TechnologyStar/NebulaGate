package service

import "fmt"

// ErrPlanExhausted indicates the plan allowance has been consumed.
type ErrPlanExhausted struct{
	AssignmentId int
	Metric string
	Remaining int64
	Needed int64
}

func (e *ErrPlanExhausted) Error() string {
	return fmt.Sprintf("plan exhausted (assignment=%d, metric=%s, remaining=%d, needed=%d)", e.AssignmentId, e.Metric, e.Remaining, e.Needed)
}

// ErrBalanceInsufficient indicates that the user's or token's quota is not enough.
type ErrBalanceInsufficient struct{
	UserId int
	TokenId int
	Remaining int
	Needed int64
}

func (e *ErrBalanceInsufficient) Error() string {
	return fmt.Sprintf("balance insufficient (user=%d, token=%d, remaining=%d, needed=%d)", e.UserId, e.TokenId, e.Remaining, e.Needed)
}

// ErrIdempotent indicates the request has already been processed.
type ErrIdempotent struct{
	RequestId string
}

func (e *ErrIdempotent) Error() string {
	return fmt.Sprintf("request already processed: %s", e.RequestId)
}
