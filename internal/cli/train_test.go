package cli

import (
	"bytes"
	"context"
	"path/filepath"
	"strings"
	"testing"

	"typer/internal/train"
)

func setupTestProfileStore(t *testing.T) {
	t.Helper()
	original := newProfileStore
	t.Cleanup(func() { newProfileStore = original })
	path := filepath.Join(t.TempDir(), "profile.json")
	newProfileStore = func() (*train.ProfileStore, error) {
		return train.NewProfileStoreAt(path), nil
	}
}

func TestTrainStatusNoProfile(t *testing.T) {
	setupTestProfileStore(t)
	home := t.TempDir()
	setTestUserDirs(t, home)

	var out bytes.Buffer
	if err := runTrainStatus(nil, &out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "No training profile") {
		t.Fatalf("output = %q", out.String())
	}
}

func TestTrainListNoProfile(t *testing.T) {
	setupTestProfileStore(t)
	home := t.TempDir()
	setTestUserDirs(t, home)

	var out bytes.Buffer
	if err := runTrainList(&out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "1.1") || !strings.Contains(out.String(), "No profile yet") {
		t.Fatalf("output = %q", out.String())
	}
}

func TestTrainNoProfilePromptsEvaluate(t *testing.T) {
	setupTestProfileStore(t)
	home := t.TempDir()
	setTestUserDirs(t, home)

	var out bytes.Buffer
	if err := runTrain(context.Background(), nil, strings.NewReader(""), &out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "typer train -e") {
		t.Fatalf("output = %q", out.String())
	}
}

func TestTrainStatusWithProfile(t *testing.T) {
	setupTestProfileStore(t)
	home := t.TempDir()
	setTestUserDirs(t, home)

	store, _ := newProfileStore()
	p := train.NewProfile(train.PlacementResult{
		NetWPM:         40,
		Accuracy:       93,
		AssignedTier:   train.TierBuilding,
		AssignedLesson: "2.4",
	})
	if err := store.Save(p); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	if err := runTrainStatus(nil, &out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "Building") || !strings.Contains(out.String(), "2.4") {
		t.Fatalf("output = %q", out.String())
	}
}

func TestTrainHelp(t *testing.T) {
	var out bytes.Buffer
	if err := runTrain(context.Background(), []string{"--help"}, strings.NewReader(""), &out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "--evaluate") {
		t.Fatalf("output = %q", out.String())
	}
}

func TestRootHelpIncludesTrain(t *testing.T) {
	var out bytes.Buffer
	printHelp(&out)
	if !strings.Contains(out.String(), "train") {
		t.Fatalf("root help missing train: %q", out.String())
	}
}
