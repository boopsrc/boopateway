package entity

import "net/http"

const MaxImageBytes = 8 << 20 // 8 MiB

type Image struct {
	Data        []byte
	ContentType string // image/jpeg | image/png
}

func NewImage(data []byte) (Image, error) {
	if len(data) == 0 {
		return Image{}, ValidationError("campo image é obrigatório")
	}
	if len(data) > MaxImageBytes {
		return Image{}, ValidationError("imagem maior que 8 MiB")
	}
	ct := http.DetectContentType(data)
	if ct != "image/jpeg" && ct != "image/png" {
		return Image{}, ValidationError("imagem deve ser JPEG ou PNG")
	}
	return Image{Data: data, ContentType: ct}, nil
}

func (i Image) Extension() string {
	if i.ContentType == "image/png" {
		return ".png"
	}
	return ".jpg"
}
