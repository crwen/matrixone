// Copyright 2021 Matrix Origin
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package plan

import (
	"fmt"

	"github.com/matrixorigin/matrixone/pkg/container/types"
	"github.com/matrixorigin/matrixone/pkg/sql/parsers/tree"
)

func (b *build) BuildExplainQuery(stmt *tree.ExplainStmt, plan *ExplainQuery) error {
	panic("implement me")
	// return nil
}

func BuildExplainResultColumns() []*Attribute {
	return []*Attribute{
		{
			Ref:  1,
			Name: fmt.Sprintf("QUERY PLAN"),
			Type: types.Type{
				Oid:  types.T_varchar,
				Size: 24,
			},
		},
	}
}
