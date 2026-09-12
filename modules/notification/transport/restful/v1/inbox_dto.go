package v1

import (
	it "github.com/sky-as-code/nikki-erp/modules/notification/interfaces/delivery"
)

// InboxItemDto is one notification as the client sees it.
//
// is_read is calculated here rather than stored, so that it can never disagree with read_at
// (BR 7). Datetimes are strings, which is the DTO convention across this codebase.
type InboxItemDto struct {
	NotificationId     string         `json:"notification_id"`
	StreamSeq          int64          `json:"stream_seq"`
	Title              string         `json:"title"`
	Message            string         `json:"message"`
	Severity           string         `json:"severity"`
	SourceModule       string         `json:"source_module"`
	SourceResourceName *string        `json:"source_resource_name,omitempty"`
	SourceResourceKey  map[string]any `json:"source_resource_key,omitempty"`
	Metadata           map[string]any `json:"metadata,omitempty"`
	CreatedAt          string         `json:"created_at"`
	ReadAt             *string        `json:"read_at,omitempty"`
	IsRead             bool           `json:"is_read"`
}

type InboxPageDto struct {
	Items []InboxItemDto `json:"items"`

	// NextCursor is the stream_seq to pass back for the following page, absent on the last one.
	NextCursor *int64 `json:"next_cursor,omitempty"`
}

type UnreadCountDto struct {
	Count int `json:"count"`
}

// MarkReadDto reports what the call actually changed, which the caller cannot infer: ids owned by
// somebody else are counted as neither updated nor already read (BR 21).
type MarkReadDto struct {
	RequestedCount   int `json:"requested_count"`
	UpdatedCount     int `json:"updated_count"`
	AlreadyReadCount int `json:"already_read_count"`
}

func toInboxItemDto(item it.InboxItem) InboxItemDto {
	return InboxItemDto{
		NotificationId:     string(item.NotificationId),
		StreamSeq:          item.StreamSeq,
		Title:              item.Title,
		Message:            item.Message,
		Severity:           item.Severity,
		SourceModule:       item.SourceModule,
		SourceResourceName: item.SourceResourceName,
		SourceResourceKey:  item.SourceResourceKey,
		Metadata:           item.Metadata,
		CreatedAt:          item.CreatedAt,
		ReadAt:             item.ReadAt,
		IsRead:             item.IsRead(),
	}
}

func toInboxPageDto(data it.InboxResultData) any {
	items := make([]InboxItemDto, 0, len(data.Items))
	for _, item := range data.Items {
		items = append(items, toInboxItemDto(item))
	}
	return InboxPageDto{Items: items, NextCursor: data.NextCursor}
}

func toUnreadCountDto(data it.UnreadCountResultData) any {
	return UnreadCountDto{Count: data.Count}
}

func toMarkReadDto(data it.MarkReadResultData) any {
	return MarkReadDto{
		RequestedCount:   data.RequestedCount,
		UpdatedCount:     data.UpdatedCount,
		AlreadyReadCount: data.AlreadyReadCount,
	}
}
