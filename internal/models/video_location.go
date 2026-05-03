package models

import "time"

// VideoLocation records where a video content entity appears on disk.
type VideoLocation struct {
	ID           int64     `json:"id" gorm:"primaryKey"`
	VideoID      int64     `json:"video_id" gorm:"index;not null"`
	DirectoryID  int64     `json:"directory_id" gorm:"index;not null;uniqueIndex:idx_video_location_directory_path"`
	RelativePath string    `json:"relative_path" gorm:"not null;uniqueIndex:idx_video_location_directory_path"`
	ModifiedAt   time.Time `json:"modified_at"`
	JavID        *int64    `json:"jav_id" gorm:"index"`
	Jav          *Jav      `json:"jav,omitempty" gorm:"-"`
	IsDelete     bool      `json:"is_delete" gorm:"index"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	Video        Video     `json:"-" gorm:"foreignKey:VideoID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	DirectoryRef Directory `json:"directory,omitempty" gorm:"foreignKey:DirectoryID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
}
