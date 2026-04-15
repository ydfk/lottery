package lottery

import (
	"fmt"
	"slices"
	"strings"
	"time"

	model "go-fiber-starter/internal/model/lottery"
	"go-fiber-starter/pkg/db"

	"github.com/google/uuid"
)

type importGroupPayload struct {
	LotteryCode      string
	RecommendationID string
	Issue            string
	DrawDate         time.Time
	PurchasedAt      time.Time
	CostAmount       float64
	Notes            string
	ImagePath        string
	Entries          []ParsedEntry
}

func buildImportGroupKey(lotteryCode string, issue string) string {
	normalizedCode := strings.ToLower(strings.TrimSpace(lotteryCode))
	return normalizedCode + ":" + normalizeIssueByCode(normalizedCode, issue)
}

func buildImportGroupPayload(group importTicketGroup, images map[string]string) (importGroupPayload, error) {
	first := group.Items[0]
	payload := importGroupPayload{
		LotteryCode: first.Input.LotteryCode,
		Issue:       first.Input.Issue,
		Entries:     make([]ParsedEntry, 0, len(group.Items)),
	}

	notes := make([]string, 0, len(group.Items))
	imageNames := make([]string, 0, len(group.Items))
	recommendationIDs := make([]string, 0, len(group.Items))

	for _, item := range group.Items {
		payload.Entries = append(payload.Entries, item.Entry)
		payload.CostAmount += resolveImportedEntryCost(item.Input, item.Entry)

		if item.Input.DrawDate.IsZero() {
			continue
		}
		if payload.DrawDate.IsZero() {
			payload.DrawDate = item.Input.DrawDate
		} else if !sameDay(payload.DrawDate, item.Input.DrawDate) {
			return payload, fmt.Errorf("同一彩种同一期号的开奖日期不一致，请检查后重试")
		}
	}

	for _, item := range group.Items {
		if !item.Input.PurchasedAt.IsZero() {
			if payload.PurchasedAt.IsZero() || item.Input.PurchasedAt.Before(payload.PurchasedAt) {
				payload.PurchasedAt = item.Input.PurchasedAt
			}
		}

		if note := strings.TrimSpace(item.Input.Notes); note != "" && !slices.Contains(notes, note) {
			notes = append(notes, note)
		}

		if imageName := strings.TrimSpace(item.Input.ImageName); imageName != "" && !slices.Contains(imageNames, imageName) {
			imageNames = append(imageNames, imageName)
		}

		if recommendationID := strings.TrimSpace(item.Input.RecommendationID); recommendationID != "" && !slices.Contains(recommendationIDs, recommendationID) {
			recommendationIDs = append(recommendationIDs, recommendationID)
		}
	}

	if len(recommendationIDs) > 1 {
		return payload, fmt.Errorf("同一彩种同一期号存在多个推荐记录，请拆分后导入")
	}
	if len(recommendationIDs) == 1 {
		payload.RecommendationID = recommendationIDs[0]
	}

	if len(imageNames) > 0 {
		path, err := resolveImportedImagePath(imageNames[0], images)
		if err != nil {
			return payload, err
		}
		payload.ImagePath = path
		if len(imageNames) > 1 {
			notes = append(notes, "额外图片未关联: "+strings.Join(imageNames[1:], ", "))
		}
	}

	payload.Notes = strings.Join(notes, "\n")
	return payload, nil
}

func resolveImportedEntryCost(row importTicketRow, entry ParsedEntry) float64 {
	if row.CostAmount > 0 {
		return row.CostAmount
	}
	return calculateEntriesCost([]ParsedEntry{entry})
}

func sameDay(left time.Time, right time.Time) bool {
	return left.Year() == right.Year() && left.Month() == right.Month() && left.Day() == right.Day()
}

func matchImportRecommendation(userID string, lotteryCode string, issue string, entries []ParsedEntry) (*uuid.UUID, error) {
	if len(entries) == 0 {
		return nil, nil
	}

	recommendations := make([]model.Recommendation, 0)
	if err := currentUserScope(db.DB.Preload("Entries"), userID).
		Where("lottery_code = ? AND issue IN ?", lotteryCode, issueAliases(lotteryCode, issue)).
		Order("created_at DESC").
		Find(&recommendations).Error; err != nil {
		return nil, err
	}

	if len(recommendations) == 0 {
		return nil, nil
	}

	entrySet := buildImportEntrySet(entries)
	var matched *model.Recommendation
	for index := range recommendations {
		recommendation := &recommendations[index]
		if !recommendationEntriesMatch(entrySet, recommendation.Entries) {
			continue
		}
		if matched == nil || len(recommendation.Entries) > len(matched.Entries) {
			matched = recommendation
		}
	}

	if matched == nil {
		return nil, nil
	}
	return &matched.Id, nil
}

func buildImportEntrySet(entries []ParsedEntry) map[string]struct{} {
	result := make(map[string]struct{}, len(entries))
	for _, entry := range entries {
		result[buildImportEntryKey(formatNumbers(entry.Red), formatNumbers(entry.Blue))] = struct{}{}
	}
	return result
}

func recommendationEntriesMatch(entrySet map[string]struct{}, entries []model.RecommendationEntry) bool {
	if len(entries) == 0 {
		return false
	}

	for _, entry := range entries {
		key := buildImportEntryKey(normalizeImportedNumbers(entry.RedNumbers), normalizeImportedNumbers(entry.BlueNumbers))
		if _, ok := entrySet[key]; !ok {
			return false
		}
	}
	return true
}

func buildImportEntryKey(redNumbers string, blueNumbers string) string {
	return normalizeImportedNumbers(strings.TrimSpace(redNumbers)) + "+" + normalizeImportedNumbers(strings.TrimSpace(blueNumbers))
}
