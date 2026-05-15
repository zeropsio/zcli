package version

import (
	"bytes"
	"testing"
)

func TestAssetNameFor(t *testing.T) {
	cases := []struct {
		goos, goarch string
		want         string
	}{
		{"linux", "amd64", "zcli-linux-amd64"},
		{"linux", "386", "zcli-linux-i386"},
		{"darwin", "amd64", "zcli-darwin-amd64"},
		{"darwin", "arm64", "zcli-darwin-arm64"},
		{"windows", "amd64", "zcli-win-x64.exe"},
		{"windows", "386", "zcli-win-x64.exe"}, // windows always maps to the same asset
	}
	for _, tc := range cases {
		if got := assetNameFor(tc.goos, tc.goarch); got != tc.want {
			t.Errorf("assetNameFor(%q, %q) = %q, want %q", tc.goos, tc.goarch, got, tc.want)
		}
	}
}

func TestParseChecksum(t *testing.T) {
	body := `abc123def4567890abcdef1234567890abcdef1234567890abcdef1234567890  zcli-linux-amd64
0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef *zcli-darwin-arm64
fedcba9876543210fedcba9876543210fedcba9876543210fedcba9876543210  checksums.txt
`
	t.Run("standard format", func(t *testing.T) {
		got, err := parseChecksum(body, "zcli-linux-amd64")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := []byte{0xab, 0xc1, 0x23, 0xde, 0xf4, 0x56, 0x78, 0x90}
		if !bytes.Equal(got[:8], want) {
			t.Errorf("got first 8 bytes %x, want %x", got[:8], want)
		}
	})

	t.Run("starred name", func(t *testing.T) {
		got, err := parseChecksum(body, "zcli-darwin-arm64")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(got) != 32 {
			t.Errorf("sha256 should be 32 bytes, got %d", len(got))
		}
	})

	t.Run("missing asset", func(t *testing.T) {
		_, err := parseChecksum(body, "zcli-darwin-amd64")
		if err == nil {
			t.Fatal("expected error for missing asset")
		}
	})

	t.Run("malformed lines skipped", func(t *testing.T) {
		bad := "garbage\n\nonefield\nabc123  zcli-linux-amd64\n"
		_, err := parseChecksum(bad, "zcli-linux-amd64")
		if err != nil {
			t.Errorf("expected to find asset past bad lines: %v", err)
		}
	})

	t.Run("invalid hex", func(t *testing.T) {
		bad := "zzznot-hex  zcli-linux-amd64\n"
		_, err := parseChecksum(bad, "zcli-linux-amd64")
		if err == nil {
			t.Fatal("expected error for invalid hex")
		}
	})
}

func TestPlanUpgradeRefusesPackageManager(t *testing.T) {
	savedChannel := channel
	t.Cleanup(func() { channel = savedChannel })

	for _, stamp := range []string{"npm", "brew", "nix"} {
		channel = stamp
		plan, err := PlanUpgrade(t.Context(), UpgradeOptions{TargetVersion: "v1.0.0"})
		if err == nil {
			t.Errorf("channel %q: expected refusal, got plan %+v", stamp, plan)
			continue
		}
	}
}
