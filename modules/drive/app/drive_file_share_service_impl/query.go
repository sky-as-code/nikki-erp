package drive_file_share_service_impl

import (
	"sort"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/orm"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
	"github.com/sky-as-code/nikki-erp/modules/drive/domain"
	"github.com/sky-as-code/nikki-erp/modules/drive/enum"
	it "github.com/sky-as-code/nikki-erp/modules/drive/interfaces/drive_file_share"
)

func (this *DriveFileShareServiceImpl) GetDriveFileShareById(ctx crud.Context, query it.GetDriveFileShareByIdQuery) (
	*it.GetDriveFileShareByIdResult, error) {
	return crud.GetOne(ctx, crud.GetOneParam[*domain.DriveFileShare, it.GetDriveFileShareByIdQuery, it.GetDriveFileShareByIdResult]{
		Action: "get drive file share by id",
		Query:  query,
		RepoFindOne: func(ctx crud.Context, q it.GetDriveFileShareByIdQuery, vErrs *ft.ValidationErrors) (*domain.DriveFileShare, error) {
			share, err := this.driveFileShareRepo.FindById(ctx, q.DriveFileShareId)
			if err != nil {
				return nil, err
			}
			if share == nil {
				vErrs.AppendNotFound("driveFileShareId", "drive file share")
			}
			if share != nil {
				if err := this.enrichDriveFileSharesWithViews(ctx, []*domain.DriveFileShare{share}); err != nil {
					return nil, err
				}
			}
			return share, nil
		},
		ToFailureResult: func(vErrs *ft.ValidationErrors) *it.GetDriveFileShareByIdResult {
			return &it.GetDriveFileShareByIdResult{ClientError: vErrs.ToClientError()}
		},
		ToSuccessResult: func(d *domain.DriveFileShare) *it.GetDriveFileShareByIdResult {
			return &it.GetDriveFileShareByIdResult{HasData: true, Data: d}
		},
	})
}

// Get by File ID
func (this *DriveFileShareServiceImpl) GetDriveFileShareByFileId(ctx crud.Context, query it.GetDriveFileShareByFileIdQuery) (
	*it.GetDriveFileShareByFileIdResult, error) {
	return crud.Search(ctx, crud.SearchParam[*domain.DriveFileShare, it.GetDriveFileShareByFileIdQuery, it.GetDriveFileShareByFileIdResult]{
		Action: "get drive files by parent",
		Query:  query,
		SetQueryDefaults: func(q *it.GetDriveFileShareByFileIdQuery) {
			q.SetDefaults()
		},
		ParseSearchGraph: this.driveFileRepo.ParseSearchGraph,
		RepoSearch: func(ctx crud.Context, q it.GetDriveFileShareByFileIdQuery, predicate *orm.Predicate, order []orm.OrderOption) (*crud.PagedResult[*domain.DriveFileShare], error) {
			paged, err := this.driveFileShareRepo.ListByFileRef(ctx, it.ListByFileRefParam{
				FileRef: q.DriveFileId,
				SearchParam: it.SearchParam{
					Predicate: predicate,
					Order:     order,
					Page:      *q.Page,
					Size:      *q.Size,
				},
			})
			if err != nil {
				return nil, err
			}
			if paged != nil {
				if err := this.enrichDriveFileSharesWithViews(ctx, paged.Items); err != nil {
					return nil, err
				}
			}
			return paged, nil
		},
		ToFailureResult: func(vErrs *ft.ValidationErrors) *it.GetDriveFileShareByFileIdResult {
			return &it.GetDriveFileShareByFileIdResult{ClientError: vErrs.ToClientError()}
		},
		ToSuccessResult: func(paged *crud.PagedResult[*domain.DriveFileShare]) *it.GetDriveFileShareByFileIdResult {
			return &it.GetDriveFileShareByFileIdResult{Data: paged, HasData: true}
		},
	})
}

func (this *DriveFileShareServiceImpl) GetDriveFileAncestorOwnersByFileId(ctx crud.Context, query it.GetDriveFileAncestorOwnersByFileIdQuery) (
	*it.GetDriveFileAncestorOwnersByFileIdResult, error) {
	vErrs := query.Validate()
	if vErrs.Count() > 0 {
		return &it.GetDriveFileAncestorOwnersByFileIdResult{ClientError: vErrs.ToClientError()}, nil
	}

	file, err := this.driveFileRepo.FindById(ctx, query.DriveFileId)
	if err != nil {
		return nil, err
	}

	// Exclude the owner of the target file itself (the "file owner" shouldn't appear in this endpoint).
	excludeOwnerRef := (*model.Id)(nil)
	if file != nil && file.OwnerRef != nil && *file.OwnerRef != "" {
		excludeOwnerRef = file.OwnerRef
	}

	ancestors, err := this.driveFileRepo.GetDriveFileParents(ctx, query.DriveFileId)
	if err != nil {
		return nil, err
	}

	ownerByUser := map[model.Id]*domain.DriveFileShare{}
	for _, ancestor := range ancestors {
		if ancestor == nil || ancestor.OwnerRef == nil || *ancestor.OwnerRef == "" {
			continue
		}
		if excludeOwnerRef != nil && *ancestor.OwnerRef == *excludeOwnerRef {
			continue
		}
		ownerByUser[*ancestor.OwnerRef] = &domain.DriveFileShare{
			FileRef:    query.DriveFileId,
			UserRef:    *ancestor.OwnerRef,
			Permission: enum.DriveFilePermAncestorOwner,
		}
	}

	items := make([]*domain.DriveFileShare, 0, len(ownerByUser))
	for _, s := range ownerByUser {
		items = append(items, s)
	}
	if err := this.enrichDriveFileSharesWithViews(ctx, items); err != nil {
		return nil, err
	}

	return &it.GetDriveFileAncestorOwnersByFileIdResult{
		HasData: true,
		Data:    items,
	}, nil
}

func (this *DriveFileShareServiceImpl) GetDriveFileResolvedSharesByFileId(ctx crud.Context, query it.GetDriveFileResolvedSharesByFileIdQuery) (
	*it.GetDriveFileResolvedSharesByFileIdResult, error) {
	query.SetDefaults()
	vErrs := query.Validate()
	if vErrs.Count() > 0 {
		return &it.GetDriveFileResolvedSharesByFileIdResult{ClientError: vErrs.ToClientError()}, nil
	}

	file, err := this.driveFileRepo.FindById(ctx, query.DriveFileId)
	if err != nil {
		return nil, err
	}
	ancestors, err := this.driveFileRepo.GetDriveFileParents(ctx, query.DriveFileId)
	if err != nil {
		return nil, err
	}

	refSet := map[model.Id]struct{}{query.DriveFileId: {}}
	excludedSet := map[model.Id]struct{}{}
	if file != nil && file.OwnerRef != nil && *file.OwnerRef != "" {
		excludedSet[*file.OwnerRef] = struct{}{}
	}
	for _, ancestor := range ancestors {
		if ancestor == nil || ancestor.Id == nil {
			continue
		}
		refSet[*ancestor.Id] = struct{}{}
		if ancestor.OwnerRef != nil && *ancestor.OwnerRef != "" {
			excludedSet[*ancestor.OwnerRef] = struct{}{}
		}
	}
	refs := make([]model.Id, 0, len(refSet))
	for id := range refSet {
		refs = append(refs, id)
	}
	excludedUsers := make([]model.Id, 0, len(excludedSet))
	for id := range excludedSet {
		excludedUsers = append(excludedUsers, id)
	}

	paged, err := this.driveFileShareRepo.ListResolvedByFileRefs(
		ctx,
		query.DriveFileId,
		refs,
		excludedUsers,
		*query.Page,
		*query.Size,
	)
	if err != nil {
		return nil, err
	}
	if paged != nil {
		if err := this.enrichDriveFileSharesWithViews(ctx, paged.Items); err != nil {
			return nil, err
		}
	}

	return &it.GetDriveFileResolvedSharesByFileIdResult{
		HasData: true,
		Data:    paged,
	}, nil
}

func (this *DriveFileShareServiceImpl) GetDriveFileUserShareDetails(ctx crud.Context, query it.GetDriveFileUserShareDetailsQuery) (
	*it.GetDriveFileUserShareDetailsResult, error) {
	vErrs := query.Validate()
	if vErrs.Count() > 0 {
		return &it.GetDriveFileUserShareDetailsResult{ClientError: vErrs.ToClientError()}, nil
	}

	file, err := this.driveFileRepo.FindById(ctx, query.DriveFileId)
	if err != nil {
		return nil, err
	}
	ancestors, err := this.driveFileRepo.GetDriveFileParents(ctx, query.DriveFileId)
	if err != nil {
		return nil, err
	}
	out := make([]*it.DriveFileUserShareDetail, 0, 8)

	// ---- New semantics ----
	// Output should show:
	// 1) Direct share + owner entries for ancestor files (excluding the target file)
	// 2) Effective permission for the target file itself
	// Ordered by ancestor chain: root -> ... -> closest ancestor -> target.

	storedSharePerm := func(p enum.DriveFilePerm) enum.DriveFilePerm {
		switch p {
		case enum.DriveFilePermView, enum.DriveFilePermEdit, enum.DriveFilePermEditTrash:
			return p
		default:
			return enum.DriveFilePermNone
		}
	}

	toInheritedPerm := func(base enum.DriveFilePerm) enum.DriveFilePerm {
		switch base {
		case enum.DriveFilePermView:
			return enum.DriveFilePermInheritedView
		case enum.DriveFilePermEdit:
			return enum.DriveFilePermInheritedEdit
		case enum.DriveFilePermEditTrash:
			return enum.DriveFilePermInheritedEditTrash
		default:
			return enum.DriveFilePermNone
		}
	}

	// Build full chain root -> ... -> target (inclusive).
	idToAncestor := make(map[model.Id]*domain.DriveFile, len(ancestors))
	for _, a := range ancestors {
		if a == nil || a.Id == nil {
			continue
		}
		idToAncestor[*a.Id] = a
	}

	chainChildToRoot := make([]*domain.DriveFile, 0, len(ancestors)+1)
	chainChildToRoot = append(chainChildToRoot, file) // start at target
	cur := file
	for cur != nil && cur.ParentDriveFileRef != nil && *cur.ParentDriveFileRef != "" {
		parent := idToAncestor[*cur.ParentDriveFileRef]
		if parent == nil {
			break
		}
		chainChildToRoot = append(chainChildToRoot, parent)
		cur = parent
	}

	chainRootToChild := make([]*domain.DriveFile, 0, len(chainChildToRoot))
	for i := len(chainChildToRoot) - 1; i >= 0; i-- {
		chainRootToChild = append(chainRootToChild, chainChildToRoot[i])
	}

	chainIndex := make(map[model.Id]int, len(chainRootToChild))
	for i, n := range chainRootToChild {
		if n == nil || n.Id == nil {
			continue
		}
		chainIndex[*n.Id] = i
	}

	// Collect ancestor file ids (exclude target).
	ancestorIds := make([]model.Id, 0, len(chainRootToChild))
	ownerAncestorSet := make(map[model.Id]struct{}, len(chainRootToChild))
	for _, n := range chainRootToChild {
		if n == nil || n.Id == nil {
			continue
		}
		if *n.Id == query.DriveFileId {
			continue
		}
		ancestorIds = append(ancestorIds, *n.Id)
		if n.OwnerRef != nil && *n.OwnerRef == query.UserId {
			ownerAncestorSet[*n.Id] = struct{}{}
		}
	}

	// Stored direct shares on ancestor nodes (not inherited).
	shareByFileRef := make(map[model.Id]*domain.DriveFileShare, len(ancestorIds))
	if len(ancestorIds) > 0 {
		ancShares, err := this.driveFileShareRepo.ListByFileRefsAndUserRef(ctx, ancestorIds, query.UserId)
		if err != nil {
			return nil, err
		}
		for _, s := range ancShares {
			if s == nil {
				continue
			}
			p := storedSharePerm(s.Permission)
			if p == enum.DriveFilePermNone {
				continue
			}
			// DB has max 1 row per (file_ref, user_ref); keep the first.
			if _, ok := shareByFileRef[s.FileRef]; !ok {
				shareByFileRef[s.FileRef] = s
			}
		}
	}

	// Direct share on target (stored permission).
	var directTargetShare *domain.DriveFileShare
	targetShares, err := this.driveFileShareRepo.ListByFileRefsAndUserRef(ctx, []model.Id{query.DriveFileId}, query.UserId)
	if err != nil {
		return nil, err
	}
	for _, s := range targetShares {
		if s == nil {
			continue
		}
		directTargetShare = s
		if storedSharePerm(s.Permission) != enum.DriveFilePermNone {
			break
		}
	}

	// Effective permission for the target file.
	targetPerm := enum.DriveFilePermNone
	var targetSourceShare *domain.DriveFileShare
	if file != nil && file.OwnerRef != nil && *file.OwnerRef == query.UserId {
		targetPerm = enum.DriveFilePermOwner
	} else if len(ownerAncestorSet) > 0 {
		targetPerm = enum.DriveFilePermAncestorOwner
	} else if directTargetShare != nil {
		p := storedSharePerm(directTargetShare.Permission)
		if p != enum.DriveFilePermNone {
			targetPerm = p
			targetSourceShare = directTargetShare
		}
	} else if len(shareByFileRef) > 0 {
		maxBase := enum.DriveFilePermNone
		var bestShare *domain.DriveFileShare
		for _, sh := range shareByFileRef {
			if sh == nil {
				continue
			}
			p := storedSharePerm(sh.Permission)
			if p > maxBase {
				maxBase = p
				bestShare = sh
			}
		}
		if maxBase != enum.DriveFilePermNone {
			targetPerm = toInheritedPerm(maxBase)
			targetSourceShare = bestShare
		}
	}

	// Build output entries.
	seen := make(map[model.Id]struct{}, len(chainRootToChild))

	// 1) Ancestor entries (exclude target):
	//    - if user owns ancestor => permission=owner
	//    - else if ancestor has direct share => permission=view|edit|edit-trash
	for _, n := range chainRootToChild {
		if n == nil || n.Id == nil {
			continue
		}
		id := *n.Id
		if id == query.DriveFileId {
			continue
		}

		if _, ok := ownerAncestorSet[id]; ok {
			out = append(out, &it.DriveFileUserShareDetail{
				FileRef:    id,
				UserRef:    query.UserId,
				Permission: enum.DriveFilePermOwner,
			})
			seen[id] = struct{}{}
			continue
		}

		if sh, ok := shareByFileRef[id]; ok && sh != nil {
			p := storedSharePerm(sh.Permission)
			if p == enum.DriveFilePermNone {
				continue
			}
			out = append(out, &it.DriveFileUserShareDetail{
				FileRef:    id,
				UserRef:    query.UserId,
				Permission: p,
				ModelBase:  sh.ModelBase,
				AuditableBase: sh.AuditableBase,
			})
			seen[id] = struct{}{}
		}
	}

	// 2) Target entry (effective perm):
	if targetPerm != enum.DriveFilePermNone {
		if _, ok := seen[query.DriveFileId]; !ok {
			detail := &it.DriveFileUserShareDetail{
				FileRef:    query.DriveFileId,
				UserRef:    query.UserId,
				Permission: targetPerm,
			}
			if targetSourceShare != nil {
				detail.ModelBase = targetSourceShare.ModelBase
				detail.AuditableBase = targetSourceShare.AuditableBase
			}
			out = append(out, detail)
		}
	}

	// enrich user once
	usersById, errLookup, _ := this.identityCqrs.GetUsersByIds(ctx, []model.Id{query.UserId})
	if errLookup != nil {
		return nil, errLookup
	}
	var userView *domain.DriveFileShareUser
	if usersById != nil {
		if u := usersById[query.UserId]; u != nil {
			userView = &domain.DriveFileShareUser{
				Id:          u.Id,
				DisplayName: u.DisplayName,
				Email:       u.Email,
				AvatarUrl:   u.AvatarUrl,
			}
		}
	}
	for _, item := range out {
		if item == nil {
			continue
		}
		item.User = userView
	}
	if err := this.enrichDriveFileUserShareDetailsWithFiles(ctx, out); err != nil {
		return nil, err
	}

	// Sort by ancestor chain order.
	sort.SliceStable(out, func(i, j int) bool {
		ii := chainIndex[out[i].FileRef]
		jj := chainIndex[out[j].FileRef]
		if ii != jj {
			return ii < jj
		}

		// tie-breaker: owner > ancestor-owner > direct > inherited
		prio := func(p enum.DriveFilePerm) int {
			switch p {
			case enum.DriveFilePermOwner:
				return 0
			case enum.DriveFilePermAncestorOwner:
				return 1
			case enum.DriveFilePermEditTrash, enum.DriveFilePermEdit, enum.DriveFilePermView:
				return 2
			default:
				return 3
			}
		}
		pi := prio(out[i].Permission)
		pj := prio(out[j].Permission)
		return pi < pj
	})

	return &it.GetDriveFileUserShareDetailsResult{
		HasData: true,
		Data:    out,
	}, nil
}

// Get by User

func (this *DriveFileShareServiceImpl) GetDriveFileShareByUser(ctx crud.Context, query it.GetDriveFileShareByUserQuery) (
	*it.GetDriveFileShareByUserResult, error) {
	vErrs := query.Validate()
	if vErrs.Count() > 0 {
		return &it.GetDriveFileShareByUserResult{ClientError: vErrs.ToClientError()}, nil
	}

	items, err := this.driveFileShareRepo.ListByUserRef(ctx, query.UserId)
	if err != nil {
		return nil, err
	}
	if err := this.enrichDriveFileSharesWithViews(ctx, items); err != nil {
		return nil, err
	}

	return &it.GetDriveFileShareByUserResult{
		HasData: items != nil,
		Data: &it.GetDriveFileShareByUserResultData{
			Items: items,
			Total: len(items),
		},
	}, nil
}

func (this *DriveFileShareServiceImpl) ListDriveFileSharesByFileRefsAndUser(ctx crud.Context, query it.ListDriveFileSharesByFileRefsAndUserQuery) (
	*it.ListDriveFileSharesByFileRefsAndUserResult, error) {
	vErrs := query.Validate()
	if vErrs.Count() > 0 {
		return &it.ListDriveFileSharesByFileRefsAndUserResult{ClientError: vErrs.ToClientError()}, nil
	}

	items, err := this.driveFileShareRepo.ListByFileRefsAndUserRef(ctx, query.DriveFileIds, query.UserId)
	if err != nil {
		return nil, err
	}
	if err := this.enrichDriveFileSharesWithViews(ctx, items); err != nil {
		return nil, err
	}

	return &it.ListDriveFileSharesByFileRefsAndUserResult{HasData: true, Data: items}, nil
}

// Search

func (this *DriveFileShareServiceImpl) SearchDriveFileShare(ctx crud.Context, query it.SearchDriveFileShareQuery) (
	*it.SearchDriveFileShareResult, error) {
	return crud.Search(ctx, crud.SearchParam[*domain.DriveFileShare, it.SearchDriveFileShareQuery, it.SearchDriveFileShareResult]{
		Action: "search drive file shares",
		Query:  query,
		SetQueryDefaults: func(q *it.SearchDriveFileShareQuery) {
			q.SetDefaults()
		},
		ParseSearchGraph: this.driveFileShareRepo.ParseSearchGraph,
		RepoSearch: func(ctx crud.Context, q it.SearchDriveFileShareQuery, predicate *orm.Predicate, order []orm.OrderOption) (*crud.PagedResult[*domain.DriveFileShare], error) {
			paged, err := this.driveFileShareRepo.Search(ctx, it.SearchParam{
				Predicate: predicate,
				Order:     order,
				Page:      *q.Page,
				Size:      *q.Size,
			})
			if err != nil {
				return nil, err
			}
			if paged != nil {
				if err := this.enrichDriveFileSharesWithViews(ctx, paged.Items); err != nil {
					return nil, err
				}
			}
			return paged, nil
		},
		ToFailureResult: func(vErrs *ft.ValidationErrors) *it.SearchDriveFileShareResult {
			return &it.SearchDriveFileShareResult{ClientError: vErrs.ToClientError()}
		},
		ToSuccessResult: func(paged *crud.PagedResult[*domain.DriveFileShare]) *it.SearchDriveFileShareResult {
			return &it.SearchDriveFileShareResult{Data: paged, HasData: paged.Items != nil}
		},
	})
}
