//Package securitydefinition msg type = d.
package securitydefinition

import (
	"github.com/quickfixgo/quickfix"
	"github.com/quickfixgo/quickfix/enum"
	"github.com/quickfixgo/quickfix/fix44"
	"github.com/quickfixgo/quickfix/fix44/instrument"
	"github.com/quickfixgo/quickfix/fix44/instrumentextension"
	"github.com/quickfixgo/quickfix/fix44/instrumentleg"
	"github.com/quickfixgo/quickfix/fix44/underlyinginstrument"
)

//NoUnderlyings is a repeating group in SecurityDefinition
type NoUnderlyings struct {
	//UnderlyingInstrument Component
	underlyinginstrument.UnderlyingInstrument
}

//NoLegs is a repeating group in SecurityDefinition
type NoLegs struct {
	//InstrumentLeg Component
	instrumentleg.InstrumentLeg
}

//Message is a SecurityDefinition FIX Message
type Message struct {
	FIXMsgType string `fix:"d"`
	fix44.Header
	//SecurityReqID is a required field for SecurityDefinition.
	SecurityReqID string `fix:"320"`
	//SecurityResponseID is a required field for SecurityDefinition.
	SecurityResponseID string `fix:"322"`
	//SecurityResponseType is a required field for SecurityDefinition.
	SecurityResponseType int `fix:"323"`
	//Instrument Component
	instrument.Instrument
	//InstrumentExtension Component
	instrumentextension.InstrumentExtension
	//NoUnderlyings is a non-required field for SecurityDefinition.
	NoUnderlyings []NoUnderlyings `fix:"711,omitempty"`
	//Currency is a non-required field for SecurityDefinition.
	Currency *string `fix:"15"`
	//TradingSessionID is a non-required field for SecurityDefinition.
	TradingSessionID *string `fix:"336"`
	//TradingSessionSubID is a non-required field for SecurityDefinition.
	TradingSessionSubID *string `fix:"625"`
	//Text is a non-required field for SecurityDefinition.
	Text *string `fix:"58"`
	//EncodedTextLen is a non-required field for SecurityDefinition.
	EncodedTextLen *int `fix:"354"`
	//EncodedText is a non-required field for SecurityDefinition.
	EncodedText *string `fix:"355"`
	//NoLegs is a non-required field for SecurityDefinition.
	NoLegs []NoLegs `fix:"555,omitempty"`
	//ExpirationCycle is a non-required field for SecurityDefinition.
	ExpirationCycle *int `fix:"827"`
	//RoundLot is a non-required field for SecurityDefinition.
	RoundLot *float64 `fix:"561"`
	//MinTradeVol is a non-required field for SecurityDefinition.
	MinTradeVol *float64 `fix:"562"`
	fix44.Trailer
}

//Marshal converts Message to a quickfix.Message instance
func (m Message) Marshal() quickfix.Message { return quickfix.Marshal(m) }

func (m *Message) SetSecurityReqID(v string)          { m.SecurityReqID = v }
func (m *Message) SetSecurityResponseID(v string)     { m.SecurityResponseID = v }
func (m *Message) SetSecurityResponseType(v int)      { m.SecurityResponseType = v }
func (m *Message) SetNoUnderlyings(v []NoUnderlyings) { m.NoUnderlyings = v }
func (m *Message) SetCurrency(v string)               { m.Currency = &v }
func (m *Message) SetTradingSessionID(v string)       { m.TradingSessionID = &v }
func (m *Message) SetTradingSessionSubID(v string)    { m.TradingSessionSubID = &v }
func (m *Message) SetText(v string)                   { m.Text = &v }
func (m *Message) SetEncodedTextLen(v int)            { m.EncodedTextLen = &v }
func (m *Message) SetEncodedText(v string)            { m.EncodedText = &v }
func (m *Message) SetNoLegs(v []NoLegs)               { m.NoLegs = v }
func (m *Message) SetExpirationCycle(v int)           { m.ExpirationCycle = &v }
func (m *Message) SetRoundLot(v float64)              { m.RoundLot = &v }
func (m *Message) SetMinTradeVol(v float64)           { m.MinTradeVol = &v }

//A RouteOut is the callback type that should be implemented for routing Message
type RouteOut func(msg Message, sessionID quickfix.SessionID) quickfix.MessageRejectError

//Route returns the beginstring, message type, and MessageRoute for this Message type
func Route(router RouteOut) (string, string, quickfix.MessageRoute) {
	r := func(msg quickfix.Message, sessionID quickfix.SessionID) quickfix.MessageRejectError {
		m := new(Message)
		if err := quickfix.Unmarshal(msg, m); err != nil {
			return err
		}
		return router(*m, sessionID)
	}
	return enum.BeginStringFIX44, "d", r
}
