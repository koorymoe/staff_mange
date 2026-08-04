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
	repo     *repository.ProductRequestRepository
	products *repository.ProductRepository
}

func NewProductRequestService(repo *repository.ProductRequestRepository, products *repository.ProductRepository) *ProductRequestService {
	return &ProductRequestService{repo: repo, products: products}
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

// Approve يوافق على الاقتراح *وينشئ المنتج فعلاً* بكتالوج النظام.
//
// قبل هذا، الموافقة كانت تقلب الحالة بس — التقني يقترح، المدير
// يوافق، وما ينضاف ولا منتج. المواصفات والمصدر والموديل والسعر
// كلهن ينتقلن من الاقتراح للمنتج، والتوفر يبقى الافتراضي والتصنيف
// يعدّله الموظف من إدارة المنتجات.
func (s *ProductRequestService) Approve(id, resolvedByID string) (*model.ProductRequest, error) {
	item, err := s.repo.Resolve(id, model.ProductRequestApproved, resolvedByID)
	if err != nil {
		return nil, err
	}
	if item != nil && s.products != nil {
		unit := "قطعة"
		if _, perr := s.products.Create(model.CreateProductRequest{
			Name:         item.ProductName,
			Unit:         &unit,
			DefaultPrice: item.Price,
			Specs:        item.Specs,
			Source:       item.Source,
			ModelName:    item.Model,
		}); perr != nil {
			// الاقتراح انوافق عليه فعلاً؛ فشل الإنشاء ما يلغي الموافقة،
			// بس لازم المستخدم يدري حتى يضيفه يدوي.
			return item, fmt.Errorf("انوافق على الطلب بس تعذر إنشاء المنتج بالكتالوج — أضفه يدوي من «إضافة منتج»")
		}
	}
	return item, nil
}

func (s *ProductRequestService) Reject(id, resolvedByID string) (*model.ProductRequest, error) {
	return s.repo.Resolve(id, model.ProductRequestRejected, resolvedByID)
}
