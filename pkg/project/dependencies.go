package project

import (
	"fmt"
	"strings"
)

func checkUnitDependencies(p *Project) error {
	for _, uniit := range p.Units {
		if err := checkDependenciesRecursive(uniit); err != nil {
			return fmt.Errorf("unresolved dependency in unit %v.%v: %w", uniit.Stack().Name, uniit.Name(), err)
		}
	}
	return nil
}

func checkDependenciesRecursive(unit Unit, chain ...string) error {
	if err := checkUnitDependenciesCircle(chain); err != nil {
		return err
	}
	chain = append(chain, unit.Key())
	for _, dep := range unit.Dependencies().Slice() {
		if err := checkDependenciesRecursive(dep.Unit, chain...); err != nil {
			return err
		}
	}
	return nil
}

func checkUnitDependenciesCircle(chain []string) error {
	if len(chain) < 2 {
		return nil
	}
	circleCheck := []string{}
	for _, str := range chain {
		for _, comareStr := range circleCheck {
			// log.Warnf("Compare: %v == %v", str, )
			if str == comareStr {
				circleCheck = append(circleCheck, str)
				return fmt.Errorf("loop: %s", strings.Join(circleCheck, " -> "))
			}
		}
		circleCheck = append(circleCheck, str)
	}
	return nil
}
