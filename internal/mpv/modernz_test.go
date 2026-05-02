package mpv

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEnsureModernZAssetsCopiesScriptOptionsAndFont(t *testing.T) {
	sourceDir := t.TempDir()
	files := map[string]string{
		"modernz.lua":       "-- test lua\n",
		"modernz.conf":      "layout=modern\n",
		"modernz-icons.ttf": "test font",
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(sourceDir, name), []byte(content), 0o644); err != nil {
			t.Fatalf("write source asset: %v", err)
		}
	}

	t.Setenv(modernZEnvDir, sourceDir)

	assets, err := ensureModernZAssets()
	if err != nil {
		t.Fatalf("ensureModernZAssets returned error: %v", err)
	}

	expected := map[string]string{
		filepath.Join(assets.ConfigDir, "scripts", "modernz.lua"):      files["modernz.lua"],
		filepath.Join(assets.ConfigDir, "script-opts", "modernz.conf"): files["modernz.conf"],
		filepath.Join(assets.ConfigDir, "fonts", "modernz-icons.ttf"):  files["modernz-icons.ttf"],
		assets.ScriptPath: files["modernz.lua"],
	}
	for path, content := range expected {
		got, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read copied asset %s: %v", path, err)
		}
		if string(got) != content {
			t.Fatalf("expected copied asset %s to contain %q, got %q", path, content, string(got))
		}
	}
}
