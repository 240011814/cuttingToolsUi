package service

import (
	"backend/model"
)

type VocabularyService struct{}

func NewVocabularyService() *VocabularyService {
	return &VocabularyService{}
}

// AddWord 添加生词
func (s *VocabularyService) AddWord(userID uint, req model.CreateVocabularyRequest) (*model.Vocabulary, error) {
	word := model.Vocabulary{
		UserID:        userID,
		Word:          req.Word,
		Phonetic:      req.Phonetic,
		Definition:     req.Definition,
		Example:        req.Example,
		SourceContext:  req.SourceContext,
		ConfusingWords: req.ConfusingWords,
	}

	if err := DB.Create(&word).Error; err != nil {
		return nil, err
	}
	return &word, nil
}

// GetUserVocabulary 获取用户生词列表
func (s *VocabularyService) GetUserVocabulary(userID uint, keyword string, isMastered *bool) ([]model.Vocabulary, error) {
	var list []model.Vocabulary
	query := DB.Where("user_id = ?", userID)
	if keyword != "" {
		query = query.Where("word LIKE ?", "%"+keyword+"%")
	}
	if isMastered != nil {
		query = query.Where("is_mastered = ?", *isMastered)
	}
	if err := query.Order("created_at DESC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// DeleteWord 删除生词
func (s *VocabularyService) DeleteWord(userID, id uint) error {
	return DB.Where("id = ? AND user_id = ?", id, userID).Delete(&model.Vocabulary{}).Error
}

// UpdateWord 更新生词
func (s *VocabularyService) UpdateWord(userID, id uint, req model.UpdateVocabularyRequest) error {
	updates := map[string]interface{}{}
	if req.Phonetic != "" {
		updates["phonetic"] = req.Phonetic
	}
	if req.Definition != "" {
		updates["definition"] = req.Definition
	}
	if req.Example != "" {
		updates["example"] = req.Example
	}
	if req.ConfusingWords != "" {
		updates["confusing_words"] = req.ConfusingWords
	}
	if req.IsMastered != nil {
		updates["is_mastered"] = *req.IsMastered
	}

	return DB.Model(&model.Vocabulary{}).
		Where("id = ? AND user_id = ?", id, userID).
		Updates(updates).Error
}
