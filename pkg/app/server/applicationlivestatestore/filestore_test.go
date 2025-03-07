// Copyright 2023 The PipeCD Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package applicationlivestatestore

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/pipe-cd/pipecd/pkg/filestore"
	"github.com/pipe-cd/pipecd/pkg/filestore/filestoretest"
	"github.com/pipe-cd/pipecd/pkg/model"
)

func TestFileStoreGet(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	store := filestoretest.NewMockStore(ctrl)

	testcases := []struct {
		name          string
		applicationID string
		content       string
		readerErr     error

		expected    *model.ApplicationLiveStateSnapshot
		expectedErr error
	}{
		{
			name:          "file not found in filestore",
			applicationID: "application-id",

			content:   "",
			readerErr: filestore.ErrNotFound,

			expectedErr: filestore.ErrNotFound,
		},
		{
			name:          "found file in filestore",
			applicationID: "application-id",

			content: `{
				"application_id": "application-id",
				"piped_id": "piped-id",
				"project_id": "project-id",
				"kind": 0,
				"kubernetes": {
					"resources": [
						{
							"id": "id-1",
							"name": "test-name",
							"api_version": "networking.k8s.io/v1beta1",
							"kind": "Ingress",
							"namespace": "default",
							"created_at": 1590000000,
							"updated_at": 1590000000
						},
						{
							"id": "id-2",
							"name": "test-name",
							"api_version": "v1",
							"kind": "Service",
							"namespace": "default",
							"created_at": 1590000000,
							"updated_at": 1590000000
						}
					]
				},
				"version": {
					"index": 1,
					"timestamp": 1590000000
				}
			}`,

			expected: &model.ApplicationLiveStateSnapshot{
				ApplicationId: "application-id",
				PipedId:       "piped-id",
				ProjectId:     "project-id",
				Kind:          model.ApplicationKind_KUBERNETES,
				Kubernetes: &model.KubernetesApplicationLiveState{
					Resources: []*model.KubernetesResourceState{
						{
							Id:         "id-1",
							Name:       "test-name",
							ApiVersion: "networking.k8s.io/v1beta1",
							Kind:       "Ingress",
							Namespace:  "default",
							CreatedAt:  1590000000,
							UpdatedAt:  1590000000,
						},
						{
							Id:         "id-2",
							Name:       "test-name",
							ApiVersion: "v1",
							Kind:       "Service",
							Namespace:  "default",
							CreatedAt:  1590000000,
							UpdatedAt:  1590000000,
						},
					},
				},
				Version: &model.ApplicationLiveStateVersion{
					Index:     1,
					Timestamp: 1590000000,
				},
			},
		},
	}

	fs := applicationLiveStateFileStore{
		backend: store,
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			path := applicationLiveStatePath(tc.applicationID)
			content := []byte(tc.content)

			store.EXPECT().Get(context.TODO(), path).Return(content, tc.readerErr)
			state, err := fs.Get(context.TODO(), tc.applicationID)
			if err != nil {
				if tc.expectedErr == nil {
					assert.NoError(t, err)
					return
				}
				assert.Error(t, err, tc.expectedErr)
				return
			}
			assert.Equal(t, tc.expected, state)
		})
	}
}
