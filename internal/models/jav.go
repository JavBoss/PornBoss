package models

import "time"

// Jav stores metadata fetched for a given code (may map to multiple videos).
type Jav struct {
	ID          int64     `json:"id" gorm:"primaryKey"`
	Code        string    `json:"code" gorm:"uniqueIndex"`
	Title       string    `json:"title"`
	ReleaseUnix int64     `json:"release_unix"`
	DurationMin int       `json:"duration_min"`
	Provider    int       `json:"provider" gorm:"not null;default:0"`
	FetchedAt   time.Time `json:"fetched_at"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Tags        []JavTag  `json:"tags,omitempty" gorm:"many2many:jav_tag_map"`
	Idols       []JavIdol `json:"idols,omitempty" gorm:"many2many:jav_idol_map"`
	Videos      []Video   `json:"videos,omitempty" gorm:"foreignKey:JavID"`
}

type JavTag struct {
	ID        int64     `json:"id" gorm:"primaryKey"`
	Name      string    `json:"name" gorm:"uniqueIndex:idx_jav_tag_name_source"`
	Provider  int       `json:"provider" gorm:"not null;default:0;index:idx_jav_tag_provider;uniqueIndex:idx_jav_tag_name_source"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type JavIdol struct {
	ID           int64      `json:"id" gorm:"primaryKey"`
	Name         string     `json:"name" gorm:"uniqueIndex"`
	RomanName    string     `json:"roman_name"`
	JapaneseName string     `json:"japanese_name"`
	ChineseName  string     `json:"chinese_name"`
	HeightCM     *int       `json:"height_cm"`
	BirthDate    *time.Time `json:"birth_date"`
	Bust         *int       `json:"bust"`
	Waist        *int       `json:"waist"`
	Hips         *int       `json:"hips"`
	Cup          *int       `json:"cup"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// Many-to-many join tables.
type JavTagMap struct {
	JavID     int64     `gorm:"primaryKey"`
	Jav       Jav       `gorm:"foreignKey:JavID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	JavTagID  int64     `gorm:"primaryKey"`
	JavTag    JavTag    `gorm:"foreignKey:JavTagID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

type JavIdolMap struct {
	JavID     int64     `gorm:"primaryKey"`
	Jav       Jav       `gorm:"foreignKey:JavID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	JavIdolID int64     `gorm:"primaryKey"`
	JavIdol   JavIdol   `gorm:"foreignKey:JavIdolID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}
