# Drive2 — kế hoạch hoàn chỉnh (đồng bộ spec + Bruno)

Tài liệu này là **plan nội bộ module** (bổ sung khi không dùng `ai-prompts/`). Trạng thái cập nhật theo commit hiện tại.

## Đã có

- Module `drive2` trong `loader/module_static.go`; `Init()` = adapter + repo + app + transport; `RegisterModels` (file + share).
- REST nhóm `{{parent}}/v1/drive2` (`constants.RestV1Drive2Prefix`), file + share (gồm `shares/search`, `users/:user_id/file-shares`).
- `DriveFileShareService`: CRUD/list/search; **aggregate** `GetDriveFileAncestorOwnersByFileId`, `GetDriveFileResolvedSharesByFileId`, `GetDriveFileUserShareDetails`, `GetDriveFileShareByUser` (logic port từ `drive`, SQL resolved giống `dri_file_shares`).
- Repo `ListResolvedByFileRefs` (SQL rank + phân trang).

## Việc còn lại (ngoài scope lần sửa này hoặc cần rà thêm)

- **SearchAccessible** / search shared: predicate đầy đủ như `drive` nếu vẫn thiếu edge case.
- **Delete file** / trash: so khớp từng bước với `drive` (pending-delete, partial fail) nếu product bắt buộc.
- **CQRS transport** `modules/drive2/transport/cqrs` + đăng ký handler (nếu vẫn dùng bus như `drive`).
- **Signed URL + Redis** (`drive_file_signed_url_service.go`).
- **Background** purge trash (consumer/scheduler).
- **Bruno**: mở rộng thêm case lỗi validate / multipart giống bộ `Drive` (hiện mới skeleton `Drive2`).

## Bruno

Thư mục: `scripts/api/NikkiERP - Tests/Drive2/` — URL dùng prefix `/v1/drive2`, biến `{{drive2_file_id}}` (tự set trong môi trường Bruno).
