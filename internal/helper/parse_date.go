package helper

func parseExcelDate(val string) (time.Time, error) {
	return time.Parse("2006-01-02", strings.TrimSpace(val))
}