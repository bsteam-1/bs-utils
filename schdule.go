package utils

import (
	"encoding/json"
	"fmt"
	"time"
)

// TimePoint represents a specific time with hour and minute
type TimePoint struct {
	Hour   int `json:"hour"`
	Minute int `json:"minute"`
}

// DaySchedule represents the start and end times for a day
type DaySchedule struct {
	Start TimePoint `json:"start"`
	End   TimePoint `json:"end"`
}

// DeviceSchedule represents the complete schedule for a device
type DeviceSchedule struct {
	Schedules map[string]DaySchedule `json:"schedules"`
}

// Schedule manages the device operation schedule
type Schedule struct {
	deviceID string
	schedule DeviceSchedule
}

// NewSchedule creates a new Schedule instance
func NewSchedule(deviceID string) *Schedule {
	// Initialize with default schedules
	defaultSchedule := DeviceSchedule{
		Schedules: map[string]DaySchedule{
			"mon": {Start: TimePoint{23, 0}, End: TimePoint{9, 0}},
			"tue": {Start: TimePoint{23, 10}, End: TimePoint{9, 0}},
			"wed": {Start: TimePoint{23, 30}, End: TimePoint{9, 30}},
			"thu": {Start: TimePoint{23, 0}, End: TimePoint{9, 0}},
			"fri": {Start: TimePoint{23, 10}, End: TimePoint{9, 0}},
			"sat": {Start: TimePoint{23, 10}, End: TimePoint{12, 0}},
			"sun": {Start: TimePoint{22, 30}, End: TimePoint{12, 0}},
		},
	}

	return &Schedule{
		deviceID: deviceID,
		schedule: defaultSchedule,
	}
}

// LoadSchedule loads schedule from JSON string
func (s *Schedule) LoadSchedule(jsonStr string) error {
	return json.Unmarshal([]byte(jsonStr), &s.schedule)
}

// SaveSchedule saves current schedule to JSON string
func (s *Schedule) SaveSchedule() (string, error) {
	data, err := json.Marshal(s.schedule)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// getDayKey converts time.Weekday to our schedule key
func getDayKey(day time.Weekday) string {
	days := map[time.Weekday]string{
		time.Monday:    "mon",
		time.Tuesday:   "tue",
		time.Wednesday: "wed",
		time.Thursday:  "thu",
		time.Friday:    "fri",
		time.Saturday:  "sat",
		time.Sunday:    "sun",
	}
	return days[day]
}

// IsOperating checks if the device should be operating at the given time
func (s *Schedule) IsOperating(t time.Time) bool {
	currentMinutes := t.Hour()*60 + t.Minute()

	// 현재 요일의 스케줄 확인
	currentDayKey := getDayKey(t.Weekday())
	currentSchedule, exists := s.schedule.Schedules[currentDayKey]
	if !exists {
		return false
	}

	// 이전 요일의 스케줄 확인
	previousDay := t.Add(-24 * time.Hour).Weekday()
	previousDayKey := getDayKey(previousDay)
	previousSchedule, exists := s.schedule.Schedules[previousDayKey]
	if !exists {
		return false
	}

	// 현재 시간이 자정 이전인지 확인 (예: 23:30)
	currentDayStartMinutes := currentSchedule.Start.Hour*60 + currentSchedule.Start.Minute

	// 현재 시간이 자정 이후인지 확인 (예: 01:30)
	previousDayEndMinutes := previousSchedule.End.Hour*60 + previousSchedule.End.Minute

	// 자정 이후 시간대(0시~12시)인 경우
	if currentMinutes <= 14*60 {
		// 이전 날의 종료 시간이 자정을 넘어가는 경우
		if previousSchedule.End.Hour < 14 {
			return currentMinutes <= previousDayEndMinutes
		}
		return false
	}

	// 자정 이전 시간대(12시~24시)인 경우
	return currentMinutes >= currentDayStartMinutes
}

// UpdateDaySchedule updates the schedule for a specific day
func (s *Schedule) UpdateDaySchedule(day string, start, end TimePoint) error {
	if _, exists := s.schedule.Schedules[day]; !exists {
		return fmt.Errorf("invalid day: %s", day)
	}

	// Validate time points
	if start.Hour < 0 || start.Hour > 23 || end.Hour < 0 || end.Hour > 23 ||
		start.Minute < 0 || start.Minute > 59 || end.Minute < 0 || end.Minute > 59 {
		return fmt.Errorf("invalid time values")
	}

	s.schedule.Schedules[day] = DaySchedule{Start: start, End: end}
	return nil
}
