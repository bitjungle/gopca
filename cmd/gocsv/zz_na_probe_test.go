package main

import "testing"

func TestProbeNATokens(t *testing.T) {
	app := &App{}
	data, err := app.parseCSVContent("ID,Region,Score\nP1,Nord,1\nP2,NA,2\nP3,N/A,3\nP4,,4\n", ".csv")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	t.Logf("headers=%v", data.Headers)
	for i, row := range data.Data {
		t.Logf("row %d = %q", i, row)
	}
	t.Logf("categorical=%q", data.CategoricalColumns)
}
