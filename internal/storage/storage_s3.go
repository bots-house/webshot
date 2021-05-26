package storage

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"path"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/rs/xid"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
	"golang.org/x/xerrors"
)

const (
	defaultCacheControlMaxAge = 3600 * 24 * 360
	fileMetadataLatestKey     = "Latest"
	fileMetadataTTLKey        = "Ttl"
	hashFirstChars            = 15
)

type S3 struct {
	session    *session.Session
	client     *s3.S3
	bucket     string
	subdir     string
	uploader   *s3manager.Uploader
	downloader *s3manager.Downloader
}

func NewS3(s *session.Session, bucket string, subdir string) *S3 {
	return &S3{
		session:    s,
		client:     s3.New(s),
		bucket:     bucket,
		subdir:     subdir,
		uploader:   s3manager.NewUploader(s),
		downloader: s3manager.NewDownloader(s),
	}
}

func parseMetadataTTL(md map[string]*string) (time.Duration, error) {
	ttlStr, ok := md[fileMetadataTTLKey]
	if !ok {
		return 0, ErrFileCorrupted
	}

	ttlInt, err := strconv.Atoi(*ttlStr)
	if err != nil {
		return 0, ErrFileCorrupted
	}

	return time.Duration(ttlInt) * time.Second, nil
}

func (s *S3) Get(ctx context.Context, in Meta) (io.Reader, error) {
	linkPath := s.getLinkPath(in)

	buf := &aws.WriteAtBuffer{}

	link, err := s.client.HeadObjectWithContext(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(linkPath),
	})

	if isS3NotFoundErr(err) {
		return nil, ErrFileNotFound
	} else if err != nil {
		return nil, xerrors.Errorf("get link file: %w", err)
	}

	if link.Metadata == nil {
		return nil, ErrFileCorrupted
	}

	ttl, err := parseMetadataTTL(link.Metadata)
	if err != nil {
		return nil, xerrors.Errorf("parse ttl: %w", err)
	}

	lastModifed := *link.LastModified

	if time.Now().After(lastModifed.Add(ttl)) {
		return nil, ErrFileExpired
	}

	latestFilePath, ok := link.Metadata[fileMetadataLatestKey]
	if !ok {
		return nil, ErrFileCorrupted
	}

	_, err = s.downloader.DownloadWithContext(ctx, buf, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    latestFilePath,
	})

	if err != nil {
		return nil, xerrors.Errorf("get object: %w", err)
	}

	return bytes.NewBuffer(buf.Bytes()), nil
}

func (s *S3) getLinkPath(in Meta) string {
	h := sha256.New()
	h.Write([]byte(in.URL.String()))
	h.Write([]byte(in.Opts))
	hash := hex.EncodeToString(h.Sum(nil))

	p := fmt.Sprintf("/%s/%s.link", in.URL.Hostname(), hash[:hashFirstChars])

	return path.Join(s.subdir, p)
}

func (s *S3) getFilePath(in Meta) string {
	h := sha256.New()
	h.Write([]byte(in.URL.String()))
	h.Write([]byte(in.Opts))
	hash := hex.EncodeToString(h.Sum(nil))

	loc := fmt.Sprintf("/%s/%s.%s.%s",
		in.URL.Hostname(),
		hash[:hashFirstChars],
		xid.New().String(),
		in.Format.Ext(),
	)

	return path.Join(s.subdir, loc)
}

func isS3NotFoundErr(err error) bool {
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case "NotFound": // s3.ErrCodeNoSuchKey does not work, aws is missing this error code so we hardwire a string
				return true
			default:
				return false
			}
		}
		return false
	}
	return false
}

func (s *S3) Upload(ctx context.Context, in Upload) error {
	filePath := s.getFilePath(in.Meta)
	linkPath := s.getLinkPath(in.Meta)

	defer func(s time.Time) {
		log.Ctx(ctx).Debug().
			Dur("took", time.Since(s)).
			Dur("ttl", in.TTL).
			Str("url", in.Meta.URL.String()).
			Str("path", filePath).
			Msg("upload")
	}(time.Now())

	g, ctx := errgroup.WithContext(ctx)

	// upload link
	g.Go(func() error {
		_, err := s.uploader.UploadWithContext(ctx, &s3manager.UploadInput{
			Bucket:      aws.String(s.bucket),
			Key:         aws.String(linkPath),
			Body:        bytes.NewBufferString(linkPath),
			ContentType: aws.String("application/octet-stream"),
			Metadata: aws.StringMap(map[string]string{
				fileMetadataLatestKey: filePath,
				fileMetadataTTLKey:    strconv.Itoa(int(in.TTL.Seconds())),
			}),
		})

		if err != nil {
			return xerrors.Errorf("upload link file")
		}

		return nil
	})

	// upload file
	g.Go(func() error {
		_, err := s.uploader.UploadWithContext(ctx, &s3manager.UploadInput{
			Bucket:       aws.String(s.bucket),
			Key:          aws.String(filePath),
			Body:         in.Body,
			CacheControl: aws.String(fmt.Sprintf("max-age=%d", defaultCacheControlMaxAge)),
			ContentType:  aws.String(in.Meta.Format.ContentType()),
			ACL:          aws.String("public-read"),
			Metadata: aws.StringMap(map[string]string{
				fileMetadataTTLKey: strconv.Itoa(int(in.TTL.Seconds())),
			}),
		})

		if err != nil {
			return xerrors.Errorf("upload: %w", err)
		}

		return nil
	})

	return g.Wait()
}
