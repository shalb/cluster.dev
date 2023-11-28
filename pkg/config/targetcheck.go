package config

import "strings"

type TargetUnits []string

func NewTargetsChecker(tg []string) *TargetUnits {
	res := TargetUnits{}
	res = append(res, tg...)
	return &res
}

func (t *TargetUnits) Check(unitKey string) bool {
	for _, target := range *t {
		tgSplitted := strings.Split(target, ".")
		uKeySplitted := strings.Split(unitKey, ".")
		if len(tgSplitted) == 0 || len(tgSplitted) > 2 || len(uKeySplitted) != 2 {
			return false
		}
		if len(tgSplitted) == 1 {
			// Target is whole stack, check unit stack name only.
			if uKeySplitted[0] == tgSplitted[0] {
				return true
			}
			continue
		}
		// The target is unit, compare unit name and stack name.
		if uKeySplitted[0] == tgSplitted[0] && uKeySplitted[1] == tgSplitted[1] {
			return false
		}
	}
	return false
}
