// Copyright 2018-2021 CERN
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
//
// In applying this license, CERN does not waive the privileges and immunities
// granted to it by virtue of its status as an Intergovernmental Organization
// or submit itself to any jurisdiction.

package decomposedfs_test

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	userpb "github.com/cs3org/go-cs3apis/cs3/identity/user/v1beta1"
	provider "github.com/cs3org/go-cs3apis/cs3/storage/provider/v1beta1"
	helpers "github.com/cs3org/reva/v2/pkg/storage/utils/decomposedfs/testhelpers"
	"github.com/cs3org/reva/v2/pkg/storage/utils/decomposedfs/xattrs"
	"github.com/cs3org/reva/v2/pkg/storage/utils/decomposedfs/xattrs/prefixes"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
)

type testFS struct {
	root string
}

func (t testFS) Open(name string) (fs.File, error) {
	return os.Open(filepath.Join(t.root, name))
}

var _ = Describe("Grants", func() {
	var (
		env   *helpers.TestEnv
		ref   *provider.Reference
		grant *provider.Grant
		tfs   = &testFS{}
	)

	BeforeEach(func() {
		grant = &provider.Grant{
			Grantee: &provider.Grantee{
				Type: provider.GranteeType_GRANTEE_TYPE_USER,
				Id: &provider.Grantee_UserId{
					UserId: &userpb.UserId{
						OpaqueId: "4c510ada-c86b-4815-8820-42cdf82c3d51",
					},
				},
			},
			Permissions: &provider.ResourcePermissions{
				Stat:   true,
				Move:   true,
				Delete: false,
			},
			Creator: &userpb.UserId{
				OpaqueId: helpers.OwnerID,
			},
		}
	})

	JustBeforeEach(func() {
		var err error
		env, err = helpers.NewTestEnv(nil)
		Expect(err).ToNot(HaveOccurred())

		ref = &provider.Reference{
			ResourceId: env.SpaceRootRes,
			Path:       "/dir1",
		}

	})

	AfterEach(func() {
		if env != nil {
			env.Cleanup()
		}
	})

	Context("with no permissions", func() {
		JustBeforeEach(func() {
			env.Permissions.On("AssemblePermissions", mock.Anything, mock.Anything, mock.Anything).Return(provider.ResourcePermissions{}, nil)
		})

		Describe("AddGrant", func() {
			It("hides the resource", func() {
				err := env.Fs.AddGrant(env.Ctx, ref, grant)
				Expect(err).To(MatchError(ContainSubstring("not found")))
			})
		})
	})

	Context("with insufficient permissions", func() {
		JustBeforeEach(func() {
			env.Permissions.On("AssemblePermissions", mock.Anything, mock.Anything, mock.Anything).Return(provider.ResourcePermissions{
				Stat: true,
			}, nil)
		})

		Describe("AddGrant", func() {
			It("denies adding grants", func() {
				err := env.Fs.AddGrant(env.Ctx, ref, grant)
				Expect(err).To(MatchError(ContainSubstring("permission denied")))
			})
		})
	})

	Context("with sufficient permissions", func() {
		JustBeforeEach(func() {
			env.Permissions.On("AssemblePermissions", mock.Anything, mock.Anything, mock.Anything).Return(provider.ResourcePermissions{
				Stat:        true,
				AddGrant:    true,
				ListGrants:  true,
				RemoveGrant: true,
			}, nil)
		})

		Describe("AddGrant", func() {
			It("adds grants", func() {
				n, err := env.Lookup.NodeFromResource(env.Ctx, &provider.Reference{
					ResourceId: env.SpaceRootRes,
					Path:       "/dir1",
				})
				Expect(err).ToNot(HaveOccurred())

				err = env.Fs.AddGrant(env.Ctx, ref, grant)
				Expect(err).ToNot(HaveOccurred())

				o := env.Owner.GetId()
				localPath := n.InternalPath()
				attr, err := xattrs.Get(localPath, prefixes.GrantUserAcePrefix+grant.Grantee.GetUserId().OpaqueId)
				Expect(err).ToNot(HaveOccurred())
				Expect(attr).To(Equal(fmt.Sprintf("\x00t=A:f=:p=rw:c=%s:e=0\n", o.GetOpaqueId()))) // NOTE: this tests ace package
			})

			It("creates a storage space per created grant", func() {
				err := env.Fs.AddGrant(env.Ctx, ref, grant)
				Expect(err).ToNot(HaveOccurred())

				spaceTypesPath := filepath.Join(env.Root, "indexes", "by-type")
				tfs.root = spaceTypesPath
				entries, err := fs.ReadDir(tfs, "share")
				Expect(err).ToNot(HaveOccurred())
				Expect(len(entries)).To(BeNumerically(">=", 1))
			})
		})

		Describe("ListGrants", func() {
			It("lists existing grants", func() {
				err := env.Fs.AddGrant(env.Ctx, ref, grant)
				Expect(err).ToNot(HaveOccurred())

				grants, err := env.Fs.ListGrants(env.Ctx, ref)
				Expect(err).ToNot(HaveOccurred())
				Expect(len(grants)).To(Equal(1))

				g := grants[0]
				Expect(g.Grantee.GetUserId().OpaqueId).To(Equal(grant.Grantee.GetUserId().OpaqueId))
				Expect(g.Permissions.Stat).To(BeTrue())
				Expect(g.Permissions.Move).To(BeTrue())
				Expect(g.Permissions.Delete).To(BeFalse())
			})
		})

		Describe("RemoveGrants", func() {
			It("removes the grant", func() {
				err := env.Fs.AddGrant(env.Ctx, ref, grant)
				Expect(err).ToNot(HaveOccurred())

				grants, err := env.Fs.ListGrants(env.Ctx, ref)
				Expect(err).ToNot(HaveOccurred())
				Expect(len(grants)).To(Equal(1))

				err = env.Fs.RemoveGrant(env.Ctx, ref, grant)
				Expect(err).ToNot(HaveOccurred())

				grants, err = env.Fs.ListGrants(env.Ctx, ref)
				Expect(err).ToNot(HaveOccurred())
				Expect(len(grants)).To(Equal(0))
			})
		})
	})
})
