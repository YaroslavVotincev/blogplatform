package blogs

func unique(s []string) []string {
	inResult := make(map[string]bool)
	var result []string
	for _, str := range s {
		if _, ok := inResult[str]; !ok {
			inResult[str] = true
			if str == "" {
				continue
			}
			result = append(result, str)
		}
	}
	return result
}
