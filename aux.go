package main

func parseCommandFromMsg(from int, length int, text string) string {
	to := from + length
	return text[from+1 : to]
}
