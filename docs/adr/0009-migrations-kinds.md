# 0009 — الترحيلات: مخطط أو بيانات، مو خليط
**التاريخ:** 2026-11-05 · **الحالة:** ✅ معتمد (٧ف)

## السياق
بالمجموع ٤٧٣ ترحيل مرقّم: ٣٩٧ مخطط بس، و31 بيانات بس (تصليح أو بذر)، و45 **خليط**. الخليط يصعّب المراجعة والاسترجاع: لو التصليح طلع غلط، ما نگدر نعيده بلا ما نعيد المخطط. وبعد، أرقام المطبّق منها مسجّلة بجدول `SchemaMigration`، فنقلها أو تعديلها يخلّيها تنعاد أو تختفي.

## القرار
- **القديمة (لحد 0441) ما تنلمس.** تبقى مثل ما هي، والفهرس تحت يبيّن أي وحدة بيها بيانات.
- **من 0442 وجاي:** كل ترحيل صنف واحد.
  - **مخطط:** CREATE/ALTER/DROP.
  - **بيانات:** UPDATE/INSERT/DELETE، واسمه يبدي بـ `data_`، مثلاً `0450_data_fix_x`.
  - التغيير الي يحتاج الاثنين ينقسم لترحيلين متتاليين.
- البذر الي ينعاد بكل تشغيل (`seed*` و`grant*` بـ `migrate.go`) يبقى بمكانه، لأنه مكتوب يتحمّل التكرار.

## كيف ينفرض
اختبار `TestNewMigrationsAreSingleKind` (`backend-go/internal/database/migrate_kinds_test.go`) يفشّل الـ CI إذا ترحيل جديد خلط الصنفين، أو جان بيانات واسمه ما يبدي بـ `data_`.

## الفهرس: القديمة الي بيها بيانات
**بيانات بس (31):** `0116_seed_team_inventory_tool_catalog`، `0153_split_suppliers_permission`، `0155_tool_requests_approve_permission`، `0163_unit_visibility_permissions`، `0169_tool_kit_role_scope`، `0175_leave_routes_by_shift`، `0189_execution_cost_permission`، `0187_fund_amount_permission`، `0194_complaint_contact_permission`، `0239_unlock_permissions`، `0242_scoped_permissions`، `0247_booking_visits_backfill`، `0254_sim_lock_scene_fix`، `0270_department_seed`، `0273_story_recipient_backfill`، `0274_leader_invoice_manual_backfill`، `0278_it_permissions`، `0287_service_kind_implies_paperwork`، `0310_monitor_stages_trim`، `0331_internal_paperwork`، `0334_quotation_system_to_manage_all`، `0335_report2_role_defaults`، `0346_clear_stale_waiting`، `0368_complaints_to_quality`، `0370_visit_crew_leader`، `0376_complaint_legacy_closed`، `0399_settle_old_imports`، `0420_project_stages_reorder`، `0427_ai_verdict_owner`، `0439_old_imports_audited`، `0441_inventory_shortage_autoclose`

**خليط (45):** `0125_add_material_wholesale_price`، `0145_create_project_work_type`، `0162_create_privacy_policy`، `0166_booking_multiple_services`، `0171_revolving_fund`، `0173_stock_intake_and_request_kind`، `0174_leave_requests`، `0184_internal_booking`، `0179_booking_delete_requests`، `0192_fund_discharge`، `0193_manual_procurement_detail`، `0198_discipline_start_date`، `0199_booking_schedule_window`، `0229_free_work_reason`، `0233_network_price_item`، `0242_booking_stage_tracking`، `0243_ai_core`، `0244_service_required_fields`، `0241_leave_kind_two_states`، `0244_review_kinds`، `0250_legacy_pending_reset`، `0253_sim_lab_config`، `0255_sim_device_geometry`، `0286_service_kind`، `0300_matrix_guide_rule`، `0313_project_stage_log`، `0319_media_briefs`، `0328_attendance_fix`، `0333_quotation_terms`، `0336_vehicle_tracking_import`، `0337_quotation_project_link`، `0342_procurement_decisions`، `0351_reapproved_invoices_fund_nonneg`، `0364_purchase_quotes`، `0367_supplier_relations`، `0371_booking_leader_fn`، `0374_complaint_legacy_settle`، `0394_ai_conversations_memory`، `0408_org_structure`، `0410_principles_acceptance`، `0413_work_regulations`، `0424_justification_project`، `0430_profile_survey_daily_report`، `0431_rating_duty_negligence`، `0435_audit_trail`
