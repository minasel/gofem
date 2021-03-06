// Copyright 2016 The Gofem Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package conduct

import (
	"github.com/cpmech/gosl/chk"
	"github.com/cpmech/gosl/fun"
	"github.com/cpmech/gosl/fun/dbf"
)

// M1 implements the liquid-gas conductivity model # 1
type M1 struct {

	// parameters for liquid
	λ0l float64
	λ1l float64
	αl  float64
	βl  float64

	// parameters for gas
	λ0g float64
	λ1g float64
	αg  float64
	βg  float64

	// auxiliary functions
	klr fun.RefIncRL1
	kgr fun.RefIncRL1
}

// add model to factory
func init() {
	allocators["m1"] = func() Model { return new(M1) }
}

// GetPrms gets (an example) of parameters
func (o M1) GetPrms(example bool) dbf.Params {
	return dbf.Params{
		&dbf.P{N: "lam0l", V: 0.001},
		&dbf.P{N: "lam1l", V: 1.2},
		&dbf.P{N: "alpl", V: 0.01},
		&dbf.P{N: "betl", V: 10},
		&dbf.P{N: "lam0g", V: 2.0},
		&dbf.P{N: "lam1g", V: 0.001},
		&dbf.P{N: "alpg", V: 0.01},
		&dbf.P{N: "betg", V: 10},
	}
}

// Init initialises this structure
func (o *M1) Init(prms dbf.Params) (err error) {
	for _, p := range prms {
		switch p.N {
		case "lam0l":
			o.λ0l = p.V
		case "lam1l":
			o.λ1l = p.V
		case "alpl":
			o.αl = p.V
		case "betl":
			o.βl = p.V
		case "lam0g":
			o.λ0g = p.V
		case "lam1g":
			o.λ1g = p.V
		case "alpg":
			o.αg = p.V
		case "betg":
			o.βg = p.V
		default:
			return chk.Err("parameter named %q is incorrect\n", p.N)
		}
	}
	err = o.klr.Init(dbf.Params{
		&dbf.P{N: "lam0", V: o.λ0l},
		&dbf.P{N: "lam1", V: o.λ1l},
		&dbf.P{N: "alp", V: o.αl},
		&dbf.P{N: "bet", V: o.βl},
	})
	if err != nil {
		return
	}
	err = o.kgr.Init(dbf.Params{
		&dbf.P{N: "lam0", V: o.λ0g},
		&dbf.P{N: "lam1", V: o.λ1g},
		&dbf.P{N: "alp", V: o.αg},
		&dbf.P{N: "bet", V: o.βg},
	})
	return
}

// Klr returns klr
func (o M1) Klr(sl float64) float64 {
	return o.klr.F(sl, nil)
}

// Kgr returns kgr
func (o M1) Kgr(sg float64) float64 {
	return o.kgr.F(sg, nil)
}

// DklrDsl returns ∂klr/∂sl
func (o M1) DklrDsl(sl float64) float64 {
	return o.klr.G(sl, nil)
}

// DkgrDsl returns ∂kgr/∂sg
func (o M1) DkgrDsg(sg float64) float64 {
	return o.kgr.G(sg, nil)
}
