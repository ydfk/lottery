package lottery

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	model "go-fiber-starter/internal/model/lottery"
)

type drawSchedule struct {
	weekdays map[time.Weekday]struct{}
	hour     int
	minute   int
	location *time.Location
}

func buildRecommendationPlan(definition Definition, history []model.DrawResult, now time.Time) (string, time.Time, error) {
	if len(history) == 0 {
		return "", time.Time{}, fmt.Errorf("%s 缺少历史开奖，无法推断推荐期号，请先同步开奖历史", definition.Name)
	}

	schedule, err := parseDrawSchedule(definition)
	if err != nil {
		return "", time.Time{}, err
	}

	latestDraw := findLatestDraw(history)
	latestIssue := normalizeIssueByCode(definition.Code, latestDraw.Issue)
	if latestIssue == "" {
		return "", time.Time{}, fmt.Errorf("%s 缺少最近期开奖期号，无法推断推荐目标", definition.Name)
	}

	latestDrawAt := schedule.atDate(latestDraw.DrawDate)
	targetDrawAt, err := nextScheduledDraw(now, schedule)
	if err != nil {
		return "", time.Time{}, err
	}
	if !targetDrawAt.After(latestDrawAt) {
		targetDrawAt, err = nextScheduledDraw(latestDrawAt.Add(time.Second), schedule)
		if err != nil {
			return "", time.Time{}, err
		}
	}

	issueOffset := countScheduledDrawsBetween(latestDrawAt, targetDrawAt, schedule)
	if issueOffset <= 0 {
		return "", time.Time{}, fmt.Errorf("%s 无法推断下一期推荐期号", definition.Name)
	}

	targetIssue, err := addIssueOffset(latestIssue, issueOffset)
	if err != nil {
		return "", time.Time{}, err
	}
	return targetIssue, targetDrawAt, nil
}

func parseDrawSchedule(definition Definition) (drawSchedule, error) {
	if len(definition.DrawSchedule.Weekdays) == 0 {
		return drawSchedule{}, fmt.Errorf("%s 未配置开奖日 weekDays", definition.Name)
	}

	parts := strings.Split(strings.TrimSpace(definition.DrawSchedule.Time), ":")
	if len(parts) != 2 {
		return drawSchedule{}, fmt.Errorf("%s 开奖时间配置不正确", definition.Name)
	}

	hour, err := strconv.Atoi(parts[0])
	if err != nil || hour < 0 || hour > 23 {
		return drawSchedule{}, fmt.Errorf("%s 开奖小时配置不正确", definition.Name)
	}
	minute, err := strconv.Atoi(parts[1])
	if err != nil || minute < 0 || minute > 59 {
		return drawSchedule{}, fmt.Errorf("%s 开奖分钟配置不正确", definition.Name)
	}

	weekdays := make(map[time.Weekday]struct{}, len(definition.DrawSchedule.Weekdays))
	for _, item := range definition.DrawSchedule.Weekdays {
		if item < 0 || item > 6 {
			return drawSchedule{}, fmt.Errorf("%s 开奖星期配置不正确", definition.Name)
		}
		weekdays[time.Weekday(item)] = struct{}{}
	}

	return drawSchedule{
		weekdays: weekdays,
		hour:     hour,
		minute:   minute,
		location: time.Local,
	}, nil
}

func findLatestDraw(history []model.DrawResult) model.DrawResult {
	latest := history[0]
	for _, item := range history[1:] {
		if item.DrawDate.After(latest.DrawDate) {
			latest = item
			continue
		}
		if item.DrawDate.Equal(latest.DrawDate) && normalizeIssueByCode(item.LotteryCode, item.Issue) > normalizeIssueByCode(latest.LotteryCode, latest.Issue) {
			latest = item
		}
	}
	return latest
}

func nextScheduledDraw(now time.Time, schedule drawSchedule) (time.Time, error) {
	current := now.In(schedule.location)
	for offset := 0; offset <= 7; offset++ {
		date := current.AddDate(0, 0, offset)
		candidate := schedule.atDate(date)
		if !schedule.matches(candidate) {
			continue
		}
		if offset > 0 || !current.After(candidate) {
			return candidate, nil
		}
	}
	return time.Time{}, fmt.Errorf("无法找到下一个开奖时间")
}

func countScheduledDrawsBetween(from time.Time, to time.Time, schedule drawSchedule) int {
	if !to.After(from) {
		return 0
	}

	start := time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, schedule.location)
	end := time.Date(to.Year(), to.Month(), to.Day(), 0, 0, 0, 0, schedule.location)
	count := 0
	for date := start; !date.After(end); date = date.AddDate(0, 0, 1) {
		candidate := schedule.atDate(date)
		if !schedule.matches(candidate) {
			continue
		}
		if candidate.After(from) && !candidate.After(to) {
			count++
		}
	}
	return count
}

func addIssueOffset(issue string, offset int) (string, error) {
	current, err := strconv.Atoi(issue)
	if err != nil {
		return "", fmt.Errorf("推荐期号格式不正确: %s", issue)
	}
	return fmt.Sprintf("%0*d", len(issue), current+offset), nil
}

func (schedule drawSchedule) matches(value time.Time) bool {
	_, exists := schedule.weekdays[value.In(schedule.location).Weekday()]
	return exists
}

func (schedule drawSchedule) atDate(value time.Time) time.Time {
	current := value.In(schedule.location)
	return time.Date(current.Year(), current.Month(), current.Day(), schedule.hour, schedule.minute, 0, 0, schedule.location)
}
