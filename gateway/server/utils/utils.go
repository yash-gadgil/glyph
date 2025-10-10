package utils

func Contains(list []string, target string) bool {
	for _, v := range list {
		if v == target {
			return true
		}
	}
	return false
}

var ValidOrderStatuses = []string{
	"open",
	"executed",
}
