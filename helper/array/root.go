package array

func IncludesString(s []string, v string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] == v {
			return true
		}
	}

	return false
}
