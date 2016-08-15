package template

import (
	"crypto/sha512"
	"fmt"
)

type InterpolatedTemplate struct {
	content []byte
	sha     string
}

func (t *InterpolatedTemplate) Content() []byte {
	return t.content
}

func (t *InterpolatedTemplate) SHA() string {
	return t.sha
}

func NewInterpolatedTemplate(content []byte) InterpolatedTemplate {
	sha_512 := sha512.New()
	_, err := sha_512.Write(content)
	if err != nil {
		panic("Error calculating sha_512 of interpolated template")
	}
	shaSumString := fmt.Sprintf("%x", sha_512.Sum(nil))
	return InterpolatedTemplate{
		content: content,
		sha:     shaSumString,
	}
}
