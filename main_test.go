package main

import (
	"bytes"
	"testing"

	"github.com/localshred/koolkit/pkg/plugin"
)

func TestDescribeContract(t *testing.T) {
	var out bytes.Buffer
	handled, code := plugin.Handle([]string{"--kk-describe"}, pluginMeta(), &out)
	if !handled || code != 0 {
		t.Fatalf("describe not handled: handled=%v code=%d", handled, code)
	}
	if out.Len() == 0 {
		t.Fatal("describe produced no output")
	}
}

func TestRequiresListsClaude(t *testing.T) {
	var out bytes.Buffer
	plugin.Handle([]string{"--kk-requires"}, pluginMeta(), &out)
	if got := out.String(); got != "claude\n" {
		t.Fatalf("requires = %q, want \"claude\\n\"", got)
	}
}
