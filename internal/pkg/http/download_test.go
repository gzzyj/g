// Copyright (c) 2019 voidint <voidint@126.com>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy of
// this software and associated documentation files (the "Software"), to deal in
// the Software without restriction, including without limitation the rights to
// use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
// the Software, and to permit persons to whom the Software is furnished to do so,
// subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
// FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
// COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
// IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
// CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package http

import (
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/gzzyj/g/internal/pkg/errs"
	"github.com/stretchr/testify/assert"
)

func TestIsSuccess(t *testing.T) {
	type args struct {
		statusCode int
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: http.StatusText(http.StatusOK),
			args: args{http.StatusOK},
			want: true,
		},
		{
			name: http.StatusText(http.StatusCreated),
			args: args{http.StatusCreated},
			want: true,
		},
		{
			name: http.StatusText(http.StatusIMUsed),
			args: args{http.StatusIMUsed},
			want: true,
		},
		{
			name: http.StatusText(http.StatusContinue),
			args: args{http.StatusContinue},
			want: false,
		},
		{
			name: http.StatusText(http.StatusMultipleChoices),
			args: args{http.StatusMultipleChoices},
			want: false,
		},
		{
			name: http.StatusText(http.StatusMovedPermanently),
			args: args{http.StatusMovedPermanently},
			want: false,
		},
		{
			name: http.StatusText(http.StatusBadRequest),
			args: args{http.StatusBadRequest},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, IsSuccess(tt.args.statusCode))
		})
	}
}

func TestDownload(t *testing.T) {
	e := errors.New("unknown error")
	url := "http://github.com/gzzyj/g"
	filename := fmt.Sprintf("%d.txt", time.Now().Unix())
	defer os.Remove(filename)

	rr1 := httptest.NewRecorder()
	rr1.WriteHeader(http.StatusNotFound)

	rr2 := httptest.NewRecorder()
	rr2.WriteHeader(http.StatusOK)
	_, _ = rr2.WriteString("hello world")

	patches := gomonkey.ApplyMethodSeq(&http.Client{}, "Do", []gomonkey.OutputCell{
		{Values: gomonkey.Params{nil, e}},
		{Values: gomonkey.Params{rr1.Result(), nil}},
		{Values: gomonkey.Params{rr2.Result(), nil}},
	})
	defer patches.Reset()

	type args struct {
		srcURL       string
		filename     string
		flag         int
		perm         fs.FileMode
		withProgress bool
	}
	tests := []struct {
		name     string
		args     args
		wantSize int64
		wantErr  error
	}{
		{
			name: "发送请求返回异常响应",
			args: args{
				srcURL:       url,
				filename:     filename,
				flag:         os.O_RDWR | os.O_CREATE,
				perm:         0600,
				withProgress: true,
			},
			wantSize: 0,
			wantErr:  errs.NewDownloadError(url, e),
		},
		{
			name: "资源不存在",
			args: args{
				srcURL:       url,
				filename:     filename,
				flag:         os.O_RDWR | os.O_CREATE,
				perm:         0600,
				withProgress: true,
			},
			wantSize: 0,
			wantErr:  errs.NewURLUnreachableError(url, fmt.Errorf("%d", http.StatusNotFound)),
		},
		{
			name: "下载资源成功",
			args: args{
				srcURL:       url,
				filename:     filename,
				flag:         os.O_RDWR | os.O_CREATE,
				perm:         0600,
				withProgress: true,
			},
			wantSize: int64(len([]byte("hello world"))),
			wantErr:  nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotSize, err := Download(tt.args.srcURL, tt.args.filename, tt.args.flag, tt.args.perm, tt.args.withProgress)
			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.wantSize, gotSize)
		})
	}
}

func TestDownloadAsBytes(t *testing.T) {
	e := errors.New("unknown error")
	url := "http://github.com/gzzyj/g"

	rr := httptest.NewRecorder()
	rr.WriteHeader(http.StatusOK)
	_, _ = rr.WriteString("hello world")

	patches := gomonkey.ApplyMethodSeq(&http.Client{}, "Get", []gomonkey.OutputCell{
		{Values: gomonkey.Params{nil, e}},
		{Values: gomonkey.Params{rr.Result(), nil}},
	})
	defer patches.Reset()

	tests := []struct {
		name     string
		url      string
		wantData []byte
		wantErr  error
	}{
		{
			name:     "发送请求返回异常响应",
			url:      url,
			wantData: nil,
			wantErr:  errs.NewDownloadError(url, e),
		},
		{
			name:     "发送请求并得到正常响应",
			url:      url,
			wantData: []byte("hello world"),
			wantErr:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := DownloadAsBytes(tt.url)
			assert.Equal(t, err, tt.wantErr)
			assert.Equal(t, data, tt.wantData)
		})
	}
}
