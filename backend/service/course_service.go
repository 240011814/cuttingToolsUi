package service

import (
	"backend/model"
	"errors"
	"strings"

	"gorm.io/gorm"
)

type CourseService struct{}

func NewCourseService() *CourseService {
	return &CourseService{}
}

// ListCourses 获取课程包列表（支持分页、搜索、筛选）
func (s *CourseService) ListCourses(userID uint, showAll bool, keyword string, isPublic *bool, tag string, page, pageSize int) ([]model.Course, int64, error) {
	var courses []model.Course
	var total int64

	query := DB.Model(&model.Course{})
	if !showAll {
		query = query.Where("user_id = ? OR is_public = ?", userID, true)
	}
	if keyword != "" {
		query = query.Where("title LIKE ? OR description LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	if isPublic != nil {
		query = query.Where("is_public = ?", *isPublic)
	}
	if tag != "" {
		query = query.Where("FIND_IN_SET(?, tags)", tag)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if page > 0 && pageSize > 0 {
		query = query.Offset((page - 1) * pageSize).Limit(pageSize)
	}

	err := query.Order("created_at DESC").Find(&courses).Error
	return courses, total, err
}

// GetCourseByID 获取课程包详情
func (s *CourseService) GetCourseByID(userID uint, courseID uint) (*model.Course, error) {
	var course model.Course
	err := DB.First(&course, courseID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("课程包不存在")
		}
		return nil, err
	}
	// 检查权限：只有创建者或公开课程可以访问
	if course.UserID != userID && !course.IsPublic {
		return nil, errors.New("无权访问此课程包")
	}
	return &course, nil
}

// CreateCourse 创建课程包
func (s *CourseService) CreateCourse(userID uint, req model.CreateCourseRequest) (*model.Course, error) {
	tagsStr := strings.Join(req.Tags, ",")
	course := model.Course{
		UserID:      userID,
		Title:       req.Title,
		Description: req.Description,
		Tags:        tagsStr,
		IsPublic:    req.IsPublic,
	}
	err := DB.Create(&course).Error
	return &course, err
}

// UpdateCourse 更新课程包
func (s *CourseService) UpdateCourse(userID uint, courseID uint, req model.UpdateCourseRequest) (*model.Course, error) {
	var course model.Course
	err := DB.First(&course, courseID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("课程包不存在")
		}
		return nil, err
	}
	if course.UserID != userID {
		return nil, errors.New("只能编辑自己创建的课程包")
	}
	if req.Title != "" {
		course.Title = req.Title
	}
	if req.Description != "" {
		course.Description = req.Description
	}
	if req.Tags != nil {
		course.Tags = strings.Join(req.Tags, ",")
	}
	if req.IsPublic != nil {
		course.IsPublic = *req.IsPublic
	}
	err = DB.Save(&course).Error
	return &course, err
}

// DeleteCourse 删除课程包
func (s *CourseService) DeleteCourse(userID uint, courseID uint) error {
	var course model.Course
	err := DB.First(&course, courseID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("课程包不存在")
		}
		return err
	}
	if course.UserID != userID {
		return errors.New("只能删除自己创建的课程包")
	}
	// 删除课程包及其所有条目
	return DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("course_id = ?", courseID).Delete(&model.CourseItem{}).Error; err != nil {
			return err
		}
		return tx.Delete(&course).Error
	})
}

// GetCourseItems 获取课程条目列表
func (s *CourseService) GetCourseItems(courseID uint) ([]model.CourseItem, error) {
	var items []model.CourseItem
	err := DB.Where("course_id = ?", courseID).Order("sort_order ASC, id ASC").Find(&items).Error
	return items, err
}

// CreateCourseItem 创建课程条目
func (s *CourseService) CreateCourseItem(userID uint, courseID uint, req model.CreateCourseItemRequest) (*model.CourseItem, error) {
	// 检查课程包是否存在且属于当前用户
	var course model.Course
	if err := DB.First(&course, courseID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("课程包不存在")
		}
		return nil, err
	}
	if course.UserID != userID {
		return nil, errors.New("只能向自己创建的课程包添加条目")
	}
	item := model.CourseItem{
		CourseID:           courseID,
		EnglishSentence:    req.EnglishSentence,
		ChineseTranslation: req.ChineseTranslation,
		SortOrder:          req.SortOrder,
	}
	err := DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&item).Error; err != nil {
			return err
		}
		// 更新课程包的条目数量
		return tx.Model(&course).UpdateColumn("item_count", gorm.Expr("item_count + 1")).Error
	})
	return &item, err
}

// UpdateCourseItem 更新课程条目
func (s *CourseService) UpdateCourseItem(userID uint, itemID uint, req model.UpdateCourseItemRequest) (*model.CourseItem, error) {
	var item model.CourseItem
	err := DB.First(&item, itemID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("条目不存在")
		}
		return nil, err
	}
	// 检查权限
	var course model.Course
	if err := DB.First(&course, item.CourseID).Error; err != nil {
		return nil, err
	}
	if course.UserID != userID {
		return nil, errors.New("只能编辑自己创建的课程条目")
	}
	if req.EnglishSentence != "" {
		item.EnglishSentence = req.EnglishSentence
	}
	if req.ChineseTranslation != "" {
		item.ChineseTranslation = req.ChineseTranslation
	}
	if req.SortOrder != nil {
		item.SortOrder = *req.SortOrder
	}
	err = DB.Save(&item).Error
	return &item, err
}

// DeleteCourseItem 删除课程条目
func (s *CourseService) DeleteCourseItem(userID uint, itemID uint) error {
	var item model.CourseItem
	err := DB.First(&item, itemID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("条目不存在")
		}
		return err
	}
	// 检查权限
	var course model.Course
	if err := DB.First(&course, item.CourseID).Error; err != nil {
		return err
	}
	if course.UserID != userID {
		return errors.New("只能删除自己创建的课程条目")
	}
	return DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(&item).Error; err != nil {
			return err
		}
		// 更新课程包的条目数量
		return tx.Model(&course).UpdateColumn("item_count", gorm.Expr("item_count - 1")).Error
	})
}

// BatchCreateCourseItems 批量创建课程条目
func (s *CourseService) BatchCreateCourseItems(userID uint, courseID uint, items []model.CreateCourseItemRequest) ([]model.CourseItem, error) {
	// 检查课程包是否存在且属于当前用户
	var course model.Course
	if err := DB.First(&course, courseID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("课程包不存在")
		}
		return nil, err
	}
	if course.UserID != userID {
		return nil, errors.New("只能向自己创建的课程包添加条目")
	}
	courseItems := make([]model.CourseItem, len(items))
	for i, req := range items {
		courseItems[i] = model.CourseItem{
			CourseID:           courseID,
			EnglishSentence:    req.EnglishSentence,
			ChineseTranslation: req.ChineseTranslation,
			SortOrder:          req.SortOrder,
		}
	}
	err := DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&courseItems).Error; err != nil {
			return err
		}
		// 更新课程包的条目数量
		return tx.Model(&course).UpdateColumn("item_count", gorm.Expr("item_count + ?", len(items))).Error
	})
	return courseItems, err
}

// BatchDeleteCourseItems 批量删除课程条目
func (s *CourseService) BatchDeleteCourseItems(userID uint, courseID uint, ids []uint) error {
	// 检查课程包是否存在且属于当前用户
	var course model.Course
	if err := DB.First(&course, courseID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("课程包不存在")
		}
		return err
	}
	if course.UserID != userID {
		return errors.New("只能删除自己创建的课程条目")
	}
	// 检查条目是否属于该课程包
	var count int64
	if err := DB.Model(&model.CourseItem{}).Where("id IN ? AND course_id = ?", ids, courseID).Count(&count).Error; err != nil {
		return err
	}
	if int(count) != len(ids) {
		return errors.New("部分条目不存在或不属于该课程包")
	}
	return DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("id IN ?", ids).Delete(&model.CourseItem{}).Error; err != nil {
			return err
		}
		// 更新课程包的条目数量
		return tx.Model(&course).UpdateColumn("item_count", gorm.Expr("item_count - ?", len(ids))).Error
	})
}
