package text

import (
	"context"

	"typer/internal/model"
	"typer/internal/train"
)

type TrainProvider struct {
	filter *train.WordFilter
	lesson train.Lesson
}

func NewTrainProvider(filter *train.WordFilter, lesson train.Lesson) *TrainProvider {
	return &TrainProvider{
		filter: filter,
		lesson: lesson,
	}
}

func (p *TrainProvider) Name() string {
	return "train"
}

func (p *TrainProvider) Next(_ context.Context, _ Constraints) (model.Prompt, error) {
	content, err := train.BuildLessonContent(p.lesson, p.filter)
	if err != nil {
		return model.Prompt{}, err
	}
	return model.Prompt{
		ID:      p.lesson.ID,
		Content: content,
		Source:  "train",
		Mode:    model.ModeTrain,
	}, nil
}

// Lesson returns the configured lesson.
func (p *TrainProvider) Lesson() train.Lesson {
	return p.lesson
}
