package storage

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

// LocalStore يخزّن الملفات على قرص السيرفر. هذا الافتراضي: يشتغل بلا
// حساب خارجي ولا إعدادات، ويكفي تماماً لحجم شغل الشركة.
//
// نوع الملف ينحفظ بملف جنبي `.type` بدل ما نستنتجه من الامتداد —
// الاستنتاج يغلط مع الملفات الي انرفعت بلا امتداد.
type LocalStore struct {
	root string
}

func NewLocalStore(root string) (*LocalStore, error) {
	if root == "" {
		root = "data/uploads"
	}
	abs, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(abs, 0o750); err != nil {
		return nil, err
	}
	return &LocalStore{root: abs}, nil
}

func (s *LocalStore) Kind() string { return "قرص محلي (" + s.root + ")" }

// safePath يمنع الخروج من مجلد التخزين عبر ../ بالمفتاح.
func (s *LocalStore) safePath(key string) (string, error) {
	clean := filepath.Clean("/" + strings.TrimSpace(key))
	full := filepath.Join(s.root, clean)
	if !strings.HasPrefix(full, s.root+string(os.PathSeparator)) {
		return "", errors.New("مسار غير صالح")
	}
	return full, nil
}

func (s *LocalStore) Put(_ context.Context, key string, data []byte, contentType string) error {
	full, err := s.safePath(key)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(full), 0o750); err != nil {
		return err
	}
	if err := os.WriteFile(full, data, 0o640); err != nil {
		return err
	}
	return os.WriteFile(full+".type", []byte(contentType), 0o640)
}

func (s *LocalStore) Get(_ context.Context, key string) ([]byte, string, error) {
	full, err := s.safePath(key)
	if err != nil {
		return nil, "", err
	}
	data, err := os.ReadFile(full)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, "", ErrNotFound
		}
		return nil, "", err
	}
	contentType := "application/octet-stream"
	if ct, err := os.ReadFile(full + ".type"); err == nil && len(ct) > 0 {
		contentType = string(ct)
	}
	return data, contentType, nil
}

func (s *LocalStore) Delete(_ context.Context, key string) error {
	full, err := s.safePath(key)
	if err != nil {
		return err
	}
	_ = os.Remove(full + ".type")
	if err := os.Remove(full); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
