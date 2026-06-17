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

package collector

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/gzzyj/g/internal/collector/autoindex"
	"github.com/gzzyj/g/internal/collector/fancyindex"
	"github.com/gzzyj/g/internal/collector/official"
	"github.com/gzzyj/g/internal/pkg/errs"
	"github.com/stretchr/testify/assert"
)

func TestNewCollector(t *testing.T) {
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader("hello world")),
	}

	patches := gomonkey.ApplyFuncReturn(http.Get, resp, nil)
	defer patches.Reset()

	type args struct {
		urls []string
	}
	tests := []struct {
		name              string
		args              args
		wantCollectorName string
		wantErr           error
	}{
		{
			name:              "nil parameter",
			args:              args{urls: nil},
			wantCollectorName: official.Name,
		},
		{
			name:              "A slice containing an empty string",
			args:              args{urls: []string{""}},
			wantCollectorName: official.Name,
		},
		{
			name:              "The parameter is a URL slice without a trailing backslash",
			args:              args{urls: []string{"https://mirrors.aliyun.com/golang"}},
			wantCollectorName: fancyindex.Name,
		},
		{
			name:              "A slice containing the name of the official collector",
			args:              args{urls: []string{"official|https://golang.google.cn/dl/"}},
			wantCollectorName: official.Name,
		},
		{
			name:              "A slice containing the name of the fancyindex collector",
			args:              args{urls: []string{"fancyindex|https://mirrors.hust.edu.cn/golang/"}},
			wantCollectorName: fancyindex.Name,
		},
		{
			name:              "A slice containing the name of the autoindex collector",
			args:              args{urls: []string{"autoindex|https://mirrors.ustc.edu.cn/golang/"}},
			wantCollectorName: autoindex.Name,
		},
		{
			name:              "A slice containing only official collector URLs",
			args:              args{urls: []string{OfficialDownloadPageURL}},
			wantCollectorName: official.Name,
		},
		{
			name:              "A slice containing only original official collector URLs",
			args:              args{urls: []string{OriginalOfficialDownloadPageURL}},
			wantCollectorName: official.Name,
		},
		{
			name:              "A slice containing only china official mirror site collector URLs",
			args:              args{urls: []string{CNDownloadPageURL}},
			wantCollectorName: official.Name,
		},
		{
			name:              "A slice containing only Alibaba cloud mirror site collector URLs",
			args:              args{urls: []string{AliYunDownloadPageURL}},
			wantCollectorName: fancyindex.Name,
		},
		{
			name:              "A slice containing only HUST mirror site collector URLs",
			args:              args{urls: []string{HUSTDownloadPageURL}},
			wantCollectorName: fancyindex.Name,
		},
		{
			name:              "A slice containing only NJU mirror site collector URLs",
			args:              args{urls: []string{NJUDownloadPageURL}},
			wantCollectorName: fancyindex.Name,
		},
		{
			name:              "A slice containing only USTC mirror site collector URLs",
			args:              args{urls: []string{USTCDownloadPageURL}},
			wantCollectorName: autoindex.Name,
		},
		{
			name:              "Collector not found",
			args:              args{urls: []string{"hello world"}},
			wantCollectorName: "",
			wantErr:           errs.ErrCollectorNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotC, err := NewCollector(tt.args.urls...)

			assert.Equal(t, tt.wantErr, err)

			if err == nil {
				assert.Equal(t, tt.wantCollectorName, gotC.Name())
			}
		})
	}
}
