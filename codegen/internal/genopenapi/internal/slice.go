package internal

func FilteredStringSlice(list []string, itemsToExclude map[string]struct{}) []string {
	if len(list) <= len(itemsToExclude) {
		return nil
	}
	// NOTE: Sacrifice memory for time here. Overallocation is a possibility here.
	newList := make([]string, 0, len(list))

	for _, value := range list {
		if _, exclude := itemsToExclude[value]; exclude {
			continue
		}
		newList = append(newList, value)
	}

	return newList
}
