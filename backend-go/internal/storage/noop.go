package storage

import (
	"context"
	"errors"
)

// UnavailableStore باكند بديل لما التخزين ما يتهيّأ (مجلد مو قابل
// للكتابة، صلاحيات ناقصة...).
//
// ليش موجود؟ لأن فشل تهيئة التخزين كان يقتل السيرفر كله عند الإقلاع.
// مشكلة بمجلد ملفات ما يصير توقّع الحضور والحجوزات والحسابات — الرفع
// والعرض بس ينعطلون، برسالة واضحة، وباقي النظام يشتغل عادي.
type UnavailableStore struct{ reason string }

func NewUnavailableStore(reason string) *UnavailableStore {
	return &UnavailableStore{reason: reason}
}

func (s *UnavailableStore) Kind() string {
	return "معطّل — " + s.reason
}

func (s *UnavailableStore) err() error {
	return errors.New("تخزين الملفات مو مهيّأ: " + s.reason)
}

func (s *UnavailableStore) Put(context.Context, string, []byte, string) error { return s.err() }

func (s *UnavailableStore) Get(context.Context, string) ([]byte, string, error) {
	return nil, "", s.err()
}

func (s *UnavailableStore) Delete(context.Context, string) error { return s.err() }
