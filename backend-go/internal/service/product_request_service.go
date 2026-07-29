package service

import (
	"fmt"
	"strings"

	"github.com/google/uuid"

	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

// ProductRequestService يدير "إدارة المنتجات" (وحدة التقنيين) — اقتراح منتج
// جديد يُضاف لكتالوج النظام (اسم/مواصفات/مصدر/موديل/تصنيف/سعر)، يفتحه المدير
// أو التقني أو مسؤول المشتريات، ويوافق/يرفض المدير حصراً.
type ProductRequestService struct {
	repo *repository.ProductRequestRepository
}

func NewProductRequestService(repo *repository.ProductRequestRepository) *ProductRequestService {
	return &ProductRequestService{repo: repo}
}

func (s *ProductRequestService) List() ([]model.ProductRequest, error) {
	return s.repo.List()
}

func (s *ProductRequestService) Create(req model.CreateProductProposalRequest, requestedByID string) (*model.ProductRequest, error) {
	if strings.TrimSpace(req.ProductName) == "" {
		return nil, fmt.Errorf("اسم المنتج مطلوب")
	}
	return s.repo.Create(uuid.NewString(), req, requestedByID)
}

func (s *ProductRequestService) Approve(id, resolvedByID string) (*model.ProductRequest, error) {
	return s.repo.Resolve(id, model.ProductRequestApproved, resolvedByID)
}

func (s *ProductRequestService) Reject(id, resolvedByID string) (*model.ProductRequest, error) {
	return s.repo.Resolve(id, model.ProductRequestRejected, resolvedByID)
}
