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

type drawAnchor struct {
	issue  string
	drawAt time.Time
}

func buildRecommendationPlan(definition Definition, history []model.DrawResult, now time.Time) (string, time.Time, error) {
	schedule, err := parseDrawSchedule(definition)
	if err != nil {
		return "", time.Time{}, err
	}

	anchor, err := resolveScheduleAnchor(definition, history, schedule)
	if err != nil {
		return "", time.Time{}, err
	}

	return buildRecommendationPlanFromAnchor(definition, anchor.issue, anchor.drawAt, now)
}

func planRecommendationTarget(
	definition Definition,
	history []model.DrawResult,
	now time.Time,
) (string, time.Time, error) {
	return buildRecommendationPlan(definition, history, now)
}

func resolveLocalDrawByIssue(definition Definition, issue string) (time.Time, bool, error) {
	schedule, err := parseDrawSchedule(definition)
	if err != nil {
		return time.Time{}, false, err
	}

	anchor, ok, err := parseConfiguredAnchor(definition, schedule)
	if err != nil || !ok {
		return time.Time{}, ok, err
	}

	drawAt, err := resolveDrawAtByIssue(anchor, issue, schedule)
	if err != nil {
		return time.Time{}, false, err
	}
	return drawAt, true, nil
}

func resolveLatestLocalDraw(definition Definition, now time.Time) (string, time.Time, bool, error) {
	schedule, err := parseDrawSchedule(definition)
	if err != nil {
		return "", time.Time{}, false, err
	}

	anchor, ok, err := parseConfiguredAnchor(definition, schedule)
	if err != nil || !ok {
		return "", time.Time{}, ok, err
	}

	drawAt, err := latestScheduledDraw(now, schedule)
	if err != nil {
		return "", time.Time{}, false, err
	}

	issue, err := resolveIssueByDrawAt(anchor, drawAt, schedule)
	if err != nil {
		return "", time.Time{}, false, err
	}
	return issue, drawAt, true, nil
}

func buildRecommendationPlanFromAnchor(definition Definition, latestIssue string, latestDrawDate time.Time, now time.Time) (string, time.Time, error) {
	schedule, err := parseDrawSchedule(definition)
	if err != nil {
		return "", time.Time{}, err
	}

	targetDrawAt, err := nextScheduledDraw(now, schedule)
	if err != nil {
		return "", time.Time{}, err
	}

	latestDrawAt := schedule.atDate(latestDrawDate)
	if !targetDrawAt.After(latestDrawAt) {
		targetDrawAt, err = nextScheduledDraw(latestDrawAt.Add(time.Second), schedule)
		if err != nil {
			return "", time.Time{}, err
		}
	}

	targetIssue, err := addIssueOffsetBySchedule(latestIssue, countScheduledDrawsBetween(latestDrawAt, targetDrawAt, schedule), schedule)
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

func parseConfiguredAnchor(definition Definition, schedule drawSchedule) (drawAnchor, bool, error) {
	issue := normalizeIssueByCode(definition.Code, definition.DrawSchedule.AnchorIssue)
	dateText := strings.TrimSpace(definition.DrawSchedule.AnchorDate)
	if issue == "" || dateText == "" {
		return drawAnchor{}, false, nil
	}

	dateValue, err := time.ParseInLocation("2006-01-02", dateText, schedule.location)
	if err != nil {
		return drawAnchor{}, false, fmt.Errorf("%s 锚点开奖日期配置不正确", definition.Name)
	}

	return drawAnchor{
		issue:  issue,
		drawAt: schedule.atDate(dateValue),
	}, true, nil
}

func resolveScheduleAnchor(definition Definition, history []model.DrawResult, schedule drawSchedule) (drawAnchor, error) {
	if anchor, ok, err := parseConfiguredAnchor(definition, schedule); err != nil {
		return drawAnchor{}, err
	} else if ok {
		return anchor, nil
	}

	if len(history) == 0 {
		return drawAnchor{}, fmt.Errorf("%s 缺少本地锚点和历史开奖，无法推断期号", definition.Name)
	}

	latestDraw := findLatestDraw(history)
	latestIssue := normalizeIssueByCode(definition.Code, latestDraw.Issue)
	if latestIssue == "" || latestDraw.DrawDate.IsZero() {
		return drawAnchor{}, fmt.Errorf("%s 缺少有效历史开奖，无法推断期号", definition.Name)
	}

	return drawAnchor{
		issue:  latestIssue,
		drawAt: schedule.atDate(latestDraw.DrawDate),
	}, nil
}

func findLatestDraw(history []model.DrawResult) model.DrawResult {
	latest := history[0]
	for _, item := range history[1:] {
		currentIssue := normalizeIssueByCode(item.LotteryCode, item.Issue)
		latestIssue := normalizeIssueByCode(latest.LotteryCode, latest.Issue)
		if currentIssue > latestIssue {
			latest = item
			continue
		}
		if currentIssue == latestIssue && item.DrawDate.After(latest.DrawDate) {
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

func latestScheduledDraw(now time.Time, schedule drawSchedule) (time.Time, error) {
	current := now.In(schedule.location)
	for offset := 0; offset <= 7; offset++ {
		date := current.AddDate(0, 0, -offset)
		candidate := schedule.atDate(date)
		if !schedule.matches(candidate) {
			continue
		}
		if offset > 0 || !candidate.After(current) {
			return candidate, nil
		}
	}
	return time.Time{}, fmt.Errorf("无法找到最近一期开奖时间")
}

func previousScheduledDraw(now time.Time, schedule drawSchedule) (time.Time, error) {
	current := now.In(schedule.location)
	for offset := 0; offset <= 7; offset++ {
		date := current.AddDate(0, 0, -offset)
		candidate := schedule.atDate(date)
		if !schedule.matches(candidate) {
			continue
		}
		if offset > 0 || candidate.Before(current) {
			return candidate, nil
		}
	}
	return time.Time{}, fmt.Errorf("无法找到上一期开奖时间")
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

func addIssueOffsetBySchedule(issue string, offset int, schedule drawSchedule) (string, error) {
	year, sequence, seqWidth, err := parseIssueParts(issue)
	if err != nil {
		return "", err
	}

	if offset == 0 {
		return formatIssue(year, sequence, seqWidth), nil
	}

	if offset > 0 {
		remain := offset
		for remain > 0 {
			yearDrawCount := countScheduleDrawsInYear(year, schedule)
			available := yearDrawCount - sequence
			if remain <= available {
				sequence += remain
				return formatIssue(year, sequence, seqWidth), nil
			}
			remain -= available + 1
			year++
			sequence = 1
		}
	}

	remain := -offset
	for remain > 0 {
		available := sequence - 1
		if remain <= available {
			sequence -= remain
			return formatIssue(year, sequence, seqWidth), nil
		}
		remain -= available + 1
		year--
		sequence = countScheduleDrawsInYear(year, schedule)
	}
	return formatIssue(year, sequence, seqWidth), nil
}

func resolveDrawAtByIssue(anchor drawAnchor, issue string, schedule drawSchedule) (time.Time, error) {
	offset, err := diffIssueBySchedule(anchor.issue, normalizeIssueByCode("", issue), schedule)
	if err != nil {
		return time.Time{}, err
	}
	return shiftScheduledDraw(anchor.drawAt, offset, schedule)
}

func resolveIssueByDrawAt(anchor drawAnchor, drawAt time.Time, schedule drawSchedule) (string, error) {
	offset := diffScheduledDraw(anchor.drawAt, drawAt, schedule)
	return addIssueOffsetBySchedule(anchor.issue, offset, schedule)
}

func shiftScheduledDraw(drawAt time.Time, offset int, schedule drawSchedule) (time.Time, error) {
	current := schedule.atDate(drawAt)
	if offset == 0 {
		return current, nil
	}

	var err error
	if offset > 0 {
		for step := 0; step < offset; step++ {
			current, err = nextScheduledDraw(current.Add(time.Second), schedule)
			if err != nil {
				return time.Time{}, err
			}
		}
		return current, nil
	}

	for step := 0; step < -offset; step++ {
		current, err = previousScheduledDraw(current, schedule)
		if err != nil {
			return time.Time{}, err
		}
	}
	return current, nil
}

func diffScheduledDraw(from time.Time, to time.Time, schedule drawSchedule) int {
	if to.Equal(from) {
		return 0
	}
	if to.After(from) {
		return countScheduledDrawsBetween(from, to, schedule)
	}
	return -countScheduledDrawsBetween(to, from, schedule)
}

func diffIssueBySchedule(fromIssue string, toIssue string, schedule drawSchedule) (int, error) {
	fromYear, fromSequence, _, err := parseIssueParts(fromIssue)
	if err != nil {
		return 0, err
	}
	toYear, toSequence, _, err := parseIssueParts(toIssue)
	if err != nil {
		return 0, err
	}

	if fromYear == toYear {
		return toSequence - fromSequence, nil
	}

	if toYear > fromYear {
		total := countScheduleDrawsInYear(fromYear, schedule) - fromSequence
		for year := fromYear + 1; year < toYear; year++ {
			total += countScheduleDrawsInYear(year, schedule)
		}
		total += toSequence
		return total, nil
	}

	total := fromSequence - 1
	for year := fromYear - 1; year > toYear; year-- {
		total += countScheduleDrawsInYear(year, schedule)
	}
	total += countScheduleDrawsInYear(toYear, schedule) - toSequence + 1
	return -total, nil
}

func parseIssueParts(issue string) (int, int, int, error) {
	canonical := strings.TrimSpace(issue)
	if len(canonical) < 7 || !isDigits(canonical) {
		return 0, 0, 0, fmt.Errorf("期号格式不正确: %s", issue)
	}

	year, err := strconv.Atoi(canonical[:4])
	if err != nil {
		return 0, 0, 0, fmt.Errorf("期号年份格式不正确: %s", issue)
	}
	sequence, err := strconv.Atoi(canonical[4:])
	if err != nil {
		return 0, 0, 0, fmt.Errorf("期号序号格式不正确: %s", issue)
	}
	return year, sequence, len(canonical) - 4, nil
}

func formatIssue(year int, sequence int, seqWidth int) string {
	return fmt.Sprintf("%04d%0*d", year, seqWidth, sequence)
}

func countScheduleDrawsInYear(year int, schedule drawSchedule) int {
	start := time.Date(year, 1, 1, 0, 0, 0, 0, schedule.location)
	end := time.Date(year, 12, 31, 0, 0, 0, 0, schedule.location)
	count := 0
	for date := start; !date.After(end); date = date.AddDate(0, 0, 1) {
		if schedule.matches(schedule.atDate(date)) {
			count++
		}
	}
	return count
}

func (schedule drawSchedule) matches(value time.Time) bool {
	_, exists := schedule.weekdays[value.In(schedule.location).Weekday()]
	return exists
}

func (schedule drawSchedule) atDate(value time.Time) time.Time {
	current := value.In(schedule.location)
	return time.Date(current.Year(), current.Month(), current.Day(), schedule.hour, schedule.minute, 0, 0, schedule.location)
}
