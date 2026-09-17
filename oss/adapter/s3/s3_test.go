package s3

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/f2xme/gox/oss"
)

func testStorage(t *testing.T, endpoint string, opts ...Option) *Storage {
	t.Helper()
	base := []Option{WithEndpoint(endpoint), WithRegion("auto"), WithBucket("test-bucket"), WithCredentials("test-id", "test-secret"), WithUsePathStyle(true)}
	s, err := New(append(base, opts...)...)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestValidation(t *testing.T) {
	for _, opts := range [][]Option{
		nil, {nil}, {WithBucket("bucket")},
		{WithCredentials("id", "secret")},
		{WithBucket("bucket"), WithCredentials("id", "")},
	} {
		if _, err := New(opts...); !oss.IsCode(err, oss.ErrCodeInvalidArgument) {
			t.Fatalf("New(%v): %v", opts, err)
		}
	}
	if _, err := NewWithOptions(nil); !oss.IsCode(err, oss.ErrCodeInvalidArgument) {
		t.Fatal(err)
	}
	for _, endpoint := range []string{"missing-scheme", "ftp://example.com", "https://", "https://user:pass@example.com", "https://example.com?x=y", "https://example.com/#fragment"} {
		_, err := New(WithEndpoint(endpoint), WithCredentials("id", "secret"), WithBucket("bucket"))
		if !oss.IsCode(err, oss.ErrCodeInvalidArgument) {
			t.Fatalf("endpoint %q: %v", endpoint, err)
		}
	}
	s := testStorage(t, "https://example.com")
	ctx := context.Background()
	for name, call := range map[string]func() error{
		"nil reader":     func() error { return s.Put(ctx, "key", nil) },
		"empty key":      func() error { return s.Delete(ctx, "") },
		"range":          func() error { _, err := s.Get(ctx, "key", oss.WithRange(3, 1)); return err },
		"limit":          func() error { _, err := s.List(ctx, oss.WithLimit(1001)); return err },
		"negative limit": func() error { _, err := s.List(ctx, oss.WithLimit(-1)); return err },
		"bucket":         func() error { return s.CreateBucket(ctx, " ") },
		"method":         func() error { _, err := s.SignURL(ctx, "key", oss.WithMethod("POST")); return err },
		"expiry":         func() error { _, err := s.SignURL(ctx, "key", oss.WithExpires(8*24*time.Hour)); return err },
		"short expiry":   func() error { _, err := s.SignURL(ctx, "key", oss.WithExpires(time.Millisecond)); return err },
	} {
		t.Run(name, func(t *testing.T) {
			if err := call(); !oss.IsCode(err, oss.ErrCodeInvalidArgument) {
				t.Fatal(err)
			}
		})
	}
}

func TestObjectOperations(t *testing.T) {
	ctx := context.Background()
	const key = "目录/a +%.txt"
	for _, tt := range []struct {
		name, method string
		handle       func(*testing.T, http.ResponseWriter, *http.Request)
		call         func(*testing.T, *Storage)
	}{
		{"put", "PUT", func(t *testing.T, w http.ResponseWriter, r *http.Request) {
			body, err := io.ReadAll(r.Body)
			if err != nil || string(body) != "hello" {
				t.Errorf("body=%q, err=%v", body, err)
			}
			if r.Header.Get("Content-Type") != "text/plain; charset=utf-8" || r.Header.Get("X-Amz-Meta-Author") != "gox" {
				t.Errorf("headers: %v", r.Header)
			}
			w.Header().Set("ETag", `"etag"`)
		}, func(t *testing.T, s *Storage) {
			// 隐藏 Seek，确保接口接受任意 io.Reader。
			if err := s.Put(ctx, key, struct{ io.Reader }{strings.NewReader("hello")}, oss.WithMetadata(map[string]string{"author": "gox"})); err != nil {
				t.Fatal(err)
			}
		}},
		{"get range", "GET", func(t *testing.T, w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("Range") != "bytes=1-3" {
				t.Errorf("Range=%q", r.Header.Get("Range"))
			}
			w.WriteHeader(http.StatusPartialContent)
			_, _ = io.WriteString(w, "ell")
		}, func(t *testing.T, s *Storage) {
			body, err := s.Get(ctx, key, oss.WithRange(1, 3))
			if err != nil {
				t.Fatal(err)
			}
			defer body.Close()
			data, err := io.ReadAll(body)
			if err != nil || string(data) != "ell" {
				t.Fatalf("body=%q, err=%v", data, err)
			}
		}},
		{"get open range", "GET", func(t *testing.T, w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("Range") != "bytes=2-" {
				t.Errorf("Range=%q", r.Header.Get("Range"))
			}
		}, func(t *testing.T, s *Storage) {
			body, err := s.Get(ctx, key, oss.WithRange(2, -1))
			if err != nil {
				t.Fatal(err)
			}
			body.Close()
		}},
		{"stat", "HEAD", func(t *testing.T, w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Length", "5")
			w.Header().Set("Content-Type", "text/plain")
			w.Header().Set("ETag", `"etag"`)
			w.Header().Set("Last-Modified", "Mon, 02 Jan 2006 15:04:05 GMT")
			w.Header().Set("X-Amz-Meta-Author", "gox")
		}, func(t *testing.T, s *Storage) {
			info, err := s.Stat(ctx, key)
			if err != nil {
				t.Fatal(err)
			}
			if info.Key != key || info.Size != 5 || info.ContentType != "text/plain" || info.ETag != `"etag"` || info.Metadata["author"] != "gox" || info.LastModified.Year() != 2006 {
				t.Fatalf("info=%+v", info)
			}
		}},
		{"exists", "HEAD", func(t *testing.T, w http.ResponseWriter, r *http.Request) {}, func(t *testing.T, s *Storage) {
			if exists, err := s.Exists(ctx, key); !exists || err != nil {
				t.Fatalf("exists=%v, err=%v", exists, err)
			}
		}},
		{"missing", "HEAD", func(t *testing.T, w http.ResponseWriter, r *http.Request) { w.WriteHeader(404) }, func(t *testing.T, s *Storage) {
			if exists, err := s.Exists(ctx, key); exists || err != nil {
				t.Fatalf("exists=%v, err=%v", exists, err)
			}
		}},
		{"denied", "HEAD", func(t *testing.T, w http.ResponseWriter, r *http.Request) { w.WriteHeader(403) }, func(t *testing.T, s *Storage) {
			if exists, err := s.Exists(ctx, key); exists || !oss.IsAccessDenied(err) {
				t.Fatalf("exists=%v, err=%v", exists, err)
			}
		}},
		{"delete", "DELETE", func(t *testing.T, w http.ResponseWriter, r *http.Request) { w.WriteHeader(204) }, func(t *testing.T, s *Storage) {
			if err := s.Delete(ctx, key); err != nil {
				t.Fatal(err)
			}
		}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var calls atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls.Add(1)
				if r.Method != tt.method || r.URL.Path != "/test-bucket/"+key {
					t.Errorf("request: %s %s", r.Method, r.URL.Path)
				}
				if !strings.Contains(r.Header.Get("Authorization"), "/auto/s3/aws4_request") {
					t.Errorf("missing S3 signature: %v", r.Header)
				}
				if r.Header.Get("X-Amz-Sdk-Checksum-Algorithm") != "" || r.Header.Get("X-Amz-Trailer") != "" {
					t.Error("unexpected optional checksum headers")
				}
				tt.handle(t, w, r)
			}))
			defer server.Close()
			tt.call(t, testStorage(t, server.URL, WithHTTPClient(server.Client())))
			if calls.Load() != 1 {
				t.Fatalf("requests=%d", calls.Load())
			}
		})
	}
}

func TestList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if r.URL.Path != "/test-bucket" || q.Get("list-type") != "2" || q.Get("prefix") != "photos/" || q.Get("delimiter") != "/" || q.Get("max-keys") != "2" || q.Get("continuation-token") != "opaque+/=" {
			t.Errorf("request: %s", r.URL)
		}
		w.Header().Set("Content-Type", "application/xml")
		_, _ = io.WriteString(w, `<ListBucketResult><IsTruncated>true</IsTruncated><NextContinuationToken>next+/=</NextContinuationToken><Contents><Key>photos/a.jpg</Key><Size>42</Size><ETag>etag</ETag><LastModified>2026-01-02T03:04:05Z</LastModified></Contents><CommonPrefixes><Prefix>photos/sub/</Prefix></CommonPrefixes></ListBucketResult>`)
	}))
	defer server.Close()
	out, err := testStorage(t, server.URL).List(context.Background(), oss.WithPrefix("photos/"), oss.WithDelimiter("/"), oss.WithLimit(2), oss.WithToken("opaque+/="))
	if err != nil {
		t.Fatal(err)
	}
	if !out.Truncated || out.NextToken != "next+/=" || len(out.Objects) != 1 || out.Objects[0].Key != "photos/a.jpg" || out.Objects[0].Size != 42 || out.Objects[0].ContentType != "" || len(out.Prefixes) != 1 || out.Prefixes[0] != "photos/sub/" {
		t.Fatalf("result: %+v", out)
	}
}

func TestSignURL(t *testing.T) {
	for _, method := range []string{oss.MethodGet, oss.MethodPut, oss.MethodDelete} {
		t.Run(method, func(t *testing.T) {
			s := testStorage(t, "https://account.r2.cloudflarestorage.com", WithSecurityToken("session-token"))
			raw, err := s.SignURL(context.Background(), "目录/a +%.txt", oss.WithMethod(method), oss.WithExpires(time.Hour), oss.WithSignContentType("text/plain"), oss.WithSignContentDisposition(`attachment; filename="a.txt"`))
			if err != nil {
				t.Fatal(err)
			}
			u, err := url.Parse(raw)
			if err != nil {
				t.Fatal(err)
			}
			q := u.Query()
			if u.Path != "/test-bucket/目录/a +%.txt" || q.Get("X-Amz-Signature") == "" || q.Get("X-Amz-Expires") != "3600" || !strings.Contains(q.Get("X-Amz-Credential"), "/auto/s3/aws4_request") || q.Get("X-Amz-Security-Token") != "session-token" {
				t.Fatalf("URL: %s", raw)
			}
			if method == oss.MethodPut && !strings.Contains(q.Get("X-Amz-SignedHeaders"), "content-type") {
				t.Fatal("Content-Type is not signed")
			}
			if method == oss.MethodGet && q.Get("response-content-disposition") != `attachment; filename="a.txt"` {
				t.Fatalf("response-content-disposition: %s", raw)
			}
		})
	}
	s, err := New(WithBucket("test-bucket"), WithCredentials("id", "secret"))
	if err != nil {
		t.Fatal(err)
	}
	raw, err := s.SignURL(context.Background(), "key")
	if err != nil {
		t.Fatal(err)
	}
	u, err := url.Parse(raw)
	if err != nil || !strings.HasPrefix(u.Host, "test-bucket.s3.") || !strings.Contains(u.Query().Get("X-Amz-Credential"), "/us-east-1/s3/") || u.Query().Get("X-Amz-Expires") != "900" {
		t.Fatalf("AWS default URL: %s, err=%v", raw, err)
	}
}

func TestBuckets(t *testing.T) {
	for _, region := range []string{"auto", "us-east-1", "eu-west-1"} {
		t.Run(region, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				body, _ := io.ReadAll(r.Body)
				if r.Method != "PUT" || r.URL.Path != "/new-bucket" {
					t.Errorf("request: %s %s", r.Method, r.URL)
				}
				if region == "eu-west-1" {
					if !strings.Contains(string(body), "<LocationConstraint>eu-west-1</LocationConstraint>") {
						t.Errorf("body: %s", body)
					}
				} else if len(body) != 0 {
					t.Errorf("unexpected location: %s", body)
				}
				if r.Header.Get("X-Amz-Acl") != "private" {
					t.Error("missing ACL")
				}
			}))
			defer server.Close()
			if err := testStorage(t, server.URL, WithRegion(region)).CreateBucket(context.Background(), "new-bucket", oss.WithBucketACL("private")); err != nil {
				t.Fatal(err)
			}
		})
	}
	var pages atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "DELETE" {
			if r.URL.Path != "/empty-bucket" {
				t.Errorf("path=%s", r.URL.Path)
			}
			w.WriteHeader(204)
			return
		}
		if r.URL.Query().Get("max-buckets") != "1000" {
			t.Errorf("missing bucket page size: %s", r.URL)
		}
		w.Header().Set("Content-Type", "application/xml")
		if pages.Add(1) == 1 {
			_, _ = io.WriteString(w, `<ListAllMyBucketsResult><Buckets><Bucket><Name>one</Name><CreationDate>2026-01-02T03:04:05Z</CreationDate><BucketRegion>auto</BucketRegion></Bucket></Buckets><ContinuationToken>next</ContinuationToken></ListAllMyBucketsResult>`)
		} else {
			if r.URL.Query().Get("continuation-token") != "next" {
				t.Error("missing bucket continuation token")
			}
			_, _ = io.WriteString(w, `<ListAllMyBucketsResult><Buckets><Bucket><Name>two</Name></Bucket></Buckets></ListAllMyBucketsResult>`)
		}
	}))
	defer server.Close()
	s := testStorage(t, server.URL)
	buckets, err := s.ListBuckets(context.Background())
	if err != nil || len(buckets) != 2 || buckets[0].Name != "one" || buckets[0].Region != "auto" || buckets[1].Name != "two" || pages.Load() != 2 {
		t.Fatalf("buckets=%+v, err=%v", buckets, err)
	}
	if err := s.DeleteBucket(context.Background(), "empty-bucket"); err != nil {
		t.Fatal(err)
	}
}

func TestListEncodedNames(t *testing.T) {
	for _, tt := range []struct {
		name, encoding, key, prefix, wantKey, wantPrefix string
		wantError                                        bool
	}{
		{"encoded", "url", "%E7%9B%AE%E5%BD%95%2Fa%01+%252F", "%E7%9B%AE%E5%BD%95%2F+%25%2F", "目录/a\x01+%2F", "目录/+%/", false},
		{"unencoded", "", "a+%2F", "dir+%2F/", "a+%2F", "dir+%2F/", false},
		{"invalid key", "url", "%zz", "dir/", "", "", true},
		{"invalid prefix", "url", "key", "%zz", "", "", true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Query().Get("encoding-type") != "url" {
					t.Errorf("missing URL encoding: %s", r.URL)
				}
				w.Header().Set("Content-Type", "application/xml")
				fmt.Fprintf(w, `<ListBucketResult><EncodingType>%s</EncodingType><IsTruncated>true</IsTruncated><NextContinuationToken>opaque+%%2F</NextContinuationToken><Contents><Key>%s</Key></Contents><CommonPrefixes><Prefix>%s</Prefix></CommonPrefixes></ListBucketResult>`, tt.encoding, tt.key, tt.prefix)
			}))
			defer server.Close()
			out, err := testStorage(t, server.URL).List(context.Background())
			if tt.wantError {
				if !oss.IsCode(err, oss.ErrCodeInternal) {
					t.Fatalf("error=%v", err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if len(out.Objects) != 1 || out.Objects[0].Key != tt.wantKey || len(out.Prefixes) != 1 || out.Prefixes[0] != tt.wantPrefix || out.NextToken != "opaque+%2F" {
				t.Fatalf("result=%+v objects=%+v", out, out.Objects)
			}
		})
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestBucketRegionOverride(t *testing.T) {
	for _, endpoint := range []string{"", "https://storage.example.test"} {
		for _, region := range []string{"ap-southeast-1", "us-east-1"} {
			t.Run(endpoint+"/"+region, func(t *testing.T) {
				wantRegion := region
				client := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
					wantHost := "new-bucket.s3." + wantRegion + ".amazonaws.com"
					if endpoint != "" {
						wantHost = "new-bucket.storage.example.test"
					}
					if r.URL.Host != wantHost || !strings.Contains(r.Header.Get("Authorization"), "/"+wantRegion+"/s3/aws4_request") {
						t.Errorf("host=%s authorization=%s", r.URL.Host, r.Header.Get("Authorization"))
					}
					var body []byte
					if r.Body != nil {
						var err error
						body, err = io.ReadAll(r.Body)
						if err != nil {
							t.Error(err)
						}
					}
					if wantRegion == "us-east-1" {
						if len(body) != 0 {
							t.Errorf("unexpected location: %s", body)
						}
					} else if !strings.Contains(string(body), "<LocationConstraint>"+wantRegion+"</LocationConstraint>") {
						t.Errorf("location: %s", body)
					}
					return &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(strings.NewReader("")), Request: r}, nil
				})}
				s, err := New(WithEndpoint(endpoint), WithRegion("eu-west-1"), WithBucket("test-bucket"), WithCredentials("id", "secret"), WithHTTPClient(client))
				if err != nil {
					t.Fatal(err)
				}
				if err := s.CreateBucket(context.Background(), "new-bucket", oss.WithBucketRegion(region)); err != nil {
					t.Fatal(err)
				}
				wantRegion = "eu-west-1"
				if err := s.CreateBucket(context.Background(), "new-bucket"); err != nil {
					t.Fatal(err)
				}
			})
		}
	}
}

func TestR2BucketPagination(t *testing.T) {
	for _, mode := range []string{"pages", "missing token", "duplicate token"} {
		t.Run(mode, func(t *testing.T) {
			var calls atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				page := calls.Add(1)
				if page > 3 {
					t.Error("pagination did not stop")
					w.WriteHeader(400)
					return
				}
				if page > 1 && r.URL.Query().Get("continuation-token") != fmt.Sprintf("token+%%2F=%d", page-1) {
					t.Errorf("wrong continuation token: %s", r.URL)
				}
				w.Header().Set("Content-Type", "application/xml")
				next := fmt.Sprintf("token+%%2F=%d", page)
				if mode == "missing token" {
					next = ""
				}
				if mode == "duplicate token" {
					next = "token+%2F=1"
				}
				w.Header().Set("Cf-Is-Truncated", fmt.Sprint(page < 3))
				if page < 3 {
					w.Header().Set("Cf-Next-Continuation-Token", next)
				}
				// R2 的 ContinuationToken 是本次请求的令牌；下一页令牌另有字段和响应头。
				fmt.Fprintf(w, `<ListAllMyBucketsResult><Buckets><Bucket><Name>bucket-%d</Name></Bucket></Buckets><ContinuationToken>%s</ContinuationToken><IsTruncated>%t</IsTruncated><NextContinuationToken>%s</NextContinuationToken></ListAllMyBucketsResult>`, page, r.URL.Query().Get("continuation-token"), page < 3, next)
			}))
			defer server.Close()
			out, err := testStorage(t, server.URL).ListBuckets(context.Background())
			if mode != "pages" {
				if !oss.IsCode(err, oss.ErrCodeInternal) {
					t.Fatalf("error=%v", err)
				}
				return
			}
			if err != nil || len(out) != 3 || calls.Load() != 3 {
				t.Fatalf("buckets=%v calls=%d error=%v", out, calls.Load(), err)
			}
			for i, b := range out {
				if b.Name != fmt.Sprintf("bucket-%d", i+1) {
					t.Fatalf("bucket=%+v", b)
				}
			}
		})
	}
}

func TestErrors(t *testing.T) {
	for _, tt := range []struct {
		serviceCode, code string
		status            int
	}{
		{"NoSuchKey", oss.ErrCodeNotFound, 404},
		{"AccessDenied", oss.ErrCodeAccessDenied, 403},
		{"InvalidArgument", oss.ErrCodeInvalidArgument, 400},
		{"BucketNotEmpty", oss.ErrCodeBucketNotEmpty, 409},
		{"BucketAlreadyOwnedByYou", oss.ErrCodeBucketExists, 409},
		{"Unexpected", oss.ErrCodeInternal, 418},
	} {
		t.Run(tt.serviceCode, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/xml")
				w.WriteHeader(tt.status)
				fmt.Fprintf(w, "<Error><Code>%s</Code><Message>failure</Message></Error>", tt.serviceCode)
			}))
			defer server.Close()
			err := testStorage(t, server.URL).Delete(context.Background(), "key")
			var ossErr *oss.Error
			if !errors.As(err, &ossErr) || ossErr.Code != tt.code || ossErr.Key != "key" || errors.Unwrap(err) == nil {
				t.Fatalf("error=%v", err)
			}
		})
	}
}

func TestMultipartUpload(t *testing.T) {
	for _, fail := range []bool{false, true} {
		t.Run(fmt.Sprint("failure=", fail), func(t *testing.T) {
			var uploaded atomic.Int64
			var completed, aborted atomic.Bool
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				q := r.URL.Query()
				if r.Header.Get("X-Amz-Checksum-Algorithm") != "" || r.Header.Get("X-Amz-Trailer") != "" {
					t.Error("unexpected checksum headers")
				}
				w.Header().Set("Content-Type", "application/xml")
				switch {
				case r.Method == "POST" && q.Has("uploads"):
					if r.Header.Get("Content-Type") != "application/octet-stream" {
						t.Error("missing upload content type")
					}
					_, _ = io.WriteString(w, `<InitiateMultipartUploadResult><UploadId>upload-id</UploadId></InitiateMultipartUploadResult>`)
				case r.Method == "PUT" && q.Get("uploadId") == "upload-id":
					n, err := io.Copy(io.Discard, r.Body)
					if err != nil {
						t.Error(err)
					}
					uploaded.Add(n)
					if fail {
						w.WriteHeader(403)
						_, _ = io.WriteString(w, `<Error><Code>AccessDenied</Code></Error>`)
						return
					}
					w.Header().Set("ETag", `"part"`)
				case r.Method == "POST" && q.Get("uploadId") == "upload-id":
					completed.Store(true)
					_, _ = io.WriteString(w, `<CompleteMultipartUploadResult><ETag>complete</ETag></CompleteMultipartUploadResult>`)
				case r.Method == "DELETE" && q.Get("uploadId") == "upload-id":
					aborted.Store(true)
					w.WriteHeader(204)
				default:
					t.Errorf("unexpected multipart request: %s %s", r.Method, r.URL)
					w.WriteHeader(400)
				}
			}))
			defer server.Close()
			const size = 17 * 1024 * 1024
			err := testStorage(t, server.URL).Put(context.Background(), "large.bin", struct{ io.Reader }{bytes.NewReader(make([]byte, size))})
			if fail {
				if !oss.IsAccessDenied(err) || !aborted.Load() || completed.Load() {
					t.Fatalf("error=%v aborted=%v completed=%v", err, aborted.Load(), completed.Load())
				}
			} else if err != nil || uploaded.Load() != size || !completed.Load() || aborted.Load() {
				t.Fatalf("error=%v bytes=%d completed=%v", err, uploaded.Load(), completed.Load())
			}
		})
	}
}

func TestCancellation(t *testing.T) {
	started := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { close(started); <-r.Context().Done() }))
	defer server.Close()
	s := testStorage(t, server.URL)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	done := make(chan error, 1)
	go func() { _, err := s.Get(ctx, "key"); done <- err }()
	select {
	case <-started:
		cancel()
	case <-ctx.Done():
		t.Fatal("request did not reach server")
	}
	if err := <-done; !errors.Is(err, context.Canceled) {
		t.Fatalf("error=%v", err)
	}
	if err := s.Put(ctx, "key", strings.NewReader("data")); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}

func TestIntegration(t *testing.T) {
	endpoint, bucket := os.Getenv("GOX_S3_ENDPOINT"), os.Getenv("GOX_S3_BUCKET")
	id, secret := os.Getenv("GOX_S3_ACCESS_KEY_ID"), os.Getenv("GOX_S3_ACCESS_KEY_SECRET")
	if bucket == "" || id == "" || secret == "" {
		t.Skip("set GOX_S3_BUCKET, GOX_S3_ACCESS_KEY_ID and GOX_S3_ACCESS_KEY_SECRET to test a real service")
	}
	s, err := New(WithEndpoint(endpoint), WithRegion(os.Getenv("GOX_S3_REGION")),
		WithCredentials(id, secret), WithBucket(bucket), WithUsePathStyle(true),
		WithSecurityToken(os.Getenv("GOX_S3_SECURITY_TOKEN")))
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	key := fmt.Sprintf("gox-tests/%d.txt", time.Now().UnixNano())
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := s.Delete(ctx, key); err != nil {
			t.Errorf("cleanup: %v", err)
		}
	})
	if err := s.Put(ctx, key, strings.NewReader("hello"), oss.WithContentType("text/plain"), oss.WithMetadata(map[string]string{"author": "gox"})); err != nil {
		t.Fatal(err)
	}
	info, err := s.Stat(ctx, key)
	if err != nil || info.Size != 5 || info.Metadata["author"] != "gox" {
		t.Fatalf("stat=%+v, err=%v", info, err)
	}
	link, err := s.SignURL(ctx, key)
	if err != nil {
		t.Fatal(err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, link, nil)
	if err != nil {
		t.Fatal(err)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	data, err := io.ReadAll(res.Body)
	if err != nil || res.StatusCode != 200 || string(data) != "hello" {
		t.Fatalf("download: status=%d body=%q err=%v", res.StatusCode, data, err)
	}
	out, err := s.List(ctx, oss.WithPrefix(key))
	if err != nil || len(out.Objects) != 1 || out.Objects[0].Key != key {
		t.Fatalf("list=%+v, err=%v", out, err)
	}
	if err := s.Delete(ctx, key); err != nil {
		t.Fatal(err)
	}
	if exists, err := s.Exists(ctx, key); err != nil || exists {
		t.Fatalf("exists=%v, err=%v", exists, err)
	}
}
