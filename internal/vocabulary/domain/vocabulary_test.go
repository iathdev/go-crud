package domain

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

func TestNewVocabulary(t *testing.T) {
	v, err := NewVocabulary("你好", "nǐ hǎo", "Xin chào", "Hello", 1, "daily_life")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if v.Hanzi != "你好" {
		t.Errorf("expected hanzi 你好, got %s", v.Hanzi)
	}
	if v.HSKLevel != 1 {
		t.Errorf("expected hsk level 1, got %d", v.HSKLevel)
	}
	if v.ID == uuid.Nil {
		t.Error("expected non-nil UUID")
	}
}

func TestNewVocabulary_EmptyHanzi(t *testing.T) {
	_, err := NewVocabulary("", "nǐ hǎo", "Xin chào", "", 1, "")
	if !errors.Is(err, ErrHanziRequired) {
		t.Errorf("expected ErrHanziRequired, got %v", err)
	}
}

func TestNewVocabulary_EmptyPinyin(t *testing.T) {
	_, err := NewVocabulary("你好", "", "Xin chào", "", 1, "")
	if !errors.Is(err, ErrPinyinRequired) {
		t.Errorf("expected ErrPinyinRequired, got %v", err)
	}
}

func TestNewVocabulary_NoMeaning(t *testing.T) {
	_, err := NewVocabulary("你好", "nǐ hǎo", "", "", 1, "")
	if !errors.Is(err, ErrMeaningRequired) {
		t.Errorf("expected ErrMeaningRequired, got %v", err)
	}
}

func TestNewVocabulary_InvalidHSKLevel(t *testing.T) {
	_, err := NewVocabulary("你好", "nǐ hǎo", "Xin chào", "", 0, "")
	if !errors.Is(err, ErrInvalidHSKLevel) {
		t.Errorf("expected ErrInvalidHSKLevel, got %v", err)
	}

	_, err = NewVocabulary("你好", "nǐ hǎo", "Xin chào", "", 10, "")
	if !errors.Is(err, ErrInvalidHSKLevel) {
		t.Errorf("expected ErrInvalidHSKLevel, got %v", err)
	}
}

func TestNewVocabulary_OnlyVietnameseMeaning(t *testing.T) {
	v, err := NewVocabulary("你好", "nǐ hǎo", "Xin chào", "", 1, "")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if v.MeaningVI != "Xin chào" {
		t.Errorf("expected meaning_vi Xin chào, got %s", v.MeaningVI)
	}
}

func TestVocabulary_Update(t *testing.T) {
	v, _ := NewVocabulary("你好", "nǐ hǎo", "Xin chào", "Hello", 1, "daily_life")
	err := v.Update("你们好", "nǐ men hǎo", "Xin chào các bạn", "Hello everyone", 2, "education")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if v.Hanzi != "你们好" {
		t.Errorf("expected 你们好, got %s", v.Hanzi)
	}
	if v.HSKLevel != 2 {
		t.Errorf("expected level 2, got %d", v.HSKLevel)
	}
}

func TestNewVocabularyFromParams(t *testing.T) {
	v, err := NewVocabularyFromParams(VocabularyParams{
		Hanzi:     "学习",
		Pinyin:    "xué xí",
		MeaningVI: "Học tập",
		MeaningEN: "Study",
		HSKLevel:  1,
		Examples: []Example{
			{SentenceCN: "我学习中文", SentenceVI: "Tôi học tiếng Trung"},
		},
		Radicals:      []string{"子", "冖"},
		StrokeCount:   8,
		FrequencyRank: 50,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(v.Examples) != 1 {
		t.Errorf("expected 1 example, got %d", len(v.Examples))
	}
	if len(v.Radicals) != 2 {
		t.Errorf("expected 2 radicals, got %d", len(v.Radicals))
	}
	if v.StrokeCount != 8 {
		t.Errorf("expected stroke count 8, got %d", v.StrokeCount)
	}
	if v.FrequencyRank != 50 {
		t.Errorf("expected frequency rank 50, got %d", v.FrequencyRank)
	}
}

func TestVocabulary_UpdateFromParams(t *testing.T) {
	v, _ := NewVocabularyFromParams(VocabularyParams{
		Hanzi:     "学习",
		Pinyin:    "xué xí",
		MeaningVI: "Học tập",
		HSKLevel:  1,
	})

	err := v.UpdateFromParams(VocabularyParams{
		Hanzi:           "学习",
		Pinyin:          "xué xí",
		MeaningVI:       "Học tập",
		MeaningEN:       "Study",
		HSKLevel:        2,
		RecognitionOnly: true,
		StrokeCount:     8,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if v.HSKLevel != 2 {
		t.Errorf("expected level 2, got %d", v.HSKLevel)
	}
	if !v.RecognitionOnly {
		t.Error("expected recognition_only to be true")
	}
	if v.StrokeCount != 8 {
		t.Errorf("expected stroke count 8, got %d", v.StrokeCount)
	}
}
