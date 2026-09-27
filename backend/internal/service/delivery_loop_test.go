package service

import (
	"math"
	"testing"
	"time"
)

func strp(s string) *string { return &s }

func TestBuildDeliveryLoop(t *testing.T) {
	now := time.Date(2026, 9, 27, 12, 0, 0, 0, time.UTC)
	at := func(h float64) *time.Time {
		v := now.Add(-time.Duration(h * float64(time.Hour)))
		return &v
	}
	rows := []loopRow{
		// Direct issue: queued 10h, developed 20h, released 30h after done; delivered by AI.
		{ID: 1, Identifier: "APP", SequenceID: 1, CreatedAt: *at(100), StartedAt: at(90), CompletedAt: at(70),
			StateGroup: "completed", LinkedAt: at(85), MergedAt: at(75), ShippedAt: at(40), HasAI: true, ProjectHasReleases: true},
		// Accepted intake: triaged in 4h, still queued since then.
		{ID: 2, Identifier: "APP", SequenceID: 2, CreatedAt: *at(50), IntakeSource: strp("form"),
			IntakeStatus: strp(IntakeAccepted), IntakeTriagedAt: at(46), StateGroup: "backlog", HasSpec: true},
		// Pending intake: waiting for triage.
		{ID: 3, Identifier: "APP", SequenceID: 3, CreatedAt: *at(200), IntakeSource: strp("email"),
			IntakeStatus: strp(IntakePending), StateGroup: "backlog"},
		// Rejected intake: triaged in 2h, then out of the loop.
		{ID: 4, Identifier: "APP", SequenceID: 4, CreatedAt: *at(30), IntakeSource: strp("form"),
			IntakeStatus: strp(IntakeRejected), IntakeTriagedAt: at(28), StateGroup: "backlog"},
		// Done without a release in a project that ships releases: waits for release.
		{ID: 5, Identifier: "APP", SequenceID: 5, CreatedAt: *at(60), StartedAt: at(55), CompletedAt: at(20),
			StateGroup: "completed", ProjectHasReleases: true},
		// Release dated before completion: negative release duration must be dropped.
		{ID: 6, Identifier: "APP", SequenceID: 6, CreatedAt: *at(80), StartedAt: at(79), CompletedAt: at(10),
			StateGroup: "completed", ShippedAt: at(30)},
	}
	out := buildDeliveryLoop(rows, now)

	s := out.Summary
	if s.Received != 6 || s.FromIntake != 3 || s.Accepted != 4 || s.Rejected != 1 || s.Done != 3 || s.Shipped != 2 {
		t.Fatalf("summary = %+v", s)
	}
	if s.SpecCoverage == nil || *s.SpecCoverage != 0.25 {
		t.Fatalf("spec coverage = %v", s.SpecCoverage)
	}
	if s.DeliveredAI != 1 || s.DeliveredHuman != 2 {
		t.Fatalf("delivery split = ai:%d human:%d, want ai:1 human:2", s.DeliveredAI, s.DeliveredHuman)
	}

	wantFunnel := map[string]int{"received": 6, "accepted": 4, "started": 3, "pr_linked": 1, "done": 3, "shipped": 2}
	for _, f := range out.Funnel {
		if wantFunnel[f.Key] != f.Count {
			t.Errorf("funnel %s = %d, want %d", f.Key, f.Count, wantFunnel[f.Key])
		}
	}
	// pr_linked is a side step: done's step rate is relative to started.
	if f := out.Funnel[4]; f.Key != "done" || f.StepRate != 1 {
		t.Errorf("done step = %+v", f)
	}

	stage := map[string]LoopStage{}
	for _, st := range out.Stages {
		stage[st.Key] = st
	}
	if st := stage[LoopStageTriage]; st.Samples != 2 || *st.MedianHours != 3 || st.WIP != 1 {
		t.Errorf("triage = %+v", st)
	}
	if st := stage[LoopStageRelease]; st.Samples != 1 || *st.MedianHours != 30 || st.WIP != 1 {
		t.Errorf("release = %+v (negative duration must be dropped)", st)
	}
	if st := stage[LoopStageQueue]; st.WIP != 1 || st.Samples != 3 {
		t.Errorf("queue = %+v", st)
	}
	// The pending intake item has waited 200h, longer than any finished stage median.
	if out.Bottleneck != LoopStageTriage {
		t.Errorf("bottleneck = %q", out.Bottleneck)
	}
	if b := buildDeliveryLoop(rows[:1], now).Bottleneck; b != LoopStageRelease {
		t.Errorf("single finished row bottleneck = %q, want release (30h)", b)
	}
	if out.PRReview.Samples != 1 || *out.PRReview.MedianHours != 10 {
		t.Errorf("pr review = %+v", out.PRReview)
	}

	if len(out.Stalled) != 3 || out.Stalled[0].Key != "APP-3" || out.Stalled[0].Stage != LoopStageTriage {
		t.Fatalf("stalled = %+v", out.Stalled)
	}
}

func TestPercentile(t *testing.T) {
	v := []float64{10, 1, 4, 2, 3}
	if got := percentile(v, 0.5); got != 3 {
		t.Errorf("median = %v", got)
	}
	if got := percentile(v, 0.85); math.Abs(got-6.4) > 1e-9 {
		t.Errorf("p85 = %v", got)
	}
	if v[0] != 10 {
		t.Error("percentile must not reorder its input")
	}
}
