// Copyright 2016 The Gofem Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ana

import (
	"math"

	"github.com/cpmech/gosl/fun/dbf"
	"github.com/cpmech/gosl/la"
	"github.com/cpmech/gosl/num"
	"github.com/cpmech/gosl/utl"
)

// PressCylin implements the constant-stress solution to a simple
// linear elastic plane-strain problem -- Hill's solution
//
//               , - - ,
//           , '         ' ,
//         ,                 ,
//        ,      .-'''-.      ,
//       ,      / ↖ ↑ ↗ \      ,
//       ,     |  ← P →  |     ,
//       ,      \ ↙ ↓ ↘ /      ,
//        ,      `-...-'      ,
//         ,                 ,
//           ,            , '
//             ' - , ,  '
type PressCylin struct {

	// input
	a  float64 // Inner radius
	b  float64 // Outer radius
	E  float64 // Young's modulus
	ν  float64 // Poisson's coefficient
	σy float64 // Uniaxial yield stress

	// derived data
	coef float64 // TODO
	Y    float64 // TODO
	P0   float64 // Pressure at the elastic/plastic transition (TODO: check this)
	Plim float64 // limiting pressure

	// auxiliary
	Pfx float64 // P value to be passed to fx function
}

// Init initialises this structure
func (o *PressCylin) Init(prms dbf.Params) {

	// default values
	o.a = 100    // [mm]
	o.b = 200    // [mm]
	o.E = 210000 // [MPa] Young modulus
	o.ν = 0.3    // [-] Poisson's ratio
	o.σy = 240   // [MPa] yield stress

	// parameters
	for _, p := range prms {
		switch p.N {
		case "a":
			o.a = p.V
		case "b":
			o.b = p.V
		case "E":
			o.E = p.V
		case "ν":
			o.ν = p.V
		case "σy":
			o.σy = p.V
		}
	}

	// derived
	o.coef = o.a * o.a / (o.b * o.b)
	o.Y = 2.0 * o.σy / math.Sqrt(3.0)
	o.P0 = o.Y * (1 - o.coef) / 2.0
	o.Plim = o.Y * math.Log(o.b/o.a)
}

// Plastic computes the pressure corresponding to c and the radial displacment
// at the outer surface
func (o PressCylin) Plastic(c float64) (P, ub float64) {
	P = o.Y * (math.Log(c/o.a) + (1.0-c*c/(o.b*o.b))/2.0)
	ub = o.Y * c * c * (1.0 - o.ν*o.ν) / (o.E * o.b)
	return
}

// ElastOuterU computes the elastic solution for the radial displacement
// at the outer surface
func (o PressCylin) ElastOuterU(P float64) (ub float64) {
	ub = 2.0 * P * o.b * (1.0 - o.ν*o.ν) / (o.E/o.coef - o.E)
	return
}

// CalcC computes the elastic/plastic transition zone
// TODO: check what's 'c' exactly
func (o *PressCylin) CalcC(P float64) float64 {
	var nls num.NlSolver
	defer nls.Free()
	o.Pfx = P
	Res := []float64{(o.a + o.b) / 2.0} // initial values
	nls.Init(1, o.fxFun, nil, o.dfdxFun, true, false, nil)
	nls.Solve(Res, true)
	return Res[0]
}

// Stresses compute the radial and tangential stresses
func (o PressCylin) Stresses(c, r float64) (sr, st float64) {
	b, Y := o.b, o.Y
	if r > c { // elastic
		sr = -Y * c * c * (b*b/(r*r) - 1.0) / (2.0 * b * b)
		st = Y * c * c * (b*b/(r*r) + 1.0) / (2.0 * b * b)
	} else {
		sr = Y * (-0.5 - math.Log(c/r) + c*c/(2.0*b*b))
		st = Y * (0.5 - math.Log(c/r) + c*c/(2.0*b*b))
	}
	return sr, st
}

// plot //////////////////////////////////////////////////////////////////////////

// CalcPressDisp returns the internal pressure and outer displacements for
// plotting the load-displacement graph
func (o PressCylin) CalcPressDisp(np int) (P, Ub []float64) {

	// elastic
	ne := 3
	dP0 := o.P0 / float64(ne-1)
	P = make([]float64, ne+np)
	Ub = make([]float64, ne+np)
	for i := 0; i < ne; i++ {
		P[i] = float64(i) * dP0
		Ub[i] = o.ElastOuterU(P[i])
	}

	// plastic
	C := utl.LinSpace(o.a, o.b, np)
	for i := 0; i < np; i++ {
		P[ne+i], Ub[ne+i] = o.Plastic(C[i])
	}
	return
}

// CalcStresses computes stresses
func (o PressCylin) CalcStresses(Pvals []float64, nr int) (R []float64, Sr, St [][]float64) {
	R = utl.LinSpace(o.a, o.b, nr)
	np := len(Pvals)
	Sr = utl.Alloc(np, nr)
	St = utl.Alloc(np, nr)
	for i, P := range Pvals {
		c := o.CalcC(P)
		for j := 0; j < nr; j++ {
			Sr[i][j], St[i][j] = o.Stresses(c, R[j])
		}
	}
	return
}

// auxiliary /////////////////////////////////////////////////////////////////////

// fxFun implements the nonlinear problem to be solved whe finding c
func (o PressCylin) fxFun(fx, X la.Vector) (err error) {
	x := X[0]
	fx[0] = o.Pfx/o.Y - (math.Log(x/o.a) + (1.0-x*x/(o.b*o.b))/2.0)
	return
}

// dfdxFun is the derivative of fxFun
func (o PressCylin) dfdxFun(dfdx *la.Matrix, X la.Vector) (err error) {
	x := X[0]
	dfdx.Set(0, 0, -1.0/x+x/(o.b*o.b))
	return
}
