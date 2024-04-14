package transform

func MappingPrefixes(mapping map[string]string) []string {
	prefixes := make([]string, 0)
	for prefix := range mapping {
		prefixes = append(prefixes, prefix)
	}

	return prefixes
}
