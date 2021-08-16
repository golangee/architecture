package model

// Story is a user story in the form "as a ... i want to ... so that ...".
// It contains a general title, an ID to reference it and several criteria
// to describe when it can be accepted.
type Story struct {
	Id      int               `tadl:"id"`
	Title   string            `tadl:"title"`
	AsA     string            `tadl:"as_a"`
	IWantTo string            `tadl:"i_want_to"`
	SoThat  string            `tadl:"so_that"`
	Accept  []AcceptCriterion `tadl:"accept"`
}

// AcceptCriterion describes a single testable criteria. It has some
// requirements, triggers and results.
type AcceptCriterion struct {
	Require []string `tadl:"require"`
	When    []string `tadl:"when"`
	Then    []string `tadl:"then"`
}
