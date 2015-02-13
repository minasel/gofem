// Copyright 2012 Dorival Pedroso & Raul Durand. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mreten

import (
	"testing"

	"github.com/cpmech/gosl/plt"
	"github.com/cpmech/gosl/utl"
)

func Test_vg01(tst *testing.T) {

	prevTs := utl.Tsilent
	defer func() {
		utl.Tsilent = prevTs
		if err := recover(); err != nil {
			tst.Error("[1;31mERROR:", err, "[0m\n")
		}
	}()

	//utl.Tsilent = false
	utl.TTitle("vg01")

	mdl := GetModel("testsim", "mat1", "vg", false)
	mdl.Init(mdl.GetPrms(true))

	pc0 := -1.0
	sl0 := 1.0
	pcf := 3.0
	nptsA := 21
	nptsB := 11

	//doplot := true
	doplot := false
	if doplot {
		plt.Reset()
		Plot(mdl, pc0, sl0, pcf, nptsA, "'b.-'", "'r+-'", "bc")
	}

	tolCc := 1e-10
	tolD1a, tolD1b := 1e-10, 1e-17
	tolD2a, tolD2b := 1e-10, 1e-17
	Check(tst, mdl, pc0, sl0, pcf, nptsB, tolCc, tolD1a, tolD1b, tolD2a, tolD2b, true, []float64{}, 1e-7, doplot)

	if doplot {
		PlotEnd(true)
	}
}
