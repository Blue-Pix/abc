package util

import "sort"

func SortStringMap(m map[string]string) []string {
	keys := make([]string, len(m))
	index := 0
	for key := range m {
		keys[index] = key
		index++
	}
	sort.Strings(keys)
	return keys
}

func SortGeneralMap(m map[string]interface{}) []string {
	keys := make([]string, len(m))
	index := 0
	for key := range m {
		keys[index] = key
		index++
	}
	sort.Strings(keys)
	return keys
}
