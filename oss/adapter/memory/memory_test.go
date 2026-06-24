package memory

import (
	"context"
	"io"
	"net/url"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/f2xme/gox/oss"
)

func TestStorageImplementsOSSStorage(t *testing.T) {
	var _ oss.Storage = (*Storage)(nil)
	var _ oss.BucketStorage = (*Storage)(nil)
}

func TestStoragePutGetStat(t *testing.T) {
	storage, err := New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	metadata := map[string]string{"author": "gox"}
	if err := storage.Put(context.Background(), "docs/readme.txt", strings.NewReader("hello"),
		oss.WithContentType("text/custom"),
		oss.WithMetadata(metadata),
	); err != nil {
		t.Fatalf("Put() error = %v", err)
	}
	metadata["author"] = "changed"

	body, err := storage.Get(context.Background(), "docs/readme.txt")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	defer body.Close()

	data, err := io.ReadAll(body)
	if err != nil {
		t.Fatalf("ReadAll() error = %v", err)
	}
	if string(data) != "hello" {
		t.Fatalf("Get() body = %q, want %q", data, "hello")
	}

	info, err := storage.Stat(context.Background(), "docs/readme.txt")
	if err != nil {
		t.Fatalf("Stat() error = %v", err)
	}
	if info.Key != "docs/readme.txt" || info.Size != 5 || info.ContentType != "text/custom" {
		t.Fatalf("Stat() = %+v", info)
	}
	if info.Metadata["author"] != "gox" {
		t.Fatalf("Metadata[author] = %q, want %q", info.Metadata["author"], "gox")
	}
	info.Metadata["author"] = "mutated"

	info, err = storage.Stat(context.Background(), "docs/readme.txt")
	if err != nil {
		t.Fatalf("Stat() after mutation error = %v", err)
	}
	if info.Metadata["author"] != "gox" {
		t.Fatalf("metadata was mutated through Stat result")
	}
	if info.ETag == "" || info.LastModified.IsZero() {
		t.Fatalf("Stat() missing generated fields: %+v", info)
	}
}

func TestStoragePutDetectsContentType(t *testing.T) {
	storage := MustNew()

	if err := storage.Put(context.Background(), "avatar.png", strings.NewReader("png")); err != nil {
		t.Fatalf("Put() error = %v", err)
	}

	info, err := storage.Stat(context.Background(), "avatar.png")
	if err != nil {
		t.Fatalf("Stat() error = %v", err)
	}
	if info.ContentType != "image/png" {
		t.Fatalf("ContentType = %q, want %q", info.ContentType, "image/png")
	}
}

func TestStorageGetRange(t *testing.T) {
	storage := MustNew()
	if err := storage.Put(context.Background(), "letters.txt", strings.NewReader("abcdef")); err != nil {
		t.Fatalf("Put() error = %v", err)
	}

	body, err := storage.Get(context.Background(), "letters.txt", oss.WithRange(1, 3))
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	defer body.Close()

	data, err := io.ReadAll(body)
	if err != nil {
		t.Fatalf("ReadAll() error = %v", err)
	}
	if string(data) != "bcd" {
		t.Fatalf("range body = %q, want %q", data, "bcd")
	}

	_, err = storage.Get(context.Background(), "letters.txt", oss.WithRange(8, 9))
	if !oss.IsCode(err, oss.ErrCodeInvalidArgument) {
		t.Fatalf("Get() invalid range error = %v, want invalid argument", err)
	}
}

func TestStorageGetReturnsIndependentReaders(t *testing.T) {
	storage := MustNew()
	if err := storage.Put(context.Background(), "data.bin", strings.NewReader("abc")); err != nil {
		t.Fatalf("Put() error = %v", err)
	}

	body, err := storage.Get(context.Background(), "data.bin")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	defer body.Close()

	data, err := io.ReadAll(body)
	if err != nil {
		t.Fatalf("ReadAll() error = %v", err)
	}
	data[0] = 'z'

	body, err = storage.Get(context.Background(), "data.bin")
	if err != nil {
		t.Fatalf("Get() after mutation error = %v", err)
	}
	defer body.Close()
	data, _ = io.ReadAll(body)
	if string(data) != "abc" {
		t.Fatalf("stored data = %q, want %q", data, "abc")
	}
}

func TestStorageDeleteExistsAndNotFound(t *testing.T) {
	storage := MustNew()
	ctx := context.Background()

	exists, err := storage.Exists(ctx, "missing.txt")
	if err != nil {
		t.Fatalf("Exists() error = %v", err)
	}
	if exists {
		t.Fatal("Exists() = true, want false")
	}

	_, err = storage.Get(ctx, "missing.txt")
	if !oss.IsNotFound(err) {
		t.Fatalf("Get() error = %v, want not found", err)
	}

	if err := storage.Put(ctx, "file.txt", strings.NewReader("content")); err != nil {
		t.Fatalf("Put() error = %v", err)
	}
	exists, err = storage.Exists(ctx, "file.txt")
	if err != nil {
		t.Fatalf("Exists() after put error = %v", err)
	}
	if !exists {
		t.Fatal("Exists() = false, want true")
	}

	if err := storage.Delete(ctx, "file.txt"); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	exists, err = storage.Exists(ctx, "file.txt")
	if err != nil {
		t.Fatalf("Exists() after delete error = %v", err)
	}
	if exists {
		t.Fatal("Exists() after delete = true, want false")
	}
}

func TestStorageList(t *testing.T) {
	storage := MustNew()
	ctx := context.Background()
	for _, key := range []string{
		"docs/a.txt",
		"docs/b.txt",
		"docs/nested/c.txt",
		"images/a.png",
		"root.txt",
	} {
		if err := storage.Put(ctx, key, strings.NewReader(key)); err != nil {
			t.Fatalf("Put(%q) error = %v", key, err)
		}
	}

	result, err := storage.List(ctx, oss.WithPrefix("docs/"))
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if got := objectKeys(result.Objects); !reflect.DeepEqual(got, []string{"docs/a.txt", "docs/b.txt", "docs/nested/c.txt"}) {
		t.Fatalf("objects = %#v", got)
	}

	result, err = storage.List(ctx, oss.WithPrefix("docs/"), oss.WithDelimiter("/"))
	if err != nil {
		t.Fatalf("List() with delimiter error = %v", err)
	}
	if got := objectKeys(result.Objects); !reflect.DeepEqual(got, []string{"docs/a.txt", "docs/b.txt"}) {
		t.Fatalf("objects with delimiter = %#v", got)
	}
	if !reflect.DeepEqual(result.Prefixes, []string{"docs/nested/"}) {
		t.Fatalf("prefixes = %#v", result.Prefixes)
	}
}

func TestStorageListPagination(t *testing.T) {
	storage := MustNew()
	ctx := context.Background()
	for _, key := range []string{"a.txt", "b.txt", "c.txt"} {
		if err := storage.Put(ctx, key, strings.NewReader(key)); err != nil {
			t.Fatalf("Put(%q) error = %v", key, err)
		}
	}

	first, err := storage.List(ctx, oss.WithLimit(2))
	if err != nil {
		t.Fatalf("List() first error = %v", err)
	}
	if !first.Truncated || first.NextToken == "" {
		t.Fatalf("first page = %+v, want truncated with token", first)
	}
	if got := objectKeys(first.Objects); !reflect.DeepEqual(got, []string{"a.txt", "b.txt"}) {
		t.Fatalf("first objects = %#v", got)
	}

	second, err := storage.List(ctx, oss.WithToken(first.NextToken))
	if err != nil {
		t.Fatalf("List() second error = %v", err)
	}
	if second.Truncated {
		t.Fatalf("second page truncated = true, want false")
	}
	if got := objectKeys(second.Objects); !reflect.DeepEqual(got, []string{"c.txt"}) {
		t.Fatalf("second objects = %#v", got)
	}
}

func TestStorageSignURL(t *testing.T) {
	storage := MustNew(WithSignURLBase("https://example.test/objects"))

	raw, err := storage.SignURL(context.Background(), "docs/read me.txt",
		oss.WithMethod(oss.MethodPut),
		oss.WithExpires(time.Hour),
		oss.WithSignContentType("text/plain"),
	)
	if err != nil {
		t.Fatalf("SignURL() error = %v", err)
	}

	parsed, err := url.Parse(raw)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if parsed.Scheme != "https" || parsed.Host != "example.test" || parsed.Path != "/objects/docs/read me.txt" {
		t.Fatalf("url = %q", raw)
	}
	query := parsed.Query()
	if query.Get("method") != oss.MethodPut || query.Get("expires") != "3600" || query.Get("content-type") != "text/plain" {
		t.Fatalf("query = %#v", query)
	}

	_, err = storage.SignURL(context.Background(), "file.txt", oss.WithMethod("POST"))
	if !oss.IsCode(err, oss.ErrCodeInvalidArgument) {
		t.Fatalf("SignURL() invalid method error = %v, want invalid argument", err)
	}
}

func TestNewInvalidOption(t *testing.T) {
	_, err := New(WithBucketName(""))
	if !oss.IsCode(err, oss.ErrCodeInvalidArgument) {
		t.Fatalf("New() bucket error = %v, want invalid argument", err)
	}

	_, err = New(WithSignURLBase(""))
	if !oss.IsCode(err, oss.ErrCodeInvalidArgument) {
		t.Fatalf("New() sign url error = %v, want invalid argument", err)
	}
}

func TestStorageBucketLifecycle(t *testing.T) {
	storage := MustNew(WithBucketName("default"), WithSignURLBase("https://example.test"))
	ctx := context.Background()

	buckets, err := storage.ListBuckets(ctx)
	if err != nil {
		t.Fatalf("ListBuckets() error = %v", err)
	}
	if got := bucketNames(buckets); !reflect.DeepEqual(got, []string{"default"}) {
		t.Fatalf("buckets = %#v", got)
	}

	if err := storage.CreateBucket(ctx, "archive", oss.WithBucketRegion("cn-test")); err != nil {
		t.Fatalf("CreateBucket() error = %v", err)
	}
	if err := storage.CreateBucket(ctx, "archive"); !oss.IsCode(err, oss.ErrCodeBucketExists) {
		t.Fatalf("CreateBucket() duplicate error = %v, want bucket exists", err)
	}

	buckets, err = storage.ListBuckets(ctx)
	if err != nil {
		t.Fatalf("ListBuckets() after create error = %v", err)
	}
	if got := bucketNames(buckets); !reflect.DeepEqual(got, []string{"archive", "default"}) {
		t.Fatalf("buckets after create = %#v", got)
	}
	if buckets[0].Region != "cn-test" {
		t.Fatalf("archive region = %q, want %q", buckets[0].Region, "cn-test")
	}

	if err := storage.DeleteBucket(ctx, "archive"); err != nil {
		t.Fatalf("DeleteBucket() error = %v", err)
	}
	if err := storage.DeleteBucket(ctx, "archive"); !oss.IsNotFound(err) {
		t.Fatalf("DeleteBucket() missing error = %v, want not found", err)
	}
}

func TestStorageDeleteDefaultBucket(t *testing.T) {
	storage := MustNew()
	ctx := context.Background()

	if err := storage.Put(ctx, "file.txt", strings.NewReader("content")); err != nil {
		t.Fatalf("Put() error = %v", err)
	}
	if err := storage.DeleteBucket(ctx, "memory"); !oss.IsCode(err, oss.ErrCodeBucketNotEmpty) {
		t.Fatalf("DeleteBucket() non-empty error = %v, want bucket not empty", err)
	}

	if err := storage.Delete(ctx, "file.txt"); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if err := storage.DeleteBucket(ctx, "memory"); err != nil {
		t.Fatalf("DeleteBucket() empty error = %v", err)
	}
	if err := storage.Put(ctx, "file.txt", strings.NewReader("content")); !oss.IsNotFound(err) {
		t.Fatalf("Put() after bucket delete error = %v, want not found", err)
	}
	if err := storage.CreateBucket(ctx, "memory"); err != nil {
		t.Fatalf("CreateBucket() default error = %v", err)
	}
	if err := storage.Put(ctx, "file.txt", strings.NewReader("content")); err != nil {
		t.Fatalf("Put() after recreate error = %v", err)
	}
}

func TestStorageContextCanceled(t *testing.T) {
	storage := MustNew()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := storage.Put(ctx, "file.txt", strings.NewReader("data")); err != context.Canceled {
		t.Fatalf("Put() error = %v, want context.Canceled", err)
	}
	if _, err := storage.Get(ctx, "file.txt"); err != context.Canceled {
		t.Fatalf("Get() error = %v, want context.Canceled", err)
	}
	if _, err := storage.List(ctx); err != context.Canceled {
		t.Fatalf("List() error = %v, want context.Canceled", err)
	}
	if err := storage.CreateBucket(ctx, "bucket"); err != context.Canceled {
		t.Fatalf("CreateBucket() error = %v, want context.Canceled", err)
	}
	if err := storage.DeleteBucket(ctx, "memory"); err != context.Canceled {
		t.Fatalf("DeleteBucket() error = %v, want context.Canceled", err)
	}
	if _, err := storage.ListBuckets(ctx); err != context.Canceled {
		t.Fatalf("ListBuckets() error = %v, want context.Canceled", err)
	}
}

func objectKeys(objects []*oss.Object) []string {
	keys := make([]string, len(objects))
	for i, object := range objects {
		keys[i] = object.Key
	}
	return keys
}

func bucketNames(buckets []*oss.Bucket) []string {
	names := make([]string, len(buckets))
	for i, bucket := range buckets {
		names[i] = bucket.Name
	}
	return names
}
