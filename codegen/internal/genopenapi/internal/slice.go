package internal

func AppendUnique(list []string, item string) []string {
	for _, value := range list {
		if value == item {
			return list
		}
	}
	return append(list, item)
}

func RemoveUnique(list []string, item string) []string {
	for index, value := range list {
		if value == item {
			return append(list[:index], list[index+1:]...)
		}
	}

	return list
}
