package service

import (
	"context"
	"errors"
	"regexp"
	"strings"

	"github.com/jason127vip-dot/Go-Sales/model"
	"github.com/jason127vip-dot/Go-Sales/repository"
)

var ErrInvalidBranch = errors.New("branch code must contain 1-20 letters or numbers, branch name is required, and total credit limit must be a non-negative amount with no more than two decimal places")
var branchCodePattern = regexp.MustCompile(`^[A-Z0-9]+$`)

type BranchService struct{ repository *repository.BranchRepository }

func NewBranchService(r *repository.BranchRepository) *BranchService {
	return &BranchService{repository: r}
}
func (s *BranchService) FindAll(ctx context.Context) ([]model.Branch, error) {
	return s.repository.FindAll(ctx)
}
func (s *BranchService) Save(ctx context.Context, row *model.Branch) error {
	row.Code, row.Name = strings.ToUpper(strings.TrimSpace(row.Code)), strings.TrimSpace(row.Name)
	if !branchCodePattern.MatchString(row.Code) || len(row.Code) > 20 || row.Name == "" || len([]rune(row.Name)) > 200 || !validCreditAmount(row.TotalCreditLimit) {
		return ErrInvalidBranch
	}
	return s.repository.Save(ctx, row)
}
