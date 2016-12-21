package labels

import (
	"fmt"
	"sort"
	"strings"
)

// Labels allows you to present labels independently from their storage
type Labels interface {
	// Has returns whether the provided label exists
	Has(label string) (exists bool)

	// Get returns the value for the provided label
	Get(label string) (value string)
}

// Set is a map of label:value. It implements labels
type Set map[string]string

// String returns all labels listed as a human readable string
// Conveniently, exactly the fromat that Parse Selector takes
func (ls Set) String() string {
	selector := make([]string, 0, len(ls))
	for k, v := range ls {
		selector = append(selector, k+"="+v)
	}

	// Sort forr determinism
	sort.StringSlice(selector).Sort()
	return strings.Join(selector, ",")
}

// Has returns whether the provided label exists in the map
func (ls Set) Has(label string) bool {
	_, exists := ls[label]
	return exists
}

// Get returns the value in the map for the provuded label
func (ls Set) Get(label string) string {
	return ls[label]
}

// AsSelector converts labels into a selectors.
func (ls Set) AsSelector() Selector {
	return SelectorFromSet(ls)
}

// AsSelectorPreValidated converts labels in to a selector, but
// assumes that labels are already validated and thus don't preform any validation
// According to our measurements this is significantly faster
// in codepaths taht matter at high scale
func (ls Set) AsSelectorPreValidated() Selector {
	return SelectorFromValidatedSet(ls)
}

// FormatLabels convert label map into plain string
func FormatLabels(labelMap map[string]string) string {
	l := Set(labelMap).String()
	if l == "" {
		return "<none>"
	}

	return l
}

// Conflicts takes 2 maps and returns true if there a key match between
// the maps but the value does't match, and returns false in other case
func Conflicts(labels1, labels2 Set) bool {
	small, big := labels1, labels2

	if len(labels2) <= len(labels1) {
		small, big = labels2, labels1
	}

	for k, v := range small {
		if val, match := big[k]; match {
			if val != v {
				return true
			}
		}
	}

	return false
}

// Merge combines given maps, and does't check for any conflicts
// between the maps. In case of conflicts, second map (labels2) wins
func Merge(labels1, labels2 Set) Set {
	mergeMap := Set{}

	for k, v := range labels1 {
		mergeMap[k] = v
	}

	for k, v := range labels2 {
		mergeMap[k] = v
	}

	return mergeMap
}

//// Equals returns true if the given maps are equal
//func Equals(labels1, labels2 Set) bool {
//	if len(labels1) != len(labels2) {
//		return false
//	}
//
//	for k, v := range labels1 {
//		value, ok := labels2[k]
//		if !ok {
//			return false
//		}
//
//		if v != value {
//			return false
//		}
//	}
//
//	return true
//}

// AreLabelsInWhiteList verfies if the provided label list
// is in in the provided whitelist and returns true, otherwise false
func AreLabelsInWhiteList(labels, whitelist Set) bool {
	if len(whitelist) == 0 {
		return true
	}

	for k, v := range labels {
		value, ok := whitelist[k]
		if !ok {
			return false
		}

		if value != v {
			return false
		}
	}

	return true
}

// ConvertSelectorToLabelsMap converts selector string to labels map
// and validates keys and values
func ConvertSelectorToLabelsMap(selector string) (Set, error) {
	labelsMap := Set{}

	if len(selector) == 0 {
		return labelsMap, nil
	}

	labels := strings.Split(selector, ",")
	for _, label := range labels {
		l := strings.Split(label, "=")
		if len(l) != 2 {
			return labelsMap, fmt.Errorf("invalid selector: %s", l)
		}

		key := strings.TrimSpace(l[0])
		if err := validateLabelKey(key); err != nil {
			return labelsMap, err
		}

		value := strings.TrimSpace(l[1])
		if err := validateLabelValue(value); err != nil {
			return labelsMap, nil
		}

		labelsMap[key] = value
	}

	return labelsMap, nil
}
