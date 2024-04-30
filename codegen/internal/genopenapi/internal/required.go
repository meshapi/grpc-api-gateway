package internal

type RequiredSet map[string]struct{}

func RequiredSetFromSlice(values []string) RequiredSet {
	if len(values) == 0 {
		return nil
	}

	result := make(RequiredSet, len(values))
	for _, v := range values {
		result[v] = struct{}{}
	}

	return result
}

func AppendToRequiredSet(requiredSet RequiredSet, value string) RequiredSet {
	if requiredSet == nil {
		return RequiredSet{value: struct{}{}}
	}

	requiredSet[value] = struct{}{}
	return requiredSet
}

func RequiredSliceFromRequiredSet(requiredSet RequiredSet) []string {
	if len(requiredSet) == 0 {
		return nil
	}

	result := make([]string, 0, len(requiredSet))
	for key := range requiredSet {
		result = append(result, key)
	}

	return result
}
