package promotion

import (
	"sort"
)

type giftKey struct {
	promotionId string
	rewardId    string
}

// reconcileGifts turns applied gifts into upsert intents (BR-R09): an existing gift line of the same
// promotion and reward is updated, never duplicated; every other existing gift line is removed,
// which is how a gift disappears from a draft that no longer qualifies.
func reconcileGifts(input *Input, programs []ProgramResult) ([]GiftIntent, []string) {
	existing := existingGiftLines(input)
	claimed := map[string]bool{}
	gifts := []GiftIntent{}
	for i := range programs {
		gift := programs[i].Gift
		if programs[i].Outcome != OutcomeApplied || gift == nil {
			continue
		}
		key := giftKey{promotionId: gift.PromotionId, rewardId: gift.RewardId}
		if lines := existing[key]; len(lines) > 0 {
			gift.ExistingLineKey = lines[0].Key
			claimed[lines[0].Key] = true
		}
		gifts = append(gifts, *gift)
	}

	removed := []string{}
	for _, line := range input.Lines {
		if line.LineType == LineTypePromotionGift && !claimed[line.Key] {
			removed = append(removed, line.Key)
		}
	}
	sort.Strings(removed)
	return gifts, removed
}

// existingGiftLines groups the input's gift lines by source, each group in line order, so the line
// kept on a duplicate is always the same one.
func existingGiftLines(input *Input) map[giftKey][]Line {
	grouped := map[giftKey][]Line{}
	for _, line := range input.Lines {
		if line.LineType != LineTypePromotionGift {
			continue
		}
		key := giftKey{promotionId: line.SourcePromotionId, rewardId: line.SourceRewardId}
		grouped[key] = append(grouped[key], line)
	}
	for key := range grouped {
		lines := grouped[key]
		sort.SliceStable(lines, func(i, j int) bool { return lineLess(lines[i], lines[j]) })
	}
	return grouped
}
