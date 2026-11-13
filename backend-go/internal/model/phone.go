package model

import "strings"

// LocalPhone يوحّد رقم الموبايل العراقي لصيغة 07xxxxxxxxx (11-13): +964 / 00964 /
// مسافات / أرقام عربية — كلها نفس الرقم. چان «+9647701234567» و«0770 123 4567»
// يسوون زبونين. رقم مو عراقي يرجع أرقامه بس، بلا تخمين.
func LocalPhone(p string) string {
	d := strings.Map(func(r rune) rune {
		switch {
		case r >= '0' && r <= '9':
			return r
		case r >= '٠' && r <= '٩':
			return '0' + (r - '٠')
		}
		return -1
	}, p)
	d = strings.TrimPrefix(d, "00")
	switch {
	case strings.HasPrefix(d, "964") && len(d) == 13:
		d = "0" + d[3:]
	case strings.HasPrefix(d, "7") && len(d) == 10:
		d = "0" + d
	}
	return d
}
