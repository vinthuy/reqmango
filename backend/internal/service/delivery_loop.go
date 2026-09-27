package service

import (
	"math"
	"sort"
	"strconv"
	"time"
)

// Loop stage keys, in delivery order.
const (
	LoopStageTriage  = "triage"
	LoopStageQueue   = "queue"
	LoopStageDevelop = "develop"
	LoopStageRelease = "release"
)

var loopStages = []string{LoopStageTriage, LoopStageQueue, LoopStageDevelop, LoopStageRelease}

type LoopDuration struct {
	Key         string   `json:"key"`
	Samples     int      `json:"samples"`
	MedianHours *float64 `json:"median_hours"`
	P85Hours    *float64 `json:"p85_hours"`
	AvgHours    *float64 `json:"avg_hours"`
}

type LoopStage struct {
	LoopDuration
	WIP            int      `json:"wip"`
	WIPMedianHours *float64 `json:"wip_median_hours"`
}

type LoopFunnelStep struct {
	Key      string  `json:"key"`
	Count    int     `json:"count"`
	Rate     float64 `json:"rate"`
	StepRate float64 `json:"step_rate"`
}

type LoopStalledItem struct {
	IssueID   uint64    `json:"issue_id"`
	ProjectID uint64    `json:"project_id"`
	Key       string    `json:"key"`
	Name      string    `json:"name"`
	Stage     string    `json:"stage"`
	Since     time.Time `json:"since"`
	AgeHours  float64   `json:"age_hours"`
}

type LoopSummary struct {
	Received     int      `json:"received"`
	FromIntake   int      `json:"from_intake"`
	Accepted     int      `json:"accepted"`
	Rejected     int      `json:"rejected"`
	Done         int      `json:"done"`
	Shipped      int      `json:"shipped"`
	ShipRate     float64  `json:"ship_rate"`
	SpecCoverage *float64 `json:"spec_coverage"`
}

type DeliveryLoop struct {
	Days        int               `json:"days"`
	Summary     LoopSummary       `json:"summary"`
	Funnel      []LoopFunnelStep  `json:"funnel"`
	Stages      []LoopStage       `json:"stages"`
	Bottleneck  string            `json:"bottleneck"`
	PRReview    LoopDuration      `json:"pr_review"`
	LeadToDone  LoopDuration      `json:"lead_to_done"`
	LeadToShip  LoopDuration      `json:"lead_to_ship"`
	Stalled     []LoopStalledItem `json:"stalled"`
	GeneratedAt time.Time         `json:"generated_at"`
}

// loopRow is one requirement's observed timeline.
type loopRow struct {
	ID                 uint64
	ProjectID          uint64
	Identifier         string
	SequenceID         int64
	Name               string
	CreatedAt          time.Time
	IntakeSource       *string
	IntakeStatus       *string
	IntakeTriagedAt    *time.Time
	CompletedAt        *time.Time
	StateGroup         string
	StartedAt          *time.Time
	LinkedAt           *time.Time
	MergedAt           *time.Time
	ShippedAt          *time.Time
	HasSpec            bool
	ProjectHasReleases bool
}

func (r loopRow) fromIntake() bool { return r.IntakeSource != nil && *r.IntakeSource != "" }

func (r loopRow) intakeStatus() string {
	if r.IntakeStatus == nil {
		return ""
	}
	return *r.IntakeStatus
}

func (r loopRow) accepted() bool {
	return !r.fromIntake() || r.intakeStatus() == IntakeAccepted
}

func (r loopRow) acceptedAt() *time.Time {
	if !r.fromIntake() {
		t := r.CreatedAt
		return &t
	}
	if r.intakeStatus() != IntakeAccepted {
		return nil
	}
	if r.IntakeTriagedAt != nil {
		return r.IntakeTriagedAt
	}
	t := r.CreatedAt
	return &t
}

func (r loopRow) done() bool { return r.StateGroup == "completed" }

func (r loopRow) started() bool {
	return r.StartedAt != nil || r.StateGroup == "started" || r.done()
}

// Loop reports the requirement-to-release loop for requirements raised in the last `days` days.
func (s *WorkspaceAnalyticsService) Loop(workspaceID, projectID uint64, days int) (*DeliveryLoop, error) {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	since := today.AddDate(0, 0, -(days - 1))

	projectCond := ""
	args := []interface{}{workspaceID, since}
	if projectID > 0 {
		projectCond = " AND i.project_id = ?"
		args = append(args, projectID)
	}
	var rows []loopRow
	err := s.db.Raw(`
WITH c AS (
	SELECT i.id, i.project_id, p.identifier, i.sequence_id, i.name, i.created_at,
		i.intake_source, i.intake_status, i.intake_triaged_at, i.completed_at, s."group" AS state_group,
		EXISTS (SELECT 1 FROM releases r WHERE r.project_id = i.project_id AND r.deleted_at IS NULL AND r.status = 'released') AS project_has_releases
	FROM issues i
	JOIN states s ON s.id = i.state_id
	JOIN projects p ON p.id = i.project_id AND p.deleted_at IS NULL AND p.archived_at IS NULL
	WHERE i.workspace_id = ? AND i.deleted_at IS NULL AND i.archived_at IS NULL AND i.is_draft = false
		AND i.created_at >= ?`+projectCond+`
),
st AS (
	SELECT a.issue_id, MIN(a.created_at) AS started_at
	FROM issue_activities a
	JOIN c ON c.id = a.issue_id
	JOIN states x ON x.project_id = c.project_id AND x.name = a.new_value AND x."group" = 'started' AND x.deleted_at IS NULL
	WHERE a.field = 'state' AND a.deleted_at IS NULL
	GROUP BY a.issue_id
),
gl AS (
	SELECT g.issue_id, MIN(g.created_at) AS linked_at
	FROM git_issue_links g JOIN c ON c.id = g.issue_id
	GROUP BY g.issue_id
),
gm AS (
	SELECT a.issue_id, MIN(a.created_at) AS merged_at
	FROM issue_activities a JOIN c ON c.id = a.issue_id
	WHERE a.verb = 'git_merged' AND a.deleted_at IS NULL
	GROUP BY a.issue_id
),
rs AS (
	SELECT ri.issue_id, MIN(COALESCE(r.release_date, r.updated_at)) AS shipped_at
	FROM release_issues ri
	JOIN c ON c.id = ri.issue_id
	JOIN releases r ON r.id = ri.release_id AND r.deleted_at IS NULL AND r.status = 'released'
	GROUP BY ri.issue_id
),
sp AS (
	SELECT DISTINCT pg.source_id AS issue_id
	FROM pages pg JOIN c ON c.id = pg.source_id
	WHERE pg.source_type = 'issue' AND pg.deleted_at IS NULL
)
SELECT c.*, st.started_at, gl.linked_at, gm.merged_at, rs.shipped_at, (sp.issue_id IS NOT NULL) AS has_spec
FROM c
LEFT JOIN st ON st.issue_id = c.id
LEFT JOIN gl ON gl.issue_id = c.id
LEFT JOIN gm ON gm.issue_id = c.id
LEFT JOIN rs ON rs.issue_id = c.id
LEFT JOIN sp ON sp.issue_id = c.id`, args...).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := buildDeliveryLoop(rows, now)
	out.Days = days
	return out, nil
}

// buildDeliveryLoop aggregates per-requirement timelines. Durations that come out
// negative (e.g. a release dated before the work finished) are dropped rather than
// clamped so they cannot drag medians towards zero.
func buildDeliveryLoop(rows []loopRow, now time.Time) *DeliveryLoop {
	out := &DeliveryLoop{GeneratedAt: now, Stalled: []LoopStalledItem{}}

	var nAccepted, nStarted, nLinked, nDone, nShipped, nSpec int
	samples := map[string][]float64{}
	wipAges := map[string][]float64{}
	var review, leadDone, leadShip []float64
	add := func(dst *[]float64, from, to *time.Time) {
		if from == nil || to == nil {
			return
		}
		if h := to.Sub(*from).Hours(); h >= 0 {
			*dst = append(*dst, h)
		}
	}
	addStage := func(stage string, from, to *time.Time) {
		v := samples[stage]
		add(&v, from, to)
		samples[stage] = v
	}

	for i := range rows {
		r := rows[i]
		created := r.CreatedAt
		out.Summary.Received++
		if r.fromIntake() {
			out.Summary.FromIntake++
			if st := r.intakeStatus(); st == IntakeRejected || st == IntakeDuplicate {
				out.Summary.Rejected++
			}
			if r.IntakeTriagedAt != nil {
				addStage(LoopStageTriage, &created, r.IntakeTriagedAt)
			}
		}
		if !r.accepted() {
			if st := r.intakeStatus(); st != IntakeRejected && st != IntakeDuplicate {
				out.trackWIP(wipAges, r, LoopStageTriage, created, now)
			}
			continue
		}
		nAccepted++
		if r.HasSpec {
			nSpec++
		}
		acc := r.acceptedAt()
		if r.started() {
			nStarted++
		}
		if r.LinkedAt != nil {
			nLinked++
		}
		if r.done() {
			nDone++
		}
		if r.ShippedAt != nil {
			nShipped++
		}

		addStage(LoopStageQueue, acc, r.StartedAt)
		if r.done() {
			addStage(LoopStageDevelop, r.StartedAt, r.CompletedAt)
			addStage(LoopStageRelease, r.CompletedAt, r.ShippedAt)
			add(&leadDone, &created, r.CompletedAt)
		}
		add(&review, r.LinkedAt, r.MergedAt)
		add(&leadShip, &created, r.ShippedAt)

		switch {
		case r.ShippedAt != nil || r.StateGroup == "cancelled":
		case r.done():
			if r.ProjectHasReleases && r.CompletedAt != nil {
				out.trackWIP(wipAges, r, LoopStageRelease, *r.CompletedAt, now)
			}
		case r.started():
			since := *acc
			if r.StartedAt != nil {
				since = *r.StartedAt
			}
			out.trackWIP(wipAges, r, LoopStageDevelop, since, now)
		default:
			out.trackWIP(wipAges, r, LoopStageQueue, *acc, now)
		}
	}

	out.Summary.Accepted = nAccepted
	out.Summary.Done = nDone
	out.Summary.Shipped = nShipped
	if nAccepted > 0 {
		out.Summary.ShipRate = float64(nShipped) / float64(nAccepted)
		cov := float64(nSpec) / float64(nAccepted)
		out.Summary.SpecCoverage = &cov
	}

	steps := []struct {
		key string
		n   int
	}{
		{"received", out.Summary.Received}, {"accepted", nAccepted}, {"started", nStarted},
		{"pr_linked", nLinked}, {"done", nDone}, {"shipped", nShipped},
	}
	prev := out.Summary.Received
	for _, st := range steps {
		step := LoopFunnelStep{Key: st.key, Count: st.n}
		if out.Summary.Received > 0 {
			step.Rate = float64(st.n) / float64(out.Summary.Received)
		}
		if prev > 0 {
			step.StepRate = float64(st.n) / float64(prev)
		}
		out.Funnel = append(out.Funnel, step)
		if st.key != "pr_linked" {
			prev = st.n
		}
	}

	var worst float64
	for _, key := range loopStages {
		stage := LoopStage{LoopDuration: summarize(key, samples[key]), WIP: len(wipAges[key])}
		if len(wipAges[key]) > 0 {
			m := percentile(wipAges[key], 0.5)
			stage.WIPMedianHours = &m
		}
		// Work piling up in a stage counts as much as work that already passed through
		// it, otherwise a stage nothing has finished yet can never be the bottleneck.
		for _, h := range []*float64{stage.MedianHours, stage.WIPMedianHours} {
			if h != nil && *h > worst {
				worst = *h
				out.Bottleneck = key
			}
		}
		out.Stages = append(out.Stages, stage)
	}
	out.PRReview = summarize("pr_review", review)
	out.LeadToDone = summarize("lead_to_done", leadDone)
	out.LeadToShip = summarize("lead_to_ship", leadShip)

	sort.SliceStable(out.Stalled, func(i, j int) bool { return out.Stalled[i].AgeHours > out.Stalled[j].AgeHours })
	if len(out.Stalled) > 10 {
		out.Stalled = out.Stalled[:10]
	}
	return out
}

func (out *DeliveryLoop) trackWIP(ages map[string][]float64, r loopRow, stage string, since, now time.Time) {
	age := now.Sub(since).Hours()
	if age < 0 {
		age = 0
	}
	ages[stage] = append(ages[stage], age)
	out.Stalled = append(out.Stalled, LoopStalledItem{
		IssueID: r.ID, ProjectID: r.ProjectID, Key: r.Identifier + "-" + strconv.FormatInt(r.SequenceID, 10),
		Name: r.Name, Stage: stage, Since: since, AgeHours: math.Round(age*10) / 10,
	})
}

func summarize(key string, v []float64) LoopDuration {
	d := LoopDuration{Key: key, Samples: len(v)}
	if len(v) == 0 {
		return d
	}
	med, p85 := percentile(v, 0.5), percentile(v, 0.85)
	var sum float64
	for _, x := range v {
		sum += x
	}
	avg := sum / float64(len(v))
	d.MedianHours, d.P85Hours, d.AvgHours = &med, &p85, &avg
	return d
}

// percentile uses linear interpolation between closest ranks.
func percentile(v []float64, p float64) float64 {
	s := append([]float64(nil), v...)
	sort.Float64s(s)
	if len(s) == 1 {
		return s[0]
	}
	pos := p * float64(len(s)-1)
	lo := int(math.Floor(pos))
	hi := int(math.Ceil(pos))
	return s[lo] + (s[hi]-s[lo])*(pos-float64(lo))
}
