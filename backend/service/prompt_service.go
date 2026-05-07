package service

import (
	"backend/model"
	"errors"

	"gorm.io/gorm"
)

type PromptService struct {
	db *gorm.DB
}

func NewPromptService(db *gorm.DB) *PromptService {
	return &PromptService{db: db}
}

// GetEffectivePrompt returns the currently active custom prompt for a module.
// If the user has not customized the prompt, the frontend should fall back to
// the module's built-in default prompt.
func (s *PromptService) GetEffectivePrompt(userID uint, moduleKey string) (string, error) {
	var userPrompt model.UserPrompt
	err := s.db.Where("user_id = ? AND module_key = ? AND is_active = ?", userID, moduleKey, true).First(&userPrompt).Error
	if err == nil {
		return userPrompt.CustomPrompt, nil
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", nil
	}

	return "", err
}

func (s *PromptService) ListVersions(userID uint, moduleKey string) ([]model.UserPrompt, error) {
	var list []model.UserPrompt
	err := s.db.Where("user_id = ? AND module_key = ?", userID, moduleKey).Order("version DESC").Find(&list).Error
	return list, err
}

func (s *PromptService) SaveUserPrompt(userID uint, moduleKey, content, remark string) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.UserPrompt{}).
			Where("user_id = ? AND module_key = ?", userID, moduleKey).
			Update("is_active", false).Error; err != nil {
			return err
		}

		var maxVersion int
		tx.Model(&model.UserPrompt{}).
			Where("user_id = ? AND module_key = ?", userID, moduleKey).
			Select("COALESCE(MAX(version), 0)").Scan(&maxVersion)

		newPrompt := model.UserPrompt{
			UserID:       userID,
			ModuleKey:    moduleKey,
			CustomPrompt: content,
			Version:      maxVersion + 1,
			IsActive:     true,
			Remark:       remark,
		}

		return tx.Create(&newPrompt).Error
	})
}

func (s *PromptService) SwitchVersion(userID uint, moduleKey string, versionID uint) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.UserPrompt{}).
			Where("user_id = ? AND module_key = ?", userID, moduleKey).
			Update("is_active", false).Error; err != nil {
			return err
		}

		return tx.Model(&model.UserPrompt{}).
			Where("id = ? AND user_id = ?", versionID, userID).
			Update("is_active", true).Error
	})
}

func (s *PromptService) ResetUserPrompt(userID uint, moduleKey string) error {
	return s.db.Where("user_id = ? AND module_key = ?", userID, moduleKey).Delete(&model.UserPrompt{}).Error
}

func (s *PromptService) DeleteVersion(userID uint, moduleKey string, versionID uint) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var prompt model.UserPrompt
		if err := tx.Where("id = ? AND user_id = ?", versionID, userID).First(&prompt).Error; err != nil {
			return err
		}

		if err := tx.Delete(&model.UserPrompt{}, versionID).Error; err != nil {
			return err
		}

		if prompt.IsActive {
			var latest model.UserPrompt
			err := tx.Where("user_id = ? AND module_key = ?", userID, moduleKey).
				Order("version DESC").First(&latest).Error
			if err == nil {
				return tx.Model(&latest).Update("is_active", true).Error
			}
		}

		return nil
	})
}
