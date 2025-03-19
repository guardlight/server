package theme

import (
	"errors"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type ThemeRepository struct {
	db *gorm.DB
}

func NewAnalysisManagerRepository(db *gorm.DB) *ThemeRepository {
	if err := db.AutoMigrate(
		&Theme{},
	); err != nil {
		zap.S().DPanicw("Problem automigrating the tables", "error", err)
	}

	return &ThemeRepository{
		db: db,
	}
}

func (tr *ThemeRepository) getAllThemesByUserId(id uuid.UUID) ([]Theme, error) {
	var ts []Theme
	if err := tr.db.Where("user_id = ?", id).Find(&ts).Error; err != nil {
		zap.S().Errorw("Could not get themes for user", "error", err)
		return nil, err
	}

	return ts, nil
}

func (tr *ThemeRepository) updateTheme(t Theme, uid uuid.UUID) error {
	res := tr.db.Model(Theme{
		Id:     t.Id,
		UserId: uid,
	}).Updates(Theme{
		Title:     t.Title,
		Analyzers: t.Analyzers,
	})

	if res.Error != nil {
		zap.S().Errorw("Could not update theme", "error", res.Error)
		return res.Error
	}

	if res.RowsAffected == 0 {
		zap.S().Errorw("No records updated", "theme_id", t.Id)
		return errors.New("no records affected after update")
	}

	return nil
}
