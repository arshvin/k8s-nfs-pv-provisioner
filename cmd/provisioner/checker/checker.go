package checker

//Checker is the gatekeeper interface for PVC and PV to be processed by the provisioner
type Checker interface {
	PerformChecks()
	IsAllOK() bool
	checkList() map[int]func() bool
}

//AbstractChecker is absctract checker which is implemented a couple methods
type AbstractChecker struct {
	Checker
	Results map[int]bool
}

//PerformChecks is the entry point to checking porcess for client code
func (ch *AbstractChecker) PerformChecks() {

	checks := ch.checkList()

	ch.Results = make(map[int]bool)
	for chName := range checks {
		ch.Results[chName] = false
	}

	for chName, chFn := range checks {
		if result := chFn(); result {
			ch.Results[chName] = result
		} else {
			return
		}
	}
}

//IsAllOK is method that return true is all checks were passed successfully otherwise false
func (ch AbstractChecker) IsAllOK() bool {
	for _, ok := range ch.Results {
		if !ok {
			return false
		}
	}
	return true
}
