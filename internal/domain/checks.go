package domain

// 复合约束：同一产品层级下只能有一个启用版本
type CompositeConstraint interface {
	Validate() error
}

// 验证批次必须绑定送审版本（即绑定已提交变更申请的掩模版版本）
func ValidateBatchMaskVersion(batch *VerificationBatch, cr *ChangeRequest, mv *MaskVersion) error {
	if cr.Status != CRSubmitted && cr.Status != CRVerified {
		return &ConstraintError{Message: "变更申请未送审"}
	}
	if batch.MaskVersionID != cr.MaskVersionID {
		return &ConstraintError{Message: "验证批次必须绑定送审版本的掩模版"}
	}
	if mv.Status != StatusSubmitted && mv.Status != StatusInTrial {
		return &ConstraintError{Message: "掩模版状态不是送审或试投"}
	}
	return nil
}

// 关键验证项未通过时禁止启用
func ValidateEnable(mv *MaskVersion, batch *VerificationBatch) error {
	if batch == nil || !batch.CriticalPassed {
		return &ConstraintError{Message: "关键验证项未通过"}
	}
	if mv.Status != StatusVerified {
		return &ConstraintError{Message: "掩模版未验证通过"}
	}
	return nil
}

type ConstraintError struct {
	Message string
}

func (e *ConstraintError) Error() string {
	return e.Message
}
