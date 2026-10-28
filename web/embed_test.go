package web

import (
	"strings"
	"testing"
)

func TestPublicFSEmbedsSiteScripts(t *testing.T) {
	for _, name := range []string{
		"public/js/alpine.min.js",
		"public/js/shop.js",
		"public/js/microlight.js",
		"public/js/highlight.js",
		"public/admin/index.html",
		"public/admin/js/app.js",
		"public/admin/js/pages/analytics.js",
		"public/admin/pages/analytics.html",
		"public/admin/js/pages/settings.js",
	} {
		data, err := PublicFS.ReadFile(name)
		if err != nil {
			t.Fatalf("read embedded %s: %v", name, err)
		}
		if len(data) == 0 {
			t.Fatalf("embedded %s is empty", name)
		}
	}

	alpine, err := PublicFS.ReadFile("public/js/alpine.min.js")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(alpine), `version:"3.17.5"`) {
		t.Fatal("embedded Alpine build is not 3.17.5")
	}
}
