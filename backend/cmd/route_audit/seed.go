package main

import (
	"crypto/rand"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type seedColumn struct {
	Name     string  `gorm:"column:name"`
	UDT      string  `gorm:"column:udt"`
	Nullable string  `gorm:"column:nullable"`
	Default  *string `gorm:"column:dflt"`
}

func (c seedColumn) required() bool { return c.Nullable == "NO" && c.Default == nil }

func (c seedColumn) isInt() bool { return c.UDT == "int2" || c.UDT == "int4" || c.UDT == "int8" }

// ownerColumns are filled even when nullable so victimRow can find the row.
var ownerColumns = map[string]bool{
	"workspace_id": true, "project_id": true, "issue_id": true,
	"page_id": true, "agent_id": true, "custom_field_id": true,
}

var userColumnBases = map[string]bool{
	"user": true, "created_by": true, "updated_by": true, "owner": true, "author": true,
	"assignee": true, "actor": true, "reporter": true, "invited_by": true, "lead": true,
	"reviewer": true, "triggered_by": true, "requested_by": true, "sender": true, "approver": true,
}

// seedRoutes inserts one victim-owned row for each path param that resolves
// to nothing, so the route can actually be probed.
func (a *auditor) seedRoutes(routes gin.RoutesInfo) {
	tried := map[string]bool{}
	seeded := 0
	for _, rt := range routes {
		if skip(rt.Path) {
			continue
		}
		for attempt := 0; attempt < 4; attempt++ {
			_, ok, missing := a.fill(rt.Path)
			if ok {
				break
			}
			name, prev, _ := strings.Cut(missing, "@")
			resolved := a.parentsFor(rt.Path, name)
			progress := false
			for _, table := range candidates(name, prev) {
				key := table + fmt.Sprint(resolved)
				if tried[key] || !a.db.Migrator().HasTable(table) {
					continue
				}
				tried[key] = true
				id, err := a.seedRow(table, resolved, 0)
				if err != nil {
					fmt.Printf("seed %-28s failed: %v\n", table, err)
					continue
				}
				fmt.Printf("seed %-28s id=%s for %s\n", table, id, rt.Path)
				seeded++
				progress = true
				break
			}
			if !progress {
				break
			}
		}
	}
	fmt.Printf("seeded %d rows\n", seeded)
}

// parentsFor resolves the params that precede stop in path.
func (a *auditor) parentsFor(path, stop string) map[string]uint64 {
	resolved := map[string]uint64{}
	segs := strings.Split(path, "/")
	for i, s := range segs {
		if !strings.HasPrefix(s, ":") {
			continue
		}
		if s[1:] == stop {
			break
		}
		a.resolve(s[1:], pathContext(segs, i), resolved)
	}
	return resolved
}

func (a *auditor) seedRow(table string, resolved map[string]uint64, depth int) (string, error) {
	if depth > 3 {
		return "", fmt.Errorf("foreign-key chain too deep")
	}
	var cols []seedColumn
	a.db.Raw(`SELECT column_name AS name, udt_name AS udt, is_nullable AS nullable, column_default AS dflt
		FROM information_schema.columns WHERE table_schema = current_schema() AND table_name = ?
		ORDER BY ordinal_position`, table).Scan(&cols)
	type fk struct {
		Col string `gorm:"column:col"`
		Ref string `gorm:"column:ref"`
	}
	var fks []fk
	a.db.Raw(`SELECT kcu.column_name AS col, ccu.table_name AS ref
		FROM information_schema.table_constraints tc
		JOIN information_schema.key_column_usage kcu ON kcu.constraint_name = tc.constraint_name AND kcu.table_schema = tc.table_schema
		JOIN information_schema.constraint_column_usage ccu ON ccu.constraint_name = tc.constraint_name AND ccu.table_schema = tc.table_schema
		WHERE tc.constraint_type = 'FOREIGN KEY' AND tc.table_schema = current_schema() AND tc.table_name = ?`, table).Scan(&fks)
	refs := map[string]string{}
	for _, f := range fks {
		refs[f.Col] = f.Ref
	}

	var names, marks []string
	var vals []interface{}
	for _, c := range cols {
		if c.Name == "id" && c.Default != nil {
			continue
		}
		if !c.required() && !ownerColumns[c.Name] {
			continue
		}
		v, err := a.seedValue(table, c, refs[c.Name], resolved, depth)
		if err != nil {
			return "", fmt.Errorf("%s: %w", c.Name, err)
		}
		if v == nil {
			continue
		}
		names = append(names, `"`+c.Name+`"`)
		marks = append(marks, "?")
		vals = append(vals, v)
	}
	q := fmt.Sprintf(`INSERT INTO %q DEFAULT VALUES RETURNING id`, table)
	if len(names) > 0 {
		q = fmt.Sprintf(`INSERT INTO %q (%s) VALUES (%s) RETURNING id`, table, strings.Join(names, ", "), strings.Join(marks, ", "))
	}
	var id string
	if err := a.db.Raw(strings.Replace(q, "RETURNING id", "RETURNING id::text", 1), vals...).Scan(&id).Error; err != nil {
		return "", err
	}
	if id == "" {
		return "", fmt.Errorf("insert returned no id")
	}
	return id, nil
}

// resolveText resolves params of tables keyed by a non-integer id.
func (a *auditor) resolveText(name, ctx string) (string, bool) {
	for _, t := range candidates(name, ctx) {
		var udt string
		a.db.Raw(`SELECT udt_name FROM information_schema.columns
			WHERE table_schema = current_schema() AND table_name = ? AND column_name = 'id'`, t).Scan(&udt)
		if udt == "" || strings.HasPrefix(udt, "int") || !a.db.Migrator().HasColumn(t, "workspace_id") {
			continue
		}
		var id string
		a.db.Raw(fmt.Sprintf(`SELECT id::text FROM %q WHERE workspace_id = ? ORDER BY 1 LIMIT 1`, t), a.ws).Scan(&id)
		if id != "" {
			return id, true
		}
	}
	return "", false
}

func (a *auditor) seedValue(table string, c seedColumn, ref string, resolved map[string]uint64, depth int) (interface{}, error) {
	switch c.Name {
	case "workspace_id":
		return a.ws, nil
	case "project_id":
		return pick(resolved, "projects", a.ids["project"]), nil
	case "issue_id":
		return pick(resolved, "issues", a.ids["issue"]), nil
	case "id":
		if c.isInt() {
			var n uint64
			a.db.Raw(fmt.Sprintf(`SELECT COALESCE(MAX(id), 0) + 1 FROM %q`, table)).Scan(&n)
			return n, nil
		}
		if c.UDT != "uuid" {
			b := make([]byte, 8)
			_, _ = rand.Read(b)
			return fmt.Sprintf("ra%x", b), nil
		}
	}
	if ref == "" && c.isInt() && strings.HasSuffix(c.Name, "_id") {
		base := strings.TrimSuffix(c.Name, "_id")
		if _, pid, ok := parentLink(base, resolved); ok {
			return pid, nil
		}
		ref = a.guessRef(base)
	}
	switch {
	case ref == "users":
		return a.insider, nil
	case ref == table:
		if c.required() {
			return nil, fmt.Errorf("required self reference")
		}
		return nil, nil
	case ref != "":
		if id, ok := resolved[ref]; ok {
			return id, nil
		}
		if id, ok := a.victimRow(ref, resolved); ok {
			return id, nil
		}
		if !c.required() {
			return nil, nil
		}
		id, err := a.seedRow(ref, resolved, depth+1)
		if err != nil {
			return nil, err
		}
		if n, perr := strconv.ParseUint(id, 10, 64); perr == nil {
			return n, nil
		}
		return id, nil
	}
	return a.literal(c), nil
}

func (a *auditor) guessRef(base string) string {
	if userColumnBases[base] || strings.HasSuffix(base, "_user") {
		return "users"
	}
	for _, t := range append(overrides[base+"s"], base+"s", base+"es", base) {
		if a.db.Migrator().HasTable(t) {
			return t
		}
	}
	return ""
}

var seedCounter int

func (a *auditor) literal(c seedColumn) interface{} {
	if strings.HasPrefix(c.UDT, "_") {
		return "{}"
	}
	switch c.UDT {
	case "int2", "int4", "int8", "numeric", "float4", "float8":
		return 0
	case "bool":
		return false
	case "timestamptz", "timestamp", "date":
		return time.Now()
	case "json", "jsonb":
		return "{}"
	case "uuid":
		b := make([]byte, 16)
		_, _ = rand.Read(b)
		return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
	case "bytea":
		return []byte{}
	}
	seedCounter++
	return fmt.Sprintf("ra%d", seedCounter)
}

func pick(m map[string]uint64, key string, fallback uint64) uint64 {
	if v, ok := m[key]; ok && v != 0 {
		return v
	}
	return fallback
}
