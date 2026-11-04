// Package errlog: بند ٧ج — بدل ما نبلع خطأ كتابة ثانوية بـ `_, _ = db.Exec(...)`
// نسجّله بالـ log ويا مكانه، حتى ينشاف بالسجلات بدون ما نوكّف العملية الأصلية.
package errlog

import (
	"database/sql"
	"fmt"
	"log"
	"runtime"
)

// Exec يستلم نتيجة db.Exec / tx.Exec مباشرة: errlog.Exec(db.Exec(...)).
func Exec(_ sql.Result, err error) {
	if err != nil {
		report(err)
	}
}

// Err يسجّل خطأ أي عملية ثانوية: errlog.Err(doSomething()).
func Err(err error) {
	if err != nil {
		report(err)
	}
}

func report(err error) {
	where := "?"
	if _, file, line, ok := runtime.Caller(2); ok {
		where = fmt.Sprintf("%s:%d", short(file), line)
	}
	log.Printf("[swallowed-err] %s: %v", where, err)
}

func short(p string) string {
	n := 0
	for i := len(p) - 1; i >= 0; i-- {
		if p[i] == '/' {
			n++
			if n == 2 {
				return p[i+1:]
			}
		}
	}
	return p
}
