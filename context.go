// Copyright © 2024 Meroxa, Inc.
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

package ecdysis

import (
	"context"

	"github.com/spf13/cobra"
)

type cobraCmdCtxKey struct{}

// ContextWithCobraCommand provides the cobra command to the context.
// This is useful for situations such as wanting to execute cmd.Help() directly from Execute().
func ContextWithCobraCommand(ctx context.Context, cmd *cobra.Command) context.Context {
	return context.WithValue(ctx, cobraCmdCtxKey{}, cmd)
}

// CobraCmdFromContext fetches the cobra command from the context. If the
// context does not contain a cobra command, it returns nil.
func CobraCmdFromContext(ctx context.Context) *cobra.Command {
	if cobraCmd := ctx.Value(cobraCmdCtxKey{}); cobraCmd != nil {
		return cobraCmd.(*cobra.Command) //nolint:forcetypeassert // only this package can set the value, it has to be a *cobra.Command
	}
	return nil
}
