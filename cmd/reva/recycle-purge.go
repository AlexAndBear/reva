// Copyright 2018-2023 CERN
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

package main

import (
	"io"

	rpc "github.com/cs3org/go-cs3apis/cs3/rpc/v1beta1"
	provider "github.com/cs3org/go-cs3apis/cs3/storage/provider/v1beta1"
)

func recyclePurgeCommand() *command {
	cmd := newCommand("recycle-purge")
	cmd.Description = func() string { return "purge a recycle bin" }
	cmd.Usage = func() string { return "Usage: recycle-purge [-flags] " }

	cmd.Action = func(w ...io.Writer) error {
		client, err := getClient()
		if err != nil {
			return err
		}

		ctx := getAuthContext()

		getHomeRes, err := client.GetHome(ctx, &provider.GetHomeRequest{})
		if err != nil {
			return err
		}

		req := &provider.PurgeRecycleRequest{
			Ref: &provider.Reference{
				Path: getHomeRes.Path,
			},
		}

		res, err := client.PurgeRecycle(ctx, req)
		if err != nil {
			return err
		}

		if res.Status.Code != rpc.Code_CODE_OK {
			return formatError(res.Status)
		}

		return nil
	}
	return cmd
}
