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

package github

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/Masterminds/semver/v3"
	"github.com/agiledragon/gomonkey/v2"
	"github.com/gzzyj/g/internal/pkg/errs"
	"github.com/stretchr/testify/assert"
)

func TestAsset_IsCompressedFile(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		want        bool
	}{
		{
			name:        "application/zip",
			contentType: "application/zip",
			want:        true,
		},
		{
			name:        "application/x-gzip",
			contentType: "application/x-gzip",
			want:        true,
		},
		{
			name:        "application/json",
			contentType: "application/json",
			want:        false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := Asset{
				ContentType: tt.contentType,
			}
			assert.Equal(t, tt.want, a.IsCompressedFile())
		})
	}
}

func TestReleaseUpdater_CheckForUpdates(t *testing.T) {
	current := semver.MustParse("1.5.2")

	rr2 := httptest.NewRecorder()
	rr2.WriteHeader(http.StatusBadRequest)

	rr3 := httptest.NewRecorder()
	rr3.WriteHeader(http.StatusOK)
	_, _ = rr3.WriteString(`{"tag_name": 7}`)

	rr4 := httptest.NewRecorder()
	rr4.WriteHeader(http.StatusOK)
	_, _ = rr4.WriteString(`{"tag_name": "HelloWorld"}`)

	rr5 := httptest.NewRecorder()
	rr5.WriteHeader(http.StatusOK)
	_, _ = rr5.WriteString(`{"tag_name": "1.5.2"}`)

	rr6 := httptest.NewRecorder()
	rr6.WriteHeader(http.StatusOK)
	_, _ = rr6.WriteString(`{"tag_name": "1.6.0"}`)

	patches := gomonkey.ApplyMethodSeq(&http.Client{}, "Do", []gomonkey.OutputCell{
		{Values: gomonkey.Params{nil, errors.New("unknown error")}}, // Case 1: Failed to send query request
		{Values: gomonkey.Params{rr2.Result(), nil}},                // Case 2: Received non-success response
		{Values: gomonkey.Params{rr3.Result(), nil}},                // Case 3: Response deserialization error
		{Values: gomonkey.Params{rr4.Result(), nil}},                // Case 4: Non-semantic version in response
		{Values: gomonkey.Params{rr5.Result(), nil}},                // Case 5: No newer version available
		{Values: gomonkey.Params{rr6.Result(), nil}},                // Case 6: New version exists
	})
	defer patches.Reset()

	owner := "voidint"
	repo := "g"
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", owner, repo)

	tests := []struct {
		name    string
		current *semver.Version
		wantRel *Release
		wantYes bool
		wantErr error
	}{
		{
			name:    "1、发送查询请求失败",
			current: current,
			wantRel: nil,
			wantYes: false,
			wantErr: errors.New("unknown error"),
		},
		{
			name:    "2、得到非成功响应",
			current: current,
			wantRel: nil,
			wantYes: false,
			wantErr: errs.NewURLUnreachableError(url, fmt.Errorf("%d", http.StatusBadRequest)),
		},
		{
			name:    "3、响应内容反序列化错误",
			current: current,
			wantRel: nil,
			wantYes: false,
			wantErr: errors.New("json: cannot unmarshal number into Go struct field Release.tag_name of type string"),
		},
		{
			name:    "4、响应内容中版本号为非语义化版本号",
			current: current,
			wantRel: nil,
			wantYes: false,
			wantErr: semver.ErrInvalidSemVer,
		},
		{
			name:    "5、响应内容中版本号不大于当前版本号",
			current: current,
			wantRel: nil,
			wantYes: false,
			wantErr: nil,
		},
		{
			name:    "6、存在新版本号",
			current: current,
			wantRel: &Release{TagName: "1.6.0"},
			wantYes: true,
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rel, yes, err := ReleaseUpdater{}.CheckForUpdates(tt.current, owner, repo)
			if err != nil {
				assert.Equal(t, tt.wantErr.Error(), err.Error())
			}
			assert.True(t, reflect.DeepEqual(rel, tt.wantRel))
			assert.Equal(t, tt.wantYes, yes)
		})
	}
}
