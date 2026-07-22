package daytime

import "time"

type Window struct {
	Name  string
	Start int
	End   int
}

var (
	Morning   = Window{Name: "morning", Start: 8, End: 12}
	Afternoon = Window{Name: "afternoon", Start: 12, End: 16}
	Evening   = Window{Name: "evening", Start: 16, End: 21}
	Windows   = []Window{Morning, Afternoon, Evening}
)

func StartTime(date time.Time, windowName string) time.Time {
	for _, window := range Windows {
		if window.Name == windowName {
			return time.Date(date.Year(), date.Month(), date.Day(), window.Start, 0, 0, 0, time.UTC)
		}
	}
	return time.Date(date.Year(), date.Month(), date.Day(), Morning.Start, 0, 0, 0, time.UTC)
}

func EndTime(date time.Time, windowName string) time.Time {
	for _, window := range Windows {
		if window.Name == windowName {
			return time.Date(date.Year(), date.Month(), date.Day(), window.End, 0, 0, 0, time.UTC)
		}
	}
	return time.Date(date.Year(), date.Month(), date.Day(), Evening.End, 0, 0, 0, time.UTC)
}

func WindowForTime(date, value time.Time) (string, bool) {
	for _, window := range Windows {
		start := StartTime(date, window.Name)
		end := EndTime(date, window.Name)
		if (value.Equal(start) || value.After(start)) && value.Before(end) {
			return window.Name, true
		}
	}
	return "", false
}

func Order(windowName string) int {
	for idx, window := range Windows {
		if window.Name == windowName {
			return idx
		}
	}
	return len(Windows)
}
