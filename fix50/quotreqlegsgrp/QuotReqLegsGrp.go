package quotreqlegsgrp

import (
	"github.com/quickfixgo/quickfix/fix50/instrumentleg"
	"github.com/quickfixgo/quickfix/fix50/legbenchmarkcurvedata"
	"github.com/quickfixgo/quickfix/fix50/legstipulations"
	"github.com/quickfixgo/quickfix/fix50/nestedparties"
)

//NoLegs is a repeating group in QuotReqLegsGrp
type NoLegs struct {
	//InstrumentLeg Component
	instrumentleg.InstrumentLeg
	//LegQty is a non-required field for NoLegs.
	LegQty *float64 `fix:"687"`
	//LegSwapType is a non-required field for NoLegs.
	LegSwapType *int `fix:"690"`
	//LegSettlType is a non-required field for NoLegs.
	LegSettlType *string `fix:"587"`
	//LegSettlDate is a non-required field for NoLegs.
	LegSettlDate *string `fix:"588"`
	//LegStipulations Component
	legstipulations.LegStipulations
	//NestedParties Component
	nestedparties.NestedParties
	//LegBenchmarkCurveData Component
	legbenchmarkcurvedata.LegBenchmarkCurveData
	//LegOrderQty is a non-required field for NoLegs.
	LegOrderQty *float64 `fix:"685"`
	//LegOptionRatio is a non-required field for NoLegs.
	LegOptionRatio *float64 `fix:"1017"`
	//LegPrice is a non-required field for NoLegs.
	LegPrice *float64 `fix:"566"`
	//LegRefID is a non-required field for NoLegs.
	LegRefID *string `fix:"654"`
}

//QuotReqLegsGrp is a fix50 Component
type QuotReqLegsGrp struct {
	//NoLegs is a non-required field for QuotReqLegsGrp.
	NoLegs []NoLegs `fix:"555,omitempty"`
}

func (m *QuotReqLegsGrp) SetNoLegs(v []NoLegs) { m.NoLegs = v }
