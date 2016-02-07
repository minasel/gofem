// Copyright 2015 Dorival Pedroso and Raul Durand. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fld

import (
	"testing"

	"github.com/cpmech/gosl/chk"
)

func Test_fld01(tst *testing.T) {

	//verbose()
	chk.PrintTitle("fld01")

	H := 10.0
	g := 10.0

	var water Model
	water.Init(water.GetPrms(true), H, g)

	var dryair Model
	dryair.Gas = true
	dryair.Init(dryair.GetPrms(true), H, g)

	if chk.Verbose {
		water.Plot("/tmp/gofem", "fig_fld01_water", 21)
		dryair.Plot("/tmp/gofem", "fig_fld01_dryair", 21)
	}
}