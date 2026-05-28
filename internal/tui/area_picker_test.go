package tui

// T-3 — Unit + property-based tests for the areaPicker component (BL-8).
//
// Coverage map (design.md §2.6):
//   - CP-1  TestAreaPicker_EnterOnAreaSelects / TestProp_SelectAreaSetsID
//   - CP-2  TestAreaPicker_EnterOnNoAreaClears / TestProp_NoAreaClearsID
//   - CP-3  TestAreaPicker_EscCancels / TestProp_EscPreservesSelection
//   - CP-4  TestAreaPicker_OpensCursorOnCurrent / TestProp_CursorOpensOnCurrent
//   - CP-5  TestProp_CursorInBounds
//   - CP-6  TestAreaPicker_RowsLayout / TestProp_RowLayout
//   - CP-7  TestAreaPicker_NoAreaLabelContextual (targeted; fixed input space)
//   - CP-8  TestProp_CreateNewNameRoundTrip
//   - CP-9  TestAreaPicker_CreateEmptyNameErrors / TestProp_EmptyNameNoCreate
//   - CP-10 TestAreaPicker_CreateDuplicateSelectsExisting / TestProp_DuplicateNameNoDup
//   - CP-11 TestAreaPicker_ReadOnlyHidesCreate / TestProp_ReadOnlyNeverCreates
//   - CP-15 TestAreaPicker_CreateServiceErrorKeepsOpen / TestProp_CreateErrorKeepsOpen

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jtprogru/todushka/internal/app"
	"github.com/jtprogru/todushka/internal/domain/area"
	"github.com/jtprogru/todushka/internal/domain/id"
	"github.com/jtprogru/todushka/internal/storage/fakes"
	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"
)

// failingAreaRepo embeds the in-memory fake but forces AreaCreate to fail
// with a non-ErrAlreadyExists error — used to exercise CP-15.
type failingAreaRepo struct{ *fakes.InMemRepo }

func (failingAreaRepo) AreaCreate(context.Context, area.Area) error {
	return errors.New("simulated storage failure")
}

func mkAreas(t *testing.T, svc *app.Service, names ...string) []area.Area {
	t.Helper()
	ctx := context.Background()
	out := make([]area.Area, 0, len(names))
	for _, n := range names {
		a, err := svc.AddArea(ctx, n)
		require.NoError(t, err)
		out = append(out, a)
	}
	return out
}

// ─── Unit tests ──────────────────────────────────────────────────────────

func TestAreaPicker_OpensCursorOnCurrent(t *testing.T) { // CP-4
	svc := newEditorTestService(t)
	areas := mkAreas(t, svc, "a", "b", "c")
	cur := areas[1].ID
	p := newAreaPicker(areas, &cur, "Inbox", false)
	require.Equal(t, 2, p.cursor)
	require.Equal(t, 0, newAreaPicker(areas, nil, "Inbox", false).cursor)
}

func TestAreaPicker_NoAreaLabelContextual(t *testing.T) { // CP-7
	svc := newEditorTestService(t)
	areas := mkAreas(t, svc, "x")
	require.Equal(t, "Inbox", newAreaPicker(areas, nil, "Inbox", false).rowLabels()[0])
	require.Equal(t, "No area", newAreaPicker(areas, nil, "No area", false).rowLabels()[0])
}

func TestAreaPicker_RowsLayout(t *testing.T) { // CP-6
	svc := newEditorTestService(t)
	areas := mkAreas(t, svc, "a", "b")
	p := newAreaPicker(areas, nil, "Inbox", false)
	require.Equal(t, []string{"Inbox", "a", "b", createRowLabel}, p.rowLabels())
}

func TestAreaPicker_ReadOnlyHidesCreate(t *testing.T) { // CP-6 / CP-11
	svc := newEditorTestService(t)
	areas := mkAreas(t, svc, "a", "b")
	p := newAreaPicker(areas, nil, "Inbox", true)
	require.Equal(t, []string{"Inbox", "a", "b"}, p.rowLabels())
	require.NotContains(t, p.rowLabels(), createRowLabel)
}

func TestAreaPicker_EnterOnAreaSelects(t *testing.T) { // CP-1
	svc := newEditorTestService(t)
	areas := mkAreas(t, svc, "a", "b")
	p := newAreaPicker(areas, nil, "Inbox", false)
	p.cursor = 1
	_, res := p.Update(tea.KeyMsg{Type: tea.KeyEnter}, svc)
	require.Equal(t, pickerSelected, res.outcome)
	require.NotNil(t, res.areaID)
	require.Equal(t, areas[0].ID, *res.areaID)
	require.Equal(t, "a", res.areaName)
}

func TestAreaPicker_EnterOnNoAreaClears(t *testing.T) { // CP-2
	svc := newEditorTestService(t)
	areas := mkAreas(t, svc, "a")
	p := newAreaPicker(areas, &areas[0].ID, "Inbox", false)
	p.cursor = 0
	_, res := p.Update(tea.KeyMsg{Type: tea.KeyEnter}, svc)
	require.Equal(t, pickerSelected, res.outcome)
	require.Nil(t, res.areaID)
	require.Equal(t, "", res.areaName)
}

func TestAreaPicker_EscCancels(t *testing.T) { // CP-3
	svc := newEditorTestService(t)
	p := newAreaPicker(nil, nil, "Inbox", false)
	_, res := p.Update(tea.KeyMsg{Type: tea.KeyEsc}, svc)
	require.Equal(t, pickerCancel, res.outcome)
}

func TestAreaPicker_CreateEmptyNameErrors(t *testing.T) { // CP-9
	svc := newEditorTestService(t)
	p := newAreaPicker(nil, nil, "Inbox", false)
	p.cursor = p.lastRowIndex() // create row
	p, res := p.Update(tea.KeyMsg{Type: tea.KeyEnter}, svc)
	require.Equal(t, pickerNone, res.outcome)
	require.True(t, p.creating)
	before, _ := svc.ListAreas(context.Background())
	p.nameInput.SetValue("   ")
	p, res = p.Update(tea.KeyMsg{Type: tea.KeyEnter}, svc)
	require.Equal(t, pickerNone, res.outcome)
	require.NotEmpty(t, p.err)
	after, _ := svc.ListAreas(context.Background())
	require.Equal(t, len(before), len(after))
}

func TestAreaPicker_CreateDuplicateSelectsExisting(t *testing.T) { // CP-10
	svc := newEditorTestService(t)
	areas := mkAreas(t, svc, "Work")
	p := newAreaPicker(areas, nil, "Inbox", false)
	p.creating = true
	p.nameInput.SetValue("work") // normalized duplicate
	before, _ := svc.ListAreas(context.Background())
	_, res := p.Update(tea.KeyMsg{Type: tea.KeyEnter}, svc)
	require.Equal(t, pickerSelected, res.outcome)
	require.NotNil(t, res.areaID)
	require.Equal(t, areas[0].ID, *res.areaID)
	after, _ := svc.ListAreas(context.Background())
	require.Equal(t, len(before), len(after))
}

func TestAreaPicker_CreateServiceErrorKeepsOpen(t *testing.T) { // CP-15
	svc := app.New(failingAreaRepo{fakes.New()}, fixedClock{now: time.Date(2026, 5, 25, 10, 0, 0, 0, time.UTC)})
	p := newAreaPicker(nil, nil, "Inbox", false)
	p.creating = true
	p.nameInput.SetValue("newone")
	p, res := p.Update(tea.KeyMsg{Type: tea.KeyEnter}, svc)
	require.Equal(t, pickerNone, res.outcome)
	require.NotEmpty(t, p.err)
}

// ─── Property-based tests ─────────────────────────────────────────────────

func rapidAreas(rt *rapid.T, svc *app.Service, n int) []area.Area {
	areas := make([]area.Area, 0, n)
	for i := 0; i < n; i++ {
		a, err := svc.AddArea(context.Background(), fmt.Sprintf("area-%d", i))
		require.NoError(rt, err)
		areas = append(areas, a)
	}
	return areas
}

func TestProp_SelectAreaSetsID(t *testing.T) { // CP-1
	rapid.Check(t, func(rt *rapid.T) {
		_, svc := newTestModelWithServiceRapid(rt)
		n := rapid.IntRange(1, 6).Draw(rt, "n")
		areas := rapidAreas(rt, svc, n)
		idx := rapid.IntRange(0, n-1).Draw(rt, "idx")
		p := newAreaPicker(areas, nil, "Inbox", false)
		p.cursor = idx + 1
		_, res := p.Update(tea.KeyMsg{Type: tea.KeyEnter}, svc)
		require.Equal(rt, pickerSelected, res.outcome)
		require.NotNil(rt, res.areaID)
		require.Equal(rt, areas[idx].ID, *res.areaID)
	})
}

func TestProp_NoAreaClearsID(t *testing.T) { // CP-2
	rapid.Check(t, func(rt *rapid.T) {
		_, svc := newTestModelWithServiceRapid(rt)
		n := rapid.IntRange(0, 5).Draw(rt, "n")
		areas := rapidAreas(rt, svc, n)
		p := newAreaPicker(areas, nil, "Inbox", false)
		p.cursor = 0
		_, res := p.Update(tea.KeyMsg{Type: tea.KeyEnter}, svc)
		require.Equal(rt, pickerSelected, res.outcome)
		require.Nil(rt, res.areaID)
	})
}

func TestProp_EscPreservesSelection(t *testing.T) { // CP-3
	rapid.Check(t, func(rt *rapid.T) {
		_, svc := newTestModelWithServiceRapid(rt)
		n := rapid.IntRange(0, 5).Draw(rt, "n")
		areas := rapidAreas(rt, svc, n)
		var cur *id.ID
		if n > 0 && rapid.Bool().Draw(rt, "hasCur") {
			cur = &areas[rapid.IntRange(0, n-1).Draw(rt, "ci")].ID
		}
		p := newAreaPicker(areas, cur, "Inbox", false)
		for i, moves := 0, rapid.IntRange(0, 8).Draw(rt, "moves"); i < moves; i++ {
			p, _ = p.Update(tea.KeyMsg{Type: tea.KeyDown}, svc)
		}
		_, res := p.Update(tea.KeyMsg{Type: tea.KeyEsc}, svc)
		require.Equal(rt, pickerCancel, res.outcome)
	})
}

func TestProp_CursorOpensOnCurrent(t *testing.T) { // CP-4
	rapid.Check(t, func(rt *rapid.T) {
		_, svc := newTestModelWithServiceRapid(rt)
		n := rapid.IntRange(1, 6).Draw(rt, "n")
		areas := rapidAreas(rt, svc, n)
		if rapid.Bool().Draw(rt, "nilCur") {
			require.Equal(rt, 0, newAreaPicker(areas, nil, "Inbox", false).cursor)
			return
		}
		idx := rapid.IntRange(0, n-1).Draw(rt, "idx")
		cur := areas[idx].ID
		require.Equal(rt, idx+1, newAreaPicker(areas, &cur, "Inbox", false).cursor)
	})
}

func TestProp_CursorInBounds(t *testing.T) { // CP-5
	rapid.Check(t, func(rt *rapid.T) {
		_, svc := newTestModelWithServiceRapid(rt)
		n := rapid.IntRange(0, 5).Draw(rt, "n")
		areas := rapidAreas(rt, svc, n)
		readOnly := rapid.Bool().Draw(rt, "ro")
		p := newAreaPicker(areas, nil, "Inbox", readOnly)
		seq := rapid.SliceOfN(rapid.SampledFrom([]tea.KeyType{tea.KeyUp, tea.KeyDown}), 0, 20).Draw(rt, "seq")
		for _, kt := range seq {
			p, _ = p.Update(tea.KeyMsg{Type: kt}, svc)
			require.GreaterOrEqual(rt, p.cursor, 0)
			require.LessOrEqual(rt, p.cursor, p.lastRowIndex())
		}
	})
}

func TestProp_RowLayout(t *testing.T) { // CP-6
	rapid.Check(t, func(rt *rapid.T) {
		_, svc := newTestModelWithServiceRapid(rt)
		n := rapid.IntRange(0, 5).Draw(rt, "n")
		areas := rapidAreas(rt, svc, n)
		readOnly := rapid.Bool().Draw(rt, "ro")
		want := []string{"Inbox"}
		for _, a := range areas {
			want = append(want, a.Name)
		}
		if !readOnly {
			want = append(want, createRowLabel)
		}
		require.Equal(rt, want, newAreaPicker(areas, nil, "Inbox", readOnly).rowLabels())
	})
}

func TestProp_ReadOnlyNeverCreates(t *testing.T) { // CP-11
	rapid.Check(t, func(rt *rapid.T) {
		_, svc := newTestModelWithServiceRapid(rt)
		n := rapid.IntRange(0, 4).Draw(rt, "n")
		areas := rapidAreas(rt, svc, n)
		p := newAreaPicker(areas, nil, "Inbox", true)
		before, _ := svc.ListAreas(context.Background())
		keys := rapid.SliceOfN(rapid.SampledFrom([]tea.KeyType{tea.KeyUp, tea.KeyDown, tea.KeyEnter}), 0, 15).Draw(rt, "keys")
		for _, kt := range keys {
			p, _ = p.Update(tea.KeyMsg{Type: kt}, svc)
		}
		require.NotContains(rt, p.rowLabels(), createRowLabel)
		after, _ := svc.ListAreas(context.Background())
		require.Equal(rt, len(before), len(after))
	})
}

func TestProp_CreateNewNameRoundTrip(t *testing.T) { // CP-8
	rapid.Check(t, func(rt *rapid.T) {
		_, svc := newTestModelWithServiceRapid(rt)
		name := "new-" + rapid.StringMatching(`[a-z]{3,10}`).Draw(rt, "name")
		p := newAreaPicker(nil, nil, "Inbox", false)
		p.creating = true
		p.nameInput.SetValue(name)
		_, res := p.Update(tea.KeyMsg{Type: tea.KeyEnter}, svc)
		require.Equal(rt, pickerSelected, res.outcome)
		require.NotNil(rt, res.areaID)
		require.Equal(rt, name, res.areaName)
		got, err := svc.Repo().AreaGet(context.Background(), *res.areaID)
		require.NoError(rt, err)
		require.Equal(rt, name, got.Name)
	})
}

func TestProp_EmptyNameNoCreate(t *testing.T) { // CP-9
	rapid.Check(t, func(rt *rapid.T) {
		_, svc := newTestModelWithServiceRapid(rt)
		ws := rapid.StringMatching(` {0,5}`).Draw(rt, "ws")
		p := newAreaPicker(nil, nil, "Inbox", false)
		p.creating = true
		p.nameInput.SetValue(ws)
		before, _ := svc.ListAreas(context.Background())
		p, res := p.Update(tea.KeyMsg{Type: tea.KeyEnter}, svc)
		require.Equal(rt, pickerNone, res.outcome)
		require.NotEmpty(rt, p.err)
		after, _ := svc.ListAreas(context.Background())
		require.Equal(rt, len(before), len(after))
	})
}

func TestProp_DuplicateNameNoDup(t *testing.T) { // CP-10
	rapid.Check(t, func(rt *rapid.T) {
		_, svc := newTestModelWithServiceRapid(rt)
		base := rapid.StringMatching(`[a-z]{3,8}`).Draw(rt, "base")
		existing, err := svc.AddArea(context.Background(), base)
		require.NoError(rt, err)
		p := newAreaPicker([]area.Area{existing}, nil, "Inbox", false)
		p.creating = true
		p.nameInput.SetValue(strings.ToUpper(base))
		before, _ := svc.ListAreas(context.Background())
		_, res := p.Update(tea.KeyMsg{Type: tea.KeyEnter}, svc)
		require.Equal(rt, pickerSelected, res.outcome)
		require.NotNil(rt, res.areaID)
		require.Equal(rt, existing.ID, *res.areaID)
		after, _ := svc.ListAreas(context.Background())
		require.Equal(rt, len(before), len(after))
	})
}

func TestProp_CreateErrorKeepsOpen(t *testing.T) { // CP-15
	rapid.Check(t, func(rt *rapid.T) {
		svc := app.New(failingAreaRepo{fakes.New()}, fixedClock{now: time.Date(2026, 5, 25, 10, 0, 0, 0, time.UTC)})
		name := rapid.StringMatching(`[a-z]{3,8}`).Draw(rt, "name")
		p := newAreaPicker(nil, nil, "Inbox", false)
		p.creating = true
		p.nameInput.SetValue(name)
		p, res := p.Update(tea.KeyMsg{Type: tea.KeyEnter}, svc)
		require.Equal(rt, pickerNone, res.outcome)
		require.NotEmpty(rt, p.err)
	})
}
