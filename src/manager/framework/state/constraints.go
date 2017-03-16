package state

import (
	"errors"
	"regexp"
	"strings"

	"github.com/Dataman-Cloud/swan/src/mesosproto/mesos"
	"github.com/Dataman-Cloud/swan/src/utils"
)

//go:generate goyacc -o ./constraints_gen.go  ./constraints_parser.y

var uniqWhat = []string{
	"hostname",
}

type ConstraintParamHolder struct {
	Slot  *Slot
	Offer *mesos.Offer
}

type Statement interface {
	Eval() bool
	Valid() error
	SetContext(ctx *ConstraintParamHolder)
}

// not (unique hostname)
type NotStatement struct {
	Op1 Statement
}

func (ns *NotStatement) Eval() bool {
	return !ns.Op1.Eval()
}

func (ns *NotStatement) Valid() error {
	err := ns.Op1.Valid()
	if err != nil {
		return err
	}

	return nil
}

func (ns *NotStatement) SetContext(ctx *ConstraintParamHolder) {
	ns.Op1.SetContext(ctx)
}

// and (unique hostname) (unique ip)
// and (not (unique hostname)) (unique ip)
type AndStatement struct {
	Op1 Statement
	Op2 Statement
}

func (as *AndStatement) Eval() bool {
	return as.Op2.Eval() && as.Op1.Eval()
}

func (as *AndStatement) Valid() error {
	err := as.Op1.Valid()
	if err != nil {
		return err
	}

	err1 := as.Op2.Valid()
	if err1 != nil {
		return err1
	}

	return nil
}

func (as *AndStatement) SetContext(ctx *ConstraintParamHolder) {
	as.Op1.SetContext(ctx)
	as.Op2.SetContext(ctx)
}

// or (like ip foobar) (unique hostname)
type OrStatement struct {
	ConstraintParamHolder
	Op1 Statement
	Op2 Statement
}

func (os *OrStatement) Eval() bool {
	return os.Op2.Eval() || os.Op1.Eval()
}

func (os *OrStatement) Valid() error {
	err := os.Op1.Valid()
	if err != nil {
		return err
	}

	err1 := os.Op2.Valid()
	if err1 != nil {
		return err1
	}

	return nil
}

func (os *OrStatement) SetContext(ctx *ConstraintParamHolder) {
	os.Op1.SetContext(ctx)
	os.Op2.SetContext(ctx)
}

// (unique hostname)
type UniqueStatment struct {
	ConstraintParamHolder
	What string
}

func (us *UniqueStatment) Eval() bool {
	if us.What == "hostname" {
		slotsOnHost := OfferAllocatorInstance().SlotsByHostname(us.Offer.GetHostname())
		for _, slotOnHost := range slotsOnHost { // slots belongs to same app on same host
			if strings.SplitN(slotOnHost, "-", 2)[1] == strings.SplitN(us.Slot.ID, "-", 2)[1] {
				return false
			}
		}
	}

	if us.What == "agentid" {
		slotsOnAgent := OfferAllocatorInstance().SlotsByAgentID(*us.Offer.GetAgentId().Value)
		for _, slotOnAgent := range slotsOnAgent { // slots belongs to same app on same agentID
			if strings.SplitN(slotOnAgent, "-", 2)[1] == strings.SplitN(us.Slot.ID, "-", 2)[1] {
				return false
			}
		}
	}

	return true
}

func (us *UniqueStatment) Valid() error {
	if utils.SliceContains(uniqWhat, us.What) {
		return nil
	} else {
		return errors.New("only hostname is supported for the time being")
	}
}

func (us *UniqueStatment) SetContext(ctx *ConstraintParamHolder) {
	us.Offer = ctx.Offer
	us.Slot = ctx.Slot
}

// like hostname foobar*
type LikeStatement struct {
	ConstraintParamHolder
	What  string
	Regex string
}

func (ls *LikeStatement) Eval() bool {
	r := regexp.MustCompile(ls.Regex)
	if ls.What == "hostname" {
		return r.MatchString(ls.Offer.GetHostname())
	}

	if ls.What == "agentid" {
		return r.MatchString(*ls.Offer.GetAgentId().Value)
	}

	// user defined attributes match
	for _, attr := range ls.Offer.Attributes {
		if attr.GetName() == ls.What && attr.GetType() == mesos.Value_TEXT {
			return r.MatchString(*attr.GetText().Value)
		}
	}

	return false
}

func (ls *LikeStatement) Valid() error {
	return nil
}

func (ls *LikeStatement) SetContext(ctx *ConstraintParamHolder) {
	ls.Offer = ctx.Offer
	ls.Slot = ctx.Slot
}

// equal hostname xxxx
type EqualStatement struct {
	ConstraintParamHolder
	What  string
	Regex string
}

func (ls *EqualStatement) Eval() bool {
	if ls.What == "hostname" {
		return ls.Offer.GetHostname() == ls.Regex
	}

	if ls.What == "agentid" {
		return *ls.Offer.GetAgentId().Value == ls.Regex
	}

	// user defined attributes match
	for _, attr := range ls.Offer.Attributes {
		if attr.GetName() == ls.What && attr.GetType() == mesos.Value_TEXT {
			return *attr.GetText().Value == ls.Regex
		}
	}

	return false
}

func (ls *EqualStatement) Valid() error {
	if ls.What != "" && ls.Regex != "" {
		return nil
	} else {
		return errors.New("Equal statement must contain two operand")
	}
}

func (ls *EqualStatement) SetContext(ctx *ConstraintParamHolder) {
	ls.Offer = ctx.Offer
	ls.Slot = ctx.Slot
}

// contains hostname barfoo
type ContainsStatement struct {
	ConstraintParamHolder
	What  string
	Regex string
}

func (cs *ContainsStatement) Eval() bool {
	if cs.What == "hostname" {
		return strings.Contains(cs.Offer.GetHostname(), cs.Regex)
	}

	if cs.What == "agentid" {
		return strings.Contains(*cs.Offer.GetAgentId().Value, cs.Regex)
	}

	// user defined attributes match
	for _, attr := range cs.Offer.Attributes {
		if attr.GetName() == cs.What && attr.GetType() == mesos.Value_TEXT {
			return strings.Contains(*attr.GetText().Value, cs.Regex)
		}
	}

	return false
}

func (cs *ContainsStatement) Valid() error {
	return nil
}

func (cs *ContainsStatement) SetContext(ctx *ConstraintParamHolder) {
	cs.Offer = ctx.Offer
	cs.Slot = ctx.Slot
}
