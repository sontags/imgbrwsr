package thumbcache

import (
	"image"
)

type Thumb struct {
	Name  string
	Image image.Image
}

type ThumbCache struct {
	Ptr int
	Buf []Thumb
}

func (c *ThumbCache) HasThumb(name string) bool {
	for i := range c.Buf {
		if c.Buf[i].Name == name {
			return true
		}
	}
	return false
}

func (c *ThumbCache) GetThumb(name string) Thumb {
	var thumb Thumb
	for i := range c.Buf {
		if c.Buf[i].Name == name {
			thumb = c.Buf[i]
		}
	}
	return thumb
}

func (c *ThumbCache) AddThumb(thumbnail Thumb) {
	c.Buf[c.Ptr] = thumbnail
	c.Ptr = (c.Ptr + 1) % len(c.Buf)
}
