package unit

func shave(s string, n int) string {
	l := len(s)
	if l <= n {
		return ""
	}
	return s[:l-n]
}
