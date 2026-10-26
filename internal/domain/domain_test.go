package domain

import (
	"bytes"
	"encoding/json/v2"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"testing"
	"time"
)

// contract/status-cases.json is shared with the web client (web/src/workers): both sides
// must derive the same status and stages from the same facts.
type contractCase struct {
	Name   string       `json:"name"`
	Facts  Worker       `json:"facts"`
	Status Status       `json:"status"`
	Stages []StageState `json:"stages"`
}

func TestStatusAndStagesMatchTheSharedContract(t *testing.T) {
	data, err := os.ReadFile("../../contract/status-cases.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []contractCase
	if err := json.Unmarshal(data, &cases); err != nil {
		t.Fatal(err)
	}
	if len(cases) < 15 {
		t.Fatalf("the contract should cover every status, got %d cases", len(cases))
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			if got := c.Facts.Status(); got != c.Status {
				t.Errorf("status: got %s, want %s", got, c.Status)
			}
			got := c.Facts.StageStates()
			if fmt.Sprint(got) != fmt.Sprint(c.Stages) {
				t.Errorf("stages: got %v, want %v", got, c.Stages)
			}
		})
	}
}

func TestEveryStatusHasALabelAndOnlySomeNeedAttention(t *testing.T) {
	all := []Status{
		StatusWorking, StatusNeedsYou, StatusPreviewLive, StatusWaitingOnClient, StatusClientFeedback,
		StatusGateFailed, StatusReadyToShip, StatusShipping, StatusLive, StatusRolledBack,
	}
	attention := 0
	for _, s := range all {
		if s.Label() == string(s) {
			t.Errorf("%s has no label", s)
		}
		if s.NeedsAttention() {
			attention++
		}
	}
	if attention != 4 {
		t.Errorf("Needs you, Client feedback, Gate failed and Rolled back wait on the developer; got %d", attention)
	}
	if got := Status("unknown").Label(); got != "unknown" {
		t.Errorf("an unknown status keeps its name, got %q", got)
	}
}

func TestStagesAreSevenAndInOrder(t *testing.T) {
	want := "Prompt Worktree Preview Client Gate Live Watch"
	var names []string
	for _, s := range Stages() {
		names = append(names, string(s))
	}
	if strings.Join(names, " ") != want {
		t.Fatalf("got %v", names)
	}
}

func TestANewWorkerStartsWorkingInItsWorktree(t *testing.T) {
	w := NewWorker(NewProjectID(), "Fix checkout total", AgentCodex, "fix/checkout-total", true, time.UnixMilli(1000))
	if w.Status() != StatusWorking || w.StageStates()[1] != StageRunning {
		t.Fatalf("status %s, stages %v", w.Status(), w.StageStates())
	}
	if !validSessionName(string(w.ID)) {
		t.Fatalf("worker id %q cannot name an agent session", w.ID)
	}
}

// A worker id is the name of its agent session (internal/pty): lowercase, digits, hyphens.
func validSessionName(s string) bool {
	if s == "" || len(s) > 40 || s[0] == '-' {
		return false
	}
	for _, c := range s {
		if c != '-' && (c < '0' || c > '9') && (c < 'a' || c > 'z') {
			return false
		}
	}
	return true
}

func TestIDsAreUniqueAndSortByCreation(t *testing.T) {
	first, second := NewWorkerID(), NewWorkerID()
	if first == second || first >= second {
		t.Fatalf("ids should be unique and increasing: %s then %s", first, second)
	}
	if NewWorkspaceID() == "" || NewProjectID() == "" {
		t.Fatal("empty id")
	}
}

func TestWorkspaceSlugsFitAUserNameAndAFolder(t *testing.T) {
	for _, ok := range []string{"acme", "warung-sari", "a1", "client-2026"} {
		if !ValidSlug(ok) {
			t.Errorf("%q should be valid", ok)
		}
	}
	for _, bad := range []string{"", "a", "-acme", "acme-", "Acme", "ac me", "../x", "acme_x", strings.Repeat("a", 21)} {
		if ValidSlug(bad) {
			t.Errorf("%q should be rejected", bad)
		}
	}
	if user := (Workspace{Slug: strings.Repeat("a", 20)}).AgentUser(); len(user) > 32 {
		t.Errorf("agent user %q is longer than Linux allows", user)
	}
}

func TestNamesThatBecomeFoldersOrUnits(t *testing.T) {
	for _, ok := range []string{"web", "site.v2", "api_server"} {
		if !ValidProjectName(ok) {
			t.Errorf("project name %q should be valid", ok)
		}
	}
	if id := string(NewWorkerID()); !ValidSessionID(id) {
		t.Errorf("worker id %q must be usable as a session id", id)
	}
	for _, bad := range []string{"", ".git", "..", "-x", "Web", "a/b", "a b", strings.Repeat("a", 64)} {
		if ValidProjectName(bad) {
			t.Errorf("project name %q should be rejected", bad)
		}
	}
	for _, bad := range []string{"", "-x", "S1", "a_b", "a.b", "s1;id", strings.Repeat("a", 41)} {
		if ValidSessionID(bad) {
			t.Errorf("session id %q should be rejected", bad)
		}
	}
}

func TestASecretNeverPrintsItsValue(t *testing.T) {
	s := NewSecret("cf-token-1234567890")
	var logged bytes.Buffer
	slog.New(slog.NewTextHandler(&logged, nil)).Info("saved", "token", s)
	encoded, err := json.Marshal(struct {
		Token Secret `json:"token"`
	}{s})
	if err != nil {
		t.Fatal(err)
	}
	outputs := []string{
		s.String(), fmt.Sprintf("%v %s %+v %#v", s, s, s, s), logged.String(), string(encoded),
		fmt.Errorf("could not use %v: %w", s, errors.New("denied")).Error(),
	}
	for _, out := range outputs {
		if strings.Contains(out, "cf-token") || !strings.Contains(out, "redacted") {
			t.Errorf("a secret leaked or was not redacted: %q", out)
		}
	}
	if s.Reveal() != "cf-token-1234567890" || !s.IsSet() || (Secret{}).IsSet() {
		t.Error("Reveal and IsSet must work")
	}
	if s.Last4() != "7890" || NewSecret("short").Last4() != "" {
		t.Errorf("Last4 shows only the end of a long secret, got %q", s.Last4())
	}
}

func TestWorkerJSONMatchesTheWebClientAndHidesTheServerPath(t *testing.T) {
	w := NewWorker("p1", "Title", AgentClaude, "feat/x", false, time.UnixMilli(1700000000000))
	w.Worktree = "/var/lib/astra/agents/acme/work/x"
	w.Gate = GateFact{State: GateFailed, Cause: "lint"}
	data, err := json.Marshal(w)
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	for _, want := range []string{`"createdAt":1700000000000`, `"agentActivity":"working"`, `"clientReview":false`, `"gate":{"state":"failed","cause":"lint"}`, `"deployment":{"state":"none"}`} {
		if !strings.Contains(text, want) {
			t.Errorf("JSON lacks %s:\n%s", want, text)
		}
	}
	if strings.Contains(text, "/var/lib") || strings.Contains(text, "worktree") {
		t.Errorf("the server path must not reach clients:\n%s", text)
	}
}
