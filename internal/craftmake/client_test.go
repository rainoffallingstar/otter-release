package craftmake

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestExecuteParsesVersionedEnvelope(t *testing.T) {
	binaryPath := filepath.Join(t.TempDir(), "craftmake")
	script := `#!/bin/sh
printf '%s\n' '{"protocol_version":"otter.craftmake/v1","command":"status","ok":true,"run_id":"run-20260726T013245Z-kxqjrm","state_path":"/tmp/state.sqlite","data":{"status":"succeeded"}}'
`
	if err := os.WriteFile(binaryPath, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	result, err := Execute(context.Background(), Request{Binary: binaryPath, Command: CommandStatus})
	if err != nil {
		t.Fatal(err)
	}
	if result.Failed() || result.Envelope.RunID != "run-20260726T013245Z-kxqjrm" || result.Envelope.Command != "status" {
		t.Fatalf("unexpected Craftmake result: %#v", result)
	}
}

func TestExecuteRejectsIncompatibleProtocol(t *testing.T) {
	binaryPath := filepath.Join(t.TempDir(), "craftmake")
	script := `#!/bin/sh
printf '%s\n' '{"protocol_version":"otter.craftmake/v2","command":"status","ok":true}'
`
	if err := os.WriteFile(binaryPath, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	_, err := Execute(context.Background(), Request{Binary: binaryPath, Command: CommandStatus})
	if err == nil {
		t.Fatal("expected incompatible protocol to fail")
	}
}
