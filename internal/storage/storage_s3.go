package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"path"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/rs/zerolog/log"
	"golang.org/x/xerrors"
)

const (
	defaultCacheControlMaxAge = 3600 * 24 * 360
)

type S3 struct {
	session    *session.Session
	s3         *s3.S3
	bucket     string
	subdir     string
	uploader   *s3manager.Uploader
	downloader *s3manager.Downloader
}

func NewS3(s *session.Session, bucket string, subdir string) *S3 {
	return &S3{
		session:    s,
		s3:         s3.New(s),
		bucket:     bucket,
		subdir:     subdir,
		uploader:   s3manager.NewUploader(s),
		downloader: s3manager.NewDownloader(s),
	}
}

func (s *S3) getKey(in Meta) string {
	return path.Join(s.subdir, in.Path())
}

func (s *S3) Get(ctx context.Context, in Meta) (io.Reader, error) {
	buf := &aws.WriteAtBuffer{}

	_, err := s.downloader.DownloadWithContext(ctx, buf, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(s.getKey(in)),
	})

	if err != nil {
		return nil, xerrors.Errorf("get object: %w", err)
	}

	return bytes.NewBuffer(buf.Bytes()), nil
}

func (s *S3) Has(ctx context.Context, in Meta) (v bool, err error) {
	p := aws.String(s.getKey(in))
	defer func(s time.Time) {
		log.Ctx(ctx).Debug().
			Dur("took", time.Since(s)).
			Str("url", in.URL.String()).
			Str("path", *p).
			Bool("exists", v).
			Msg("check if file exists")
	}(time.Now())

	_, err = s.s3.HeadObjectWithContext(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    p,
	})

	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case "NotFound": // s3.ErrCodeNoSuchKey does not work, aws is missing this error code so we hardwire a string
				return false, nil
			default:
				return false, err
			}
		}
		return false, err
	}
	return true, nil
}

func (s *S3) Upload(ctx context.Context, in Upload) error {
	p := s.getKey(in.Meta)

	defer func(s time.Time) {
		log.Ctx(ctx).Debug().
			Dur("took", time.Since(s)).
			Dur("ttl", in.TTL).
			Str("url", in.Meta.URL.String()).
			Str("path", p).
			Msg("upload")
	}(time.Now())

	_, err := s.uploader.UploadWithContext(ctx, &s3manager.UploadInput{
		Bucket:       aws.String(s.bucket),
		Key:          aws.String(p),
		Body:         in.Body,
		CacheControl: aws.String(fmt.Sprintf("max-age=%d", defaultCacheControlMaxAge)),
		ContentType:  aws.String(in.Meta.Format.ContentType()),
		ACL:          aws.String("public-read"),
	})

	if err != nil {
		return xerrors.Errorf("upload: %w", err)
	}

	return nil
}
