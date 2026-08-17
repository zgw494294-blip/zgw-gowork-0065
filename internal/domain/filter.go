package domain

// 按产品层级过滤版本列表
type VersionFilter struct {
	LevelID string
}

func FilterMaskVersions(all []*MaskVersion, filter VersionFilter) []*MaskVersion {
	if filter.LevelID == "" {
		return all
	}
	result := make([]*MaskVersion, 0)
	for _, v := range all {
		if v.LevelID == filter.LevelID {
			result = append(result, v)
		}
	}
	return result
}
