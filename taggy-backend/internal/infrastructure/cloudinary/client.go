package cloudinary

import (
	"context"
	"errors"
	"io"
	"strings"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/rs/zerolog"
)

var ErrNotConfigured = errors.New("image storage is not configured")

type Client struct {
	cld *cloudinary.Cloudinary
	log zerolog.Logger
}

func New(cloudinaryURL string, log zerolog.Logger) *Client {
	url := strings.TrimSpace(cloudinaryURL)
	if url == "" {
		return &Client{log: log}
	}
	cld, err := cloudinary.NewFromURL(url)
	if err != nil {
		log.Error().Err(err).Msg("cloudinary URL is invalid")
		return &Client{log: log}
	}
	return &Client{cld: cld, log: log}
}

func (c *Client) Available() bool {
	return c != nil && c.cld != nil
}

func (c *Client) UploadAvatar(ctx context.Context, publicID string, file io.Reader, filename string) (string, error) {
	if !c.Available() {
		return "", ErrNotConfigured
	}
	overwrite := true
	invalidate := true
	resp, err := c.cld.Upload.Upload(ctx, file, uploader.UploadParams{
		PublicID:         "taggy/avatars/" + publicID,
		Overwrite:        &overwrite,
		Invalidate:       &invalidate,
		ResourceType:     "image",
		FilenameOverride: filename,
		UniqueFilename:   boolPtr(false),
		UseFilename:      boolPtr(false),
	})
	if err != nil {
		c.log.Error().Err(err).Str("public_id", publicID).Msg("cloudinary avatar upload failed")
		return "", err
	}
	if strings.TrimSpace(resp.SecureURL) == "" {
		return "", errors.New("cloudinary returned an empty image URL")
	}
	c.log.Info().Str("public_id", publicID).Msg("cloudinary avatar uploaded")
	return resp.SecureURL, nil
}

func boolPtr(v bool) *bool {
	return &v
}
