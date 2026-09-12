package model

// MonitorDeskCounts عدّاد حي لكل طوابير مكتب المراقب الخمسة — رقم
// واحد بمكان واحد لكل زر، بدل ما يفتح كل تبويب ويشوف بنفسه.
type MonitorDeskCounts struct {
	Inbox    int `db:"inbox" json:"inbox"`
	Issues   int `db:"issues" json:"issues"`
	Invoices int `db:"invoices" json:"invoices"`
	Quality  int `db:"quality" json:"quality"`
	Crew     int `db:"crew" json:"crew"`
}
