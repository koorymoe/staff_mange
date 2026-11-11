package models

import (
	"sort"

	"github.com/jmoiron/sqlx"
)

// Introspect يقرا المخطط الحي من القاعدة بنفس شكل الموديلات — للمولّد وللمقارنة.
// جداول تتبّع التحديثات (SchemaMigration) مستثناة: النظام يديرها مو الموديل.
func Introspect(db *sqlx.DB) ([]Table, []Enum, error) {
	var tabs []string
	if err := db.Select(&tabs, `SELECT c.relname FROM pg_class c JOIN pg_namespace n ON n.oid = c.relnamespace
		WHERE n.nspname = 'public' AND c.relkind = 'r' AND c.relname <> 'SchemaMigration' ORDER BY c.relname`); err != nil {
		return nil, nil, err
	}
	cols := []struct {
		T, N, Ty, D string
		NN          bool
	}{}
	if err := db.Select(&cols, `SELECT c.relname AS t, a.attname AS n, format_type(a.atttypid, a.atttypmod) AS ty,
		a.attnotnull AS nn, COALESCE(pg_get_expr(d.adbin, d.adrelid), '') AS d
		FROM pg_attribute a JOIN pg_class c ON c.oid = a.attrelid JOIN pg_namespace n ON n.oid = c.relnamespace
		LEFT JOIN pg_attrdef d ON d.adrelid = a.attrelid AND d.adnum = a.attnum
		WHERE n.nspname = 'public' AND c.relkind = 'r' AND a.attnum > 0 AND NOT a.attisdropped
		ORDER BY c.relname, a.attnum`); err != nil {
		return nil, nil, err
	}
	cons := []struct{ T, N, D string }{}
	if err := db.Select(&cons, `SELECT c.relname AS t, k.conname AS n, pg_get_constraintdef(k.oid) AS d
		FROM pg_constraint k JOIN pg_class c ON c.oid = k.conrelid JOIN pg_namespace n ON n.oid = c.relnamespace
		WHERE n.nspname = 'public' AND c.relkind = 'r' ORDER BY c.relname, k.conname`); err != nil {
		return nil, nil, err
	}
	idx := []struct{ T, D string }{}
	if err := db.Select(&idx, `SELECT i.tablename AS t, i.indexdef AS d FROM pg_indexes i
		WHERE i.schemaname = 'public' AND NOT EXISTS (SELECT 1 FROM pg_constraint k WHERE k.conname = i.indexname)
		ORDER BY i.tablename, i.indexname`); err != nil {
		return nil, nil, err
	}
	m := map[string]*Table{}
	out := make([]Table, 0, len(tabs))
	for _, t := range tabs {
		out = append(out, Table{Name: t})
	}
	for i := range out {
		m[out[i].Name] = &out[i]
	}
	for _, c := range cols {
		if t := m[c.T]; t != nil {
			t.Columns = append(t.Columns, Column{Name: c.N, Type: c.Ty, NotNull: c.NN, Default: c.D})
		}
	}
	for _, c := range cons {
		if t := m[c.T]; t != nil {
			t.Constraints = append(t.Constraints, Constraint{Name: c.N, Def: c.D})
		}
	}
	for _, x := range idx {
		if t := m[x.T]; t != nil {
			t.Indexes = append(t.Indexes, x.D)
		}
	}
	ens := []struct{ N, V string }{}
	if err := db.Select(&ens, `SELECT t.typname AS n, e.enumlabel AS v FROM pg_type t
		JOIN pg_namespace n ON n.oid = t.typnamespace JOIN pg_enum e ON e.enumtypid = t.oid
		WHERE n.nspname = 'public' ORDER BY t.typname, e.enumsortorder`); err != nil {
		return nil, nil, err
	}
	em := map[string]*Enum{}
	for _, e := range ens {
		if em[e.N] == nil {
			em[e.N] = &Enum{Name: e.N}
		}
		em[e.N].Values = append(em[e.N].Values, e.V)
	}
	names := make([]string, 0, len(em))
	for n := range em {
		names = append(names, n)
	}
	sort.Strings(names)
	enumsOut := make([]Enum, 0, len(names))
	for _, n := range names {
		enumsOut = append(enumsOut, *em[n])
	}
	return out, enumsOut, nil
}

// IntrospectObjects الإضافات والتسلسلات المستقلة والدوال والـviews والـtriggers.
func IntrospectObjects(db *sqlx.DB) ([]Object, error) {
	out := []Object{}
	q := []struct{ kind, sql string }{
		{"ext", `SELECT extname AS key, 'CREATE EXTENSION IF NOT EXISTS ' || quote_ident(extname) AS def FROM pg_extension WHERE extname <> 'plpgsql'`},
		{"seq", `SELECT c.relname AS key, 'CREATE SEQUENCE IF NOT EXISTS ' || quote_ident(c.relname) AS def FROM pg_class c
			JOIN pg_namespace n ON n.oid = c.relnamespace WHERE n.nspname = 'public' AND c.relkind = 'S'
			AND NOT EXISTS (SELECT 1 FROM pg_depend d WHERE d.objid = c.oid AND d.deptype IN ('a', 'i'))`},
		{"fn", `SELECT p.proname || '(' || pg_get_function_identity_arguments(p.oid) || ')' AS key, pg_get_functiondef(p.oid) AS def
			FROM pg_proc p JOIN pg_namespace n ON n.oid = p.pronamespace WHERE n.nspname = 'public'
			AND NOT EXISTS (SELECT 1 FROM pg_depend d WHERE d.objid = p.oid AND d.deptype = 'e')`},
		{"view", `SELECT viewname AS key, 'CREATE OR REPLACE VIEW ' || quote_ident(viewname) || ' AS ' || definition AS def FROM pg_views WHERE schemaname = 'public'`},
		{"trg", `SELECT c.relname || '.' || t.tgname AS key, pg_get_triggerdef(t.oid) AS def FROM pg_trigger t
			JOIN pg_class c ON c.oid = t.tgrelid JOIN pg_namespace n ON n.oid = c.relnamespace
			WHERE n.nspname = 'public' AND NOT t.tgisinternal`},
	}
	for _, x := range q {
		rows := []struct{ Key, Def string }{}
		if err := db.Select(&rows, x.sql); err != nil {
			return nil, err
		}
		for _, r := range rows {
			out = append(out, Object{Kind: x.kind, Key: r.Key, Def: r.Def})
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Kind+out[i].Key < out[j].Kind+out[j].Key })
	return out, nil
}
