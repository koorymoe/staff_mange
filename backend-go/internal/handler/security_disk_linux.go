//go:build linux

package handler

import "syscall"

// diskUsage يقرأ مساحة القرص الفعلية — يشتغل بلينكس بس (وهذا سيرفر الإنتاج
// دايماً لينكس بحاوية Docker). المجلد "/" يمثل القرص كامل بالحاوية.
func diskUsage() (totalGB, usedGB, freeGB float64) {
	var stat syscall.Statfs_t
	if err := syscall.Statfs("/", &stat); err != nil {
		return 0, 0, 0
	}
	total := float64(stat.Blocks) * float64(stat.Bsize)
	free := float64(stat.Bfree) * float64(stat.Bsize)
	const gb = 1024 * 1024 * 1024
	return total / gb, (total - free) / gb, free / gb
}
