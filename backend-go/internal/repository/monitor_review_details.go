package repository

import (
	"log"

	"github.com/lib/pq"

	"staffmange-api/internal/model"
)

// hydrateDetails (ع 10-16): «من تجيني فاتورة تجيني كل تفاصيلها — سعرها،
// والليدر جاب هلكد، والمحاسب غيّر السعر… والحجز وتفاصيله: منو سجّله، منو
// تواصل، الموعد، صباحي لو مسائي، مراحل العمل». بنفس حارس الصندوق — المراقب
// ما ينادي شاشات ثانية (حارسها غير، و٤٠٣ = مخالفة).
func (r *MonitorReviewRepository) hydrateDetails(rows []model.MonitorReview) {
	invIDs, adjIDs, bookIDs := []string{}, []string{}, []string{}
	for _, x := range rows {
		switch x.EntityType {
		case "LEADER_INVOICE":
			invIDs = append(invIDs, x.EntityID)
		case "INVOICE_ADJUSTMENT":
			adjIDs = append(adjIDs, x.EntityID)
		}
		if x.Identity != nil && x.Identity.BookingID != "" {
			bookIDs = append(bookIDs, x.Identity.BookingID)
		} else if x.EntityType == "BOOKING" {
			bookIDs = append(bookIDs, x.EntityID)
		}
	}
	// التعديل → فاتورته
	adjInv := map[string]string{}
	if len(adjIDs) > 0 {
		pairs := []struct {
			ID  string `db:"id"`
			Inv string `db:"invoiceId"`
		}{}
		if err := r.db.Select(&pairs, `SELECT id, "invoiceId" FROM "LeaderInvoiceAdjustment" WHERE id = ANY($1)`, pq.Array(adjIDs)); err != nil {
			log.Printf("monitor details adj: %v", err)
		}
		for _, p := range pairs {
			adjInv[p.ID] = p.Inv
			invIDs = append(invIDs, p.Inv)
		}
	}
	invoices := map[string]*model.MonitorInvoiceDetail{}
	if len(invIDs) > 0 {
		list := []model.MonitorInvoiceDetail{}
		if err := r.db.Select(&list, `
			SELECT li.id, e.name AS "leaderName", li."executionCost", li."materialsTotal", li."discountValue",
			       li."netTotal", li."isFree", li.status::text AS status, b."amountCollected"::float8 AS collected,
			       NULLIF(btrim(li."externalInvoiceNumber"), '') AS "externalNo", li."createdAt"
			FROM "LeaderInvoice" li
			LEFT JOIN "Employee" e ON e.id = li."employeeId"
			LEFT JOIN "Booking" b ON b.id = li."bookingId"
			WHERE li.id = ANY($1)`, pq.Array(invIDs)); err != nil {
			log.Printf("monitor details invoice: %v", err)
		}
		for i := range list {
			d := list[i]
			d.Adjustments = []model.MonitorInvoiceAdjustment{}
			if err := r.db.Select(&d.Adjustments, `
				SELECT a."oldNetTotal", a."newNetTotal", a."oldExecutionCost", a."newExecutionCost",
				       a."oldMaterialsTotal", a."newMaterialsTotal", a.reason, e.name AS "byName", a."createdAt"
				FROM "LeaderInvoiceAdjustment" a LEFT JOIN "Employee" e ON e.id = a."adjustedById"
				WHERE a."invoiceId" = $1 ORDER BY a."createdAt"`, d.InvoiceID); err != nil {
				log.Printf("monitor details adjustments: %v", err)
			}
			invoices[d.InvoiceID] = &d
		}
	}
	bookings := map[string]*model.MonitorBookingDetail{}
	if len(bookIDs) > 0 {
		list := []struct {
			ID string `db:"id"`
			model.MonitorBookingDetail
		}{}
		if err := r.db.Select(&list, `
			SELECT b.id, b.status::text AS status, `+stationCaseSQL()+` AS station, s.name AS service,
			       ce.name AS "createdByName", b."createdAt", b."confirmationContactedAt" AS "contactedAt",
			       b."confirmedAt", COALESCE(b."confirmedByName", fe.name) AS "confirmedByName",
			       b."scheduledAt", b.shift::text AS shift, b."arrivedAt", b."startedAt", b."completedAt",
			       le.name AS "leaderName",
			       (SELECT string_agg(e.name, '، ' ORDER BY e.name) FROM booking_crew_ids(b.id) c(eid)
			          JOIN "Employee" e ON e.id = c.eid) AS crew,
			       b."amountCollected"::float8 AS collected,
			       (SELECT NULLIF(concat_ws(' · ', q.status::text, q."inspectionResult"::text), '')
			          FROM "QualityFollowUp" q WHERE q."bookingId" = b.id ORDER BY q."createdAt" DESC LIMIT 1) AS quality
			FROM "Booking" b
			LEFT JOIN "Service" s ON s.id = b."serviceId"
			LEFT JOIN "Employee" ce ON ce.id = b."createdById"
			LEFT JOIN "Employee" fe ON fe.id = b."confirmedByEmployeeId"
			LEFT JOIN "Employee" le ON le.id = booking_leader_id(b.id)
			WHERE b.id = ANY($1)`, pq.Array(bookIDs)); err != nil {
			log.Printf("monitor details booking: %v", err)
		}
		for i := range list {
			d := list[i].MonitorBookingDetail
			bookings[list[i].ID] = &d
		}
	}
	for i := range rows {
		x := &rows[i]
		d := &model.MonitorDetails{}
		switch x.EntityType {
		case "LEADER_INVOICE":
			d.Invoice = invoices[x.EntityID]
		case "INVOICE_ADJUSTMENT":
			d.Invoice = invoices[adjInv[x.EntityID]]
		}
		bid := x.EntityID
		if x.Identity != nil && x.Identity.BookingID != "" {
			bid = x.Identity.BookingID
		}
		d.Booking = bookings[bid]
		if d.Invoice != nil || d.Booking != nil {
			x.Details = d
		}
	}
}
