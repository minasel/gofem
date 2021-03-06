// Copyright 2016 The Gofem Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package solid

import (
	"testing"

	"github.com/cpmech/gosl/chk"
	"github.com/cpmech/gosl/fun/dbf"
	"github.com/cpmech/gosl/io"
	"github.com/cpmech/gosl/la"
)

// Driver run simulations with constitutive models for solids
type Driver struct {

	// input
	nsig  int   // number of stress components
	model Model // solid model

	// settings
	Silent bool    // do not show error messages
	TolD   float64 // tolerance to check consistent matrix
	VerD   bool    // verbose check of D
	WithPC bool    // with predictor-corrector data

	// check D matrix
	TstD *testing.T // if != nil, do check consistent matrix

	// results
	Res []*State    // stress/ivs results
	Eps [][]float64 // strains

	// for checking consistent matrix
	D [][]float64 // consistent matrix

	// for predictor-corrector plots
	PreCor [][]float64 // predictor-corrector stresses
}

// Init initialises driver
func (o *Driver) Init(simfnk, modelname string, ndim int, pstress bool, prms dbf.Params) (err error) {
	o.nsig = 2 * ndim
	o.model, err = New(modelname)
	if err != nil {
		return
	}
	err = o.model.Init(ndim, pstress, prms)
	if err != nil {
		return
	}
	o.D = la.MatAlloc(o.nsig, o.nsig)
	o.TolD = 1e-8
	o.VerD = chk.Verbose
	return
}

// InitWithModel initialises driver with existent model
func (o *Driver) InitWithModel(ndim int, model Model) (err error) {
	o.nsig = 2 * ndim
	o.model = model
	o.D = la.MatAlloc(o.nsig, o.nsig)
	o.TolD = 1e-8
	o.VerD = chk.Verbose
	return
}

// Run runs simulation
func (o *Driver) Run(pth *Path) (err error) {

	// specialised models
	var sml Small
	var eup SmallStrainUpdater
	switch m := o.model.(type) {
	case Small:
		sml = m
	case SmallStrainUpdater:
		eup = m
	default:
		return chk.Err("cannot handle large-deformation models yet\n")
	}

	// elastoplastic model
	epm := o.model.(EPmodel)

	// initial stresses
	σ0 := make([]float64, o.nsig)
	σ0[0] = pth.MultS * pth.Sx[0]
	σ0[1] = pth.MultS * pth.Sy[0]
	σ0[2] = pth.MultS * pth.Sz[0]

	// allocate results arrays
	nr := 1 + (pth.Size()-1)*pth.Nincs
	if nr < 2 {
		return chk.Err("size of path is incorrect. Size=%d, Nincs=%d\n", pth.Size(), pth.Nincs)
	}
	o.Res = make([]*State, nr)
	o.Eps = la.MatAlloc(nr, o.nsig)
	for i := 0; i < nr; i++ {
		o.Res[i], err = o.model.InitIntVars(σ0)
		if err != nil {
			return
		}
	}

	// put initial stress in predictor-corrector array
	o.PreCor = [][]float64{o.Res[0].Sig}

	// auxiliary variables
	Δσ := make([]float64, o.nsig)
	Δε := make([]float64, o.nsig)

	// variables for checking D
	var εold, εnew, Δεtmp []float64
	var stmp *State
	if o.TstD != nil {
		εold = make([]float64, o.nsig)
		εnew = make([]float64, o.nsig)
		Δεtmp = make([]float64, o.nsig)
		stmp, err = o.model.InitIntVars(σ0)
		if err != nil {
			return
		}
	}

	// update states
	k := 1
	for i := 1; i < pth.Size(); i++ {

		// stress path
		if pth.UseS[i] > 0 {

			return chk.Err("cannot run StrainUpdate for stress paths at the moment")

			Δσ[0] = pth.MultS * (pth.Sx[i] - pth.Sx[i-1]) / float64(pth.Nincs)
			Δσ[1] = pth.MultS * (pth.Sy[i] - pth.Sy[i-1]) / float64(pth.Nincs)
			Δσ[2] = pth.MultS * (pth.Sz[i] - pth.Sz[i-1]) / float64(pth.Nincs)
			for inc := 0; inc < pth.Nincs; inc++ {

				// update
				o.Res[k].Set(o.Res[k-1])
				copy(o.Eps[k], o.Eps[k-1])
				if eup != nil {
					err = eup.StrainUpdate(o.Res[k], Δσ)
				}
				if err != nil {
					if !o.Silent {
						io.Pfred("strain update failed\n%v\n", err)
					}
					return
				}
				k += 1
			}
		}

		// strain path
		if pth.UseE[i] > 0 {
			Δε[0] = pth.MultE * (pth.Ex[i] - pth.Ex[i-1]) / float64(pth.Nincs)
			Δε[1] = pth.MultE * (pth.Ey[i] - pth.Ey[i-1]) / float64(pth.Nincs)
			Δε[2] = pth.MultE * (pth.Ez[i] - pth.Ez[i-1]) / float64(pth.Nincs)
			for inc := 0; inc < pth.Nincs; inc++ {

				// update strains
				la.VecAdd2(o.Eps[k], 1, o.Eps[k-1], 1, Δε) // εnew = εold + Δε

				// update stresses
				o.Res[k].Set(o.Res[k-1])
				err = sml.Update(o.Res[k], o.Eps[k], Δε, 0, 0, 0)
				if err != nil {
					if !o.Silent {
						io.Pfred("stress update failed\n%v\n", err)
					}
					return
				}
				if epm != nil {
					tmp := o.Res[k-1].GetCopy()
					epm.ElastUpdate(tmp, o.Eps[k])
					o.PreCor = append(o.PreCor, tmp.Sig, o.Res[k].Sig)
				}

				// check consistent matrix
				if o.TstD != nil {
					firstIt := false
					err = sml.CalcD(o.D, o.Res[k], firstIt)
					if err != nil {
						return chk.Err("check of consistent matrix failed:\n %v\n\n", err)
					}
					copy(εold, o.Eps[k-1])
					copy(εnew, o.Eps[k])
					if o.VerD {
						io.Pf("\n")
					}
					chk.DerivVecVec(o.TstD, "D", o.TolD, o.D, εnew, 1e-3, o.VerD, func(f, x []float64) error {
						for l := 0; l < o.nsig; l++ {
							Δεtmp[l] = x[l] - εold[l]
						}
						stmp.Set(o.Res[k-1])
						err = sml.Update(stmp, x, Δεtmp, 0, 0, 0)
						if err != nil {
							return err
						}
						copy(f, stmp.Sig)
						return nil
					})
				}
				k += 1
			}
		}
	}
	return
}
