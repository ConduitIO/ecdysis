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

//go:generate mockgen -source=ecdysis_test.go -destination=behavioral_mock_test.go -package=ecdysis -typed

package ecdysis

import (
	"context"
	"fmt"
	"log/slog"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/spf13/cobra"
	"go.uber.org/mock/gomock"
)

type testCmd struct {
	flagLongFoo string
}

var (
	_ CommandWithDocs        = (*testCmd)(nil)
	_ CommandWithAliases     = (*testCmd)(nil)
	_ CommandWithFlags       = (*testCmd)(nil)
	_ CommandWithSubCommands = (*testCmd)(nil)
)

func (c *testCmd) Usage() string {
	return "cmd1"
}

func (c *testCmd) Aliases() []string {
	return []string{"foo", "bar"}
}

func (c *testCmd) Docs() Docs {
	return Docs{
		Short:   "short-foo",
		Long:    "long-bar",
		Example: "example-baz",
	}
}

func (c *testCmd) Flags() []Flag {
	return []Flag{
		{Long: "long-foo", Short: "l", Usage: "test flag", Required: false, Persistent: false, Ptr: &c.flagLongFoo},
	}
}

func (c *testCmd) SubCommands(e *Ecdysis) []*cobra.Command {
	return []*cobra.Command{
		e.MustBuildCobraCommand(&subCmd{}),
	}
}

type subCmd struct{}

func (c *subCmd) Usage() string {
	return "subCmd"
}

func TestBuildCobraCommand_Structural(t *testing.T) {
	ecdysis := New()
	cmd := &testCmd{}

	want := &cobra.Command{
		Use:     "cmd1",
		Aliases: []string{"foo", "bar"},
		Short:   "short-foo",
		Long:    "long-bar",
		Example: "example-baz",
	}
	want.Flags().StringVarP(&cmd.flagLongFoo, "long-foo", "l", "", "test flag")
	want.AddCommand(&cobra.Command{Use: "subCmd"})

	got := ecdysis.MustBuildCobraCommand(cmd)

	// Since we can't compare functions, we ignore RunE (coming from `buildCommandEvent`)
	got.RunE = nil

	// Since we can't compare functions, we ignore PostRunE (coming from `buildCommandAutoUpdate`)
	got.PostRunE = nil

	if v := cmp.Diff(got, want, cmpopts.IgnoreUnexported(cobra.Command{})); v != "" {
		t.Fatal(v)
	}
}

// BehavioralTestCommand is an interface out of which a mock will be generated.
type BehavioralTestCommand interface {
	CommandWithArgs
	CommandWithLogger
	CommandWithExecute
}

func TestBuildCobraCommand_Behavioral(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	cmd := NewMockBehavioralTestCommand(ctrl)

	wantLogger := slog.New(slog.NewTextHandler(nil, nil))
	ecdysis := New(WithDecorators(CommandWithLoggerDecorator{Logger: wantLogger}))

	// When building we only expect Usage and Logger to be called.
	call := cmd.EXPECT().Usage().Return("mock").Call
	call = cmd.EXPECT().Logger(wantLogger).After(call)

	got := ecdysis.MustBuildCobraCommand(cmd)

	// Set up the remaining expectations before executing the command.
	call = cmd.EXPECT().Args(gomock.Any()).Return(nil).After(call)
	cmd.EXPECT().Execute(ctx).Return(nil).After(call)

	err := got.ExecuteContext(ctx)
	if err != nil {
		t.Fatalf("not expected error, got %q", err.Error())
	}
}

type testCmdWithFlags struct {
	flag1  string
	flag2  int
	flag3  int8
	flag4  int16
	flag5  int32
	flag6  int64
	flag7  float32
	flag8  float64
	flag9  bool
	flag10 time.Duration
	flag11 []bool
	flag12 []float32
	flag13 []float64
	flag14 []int32
	flag15 []int64
	flag16 []int
	flag17 []string
}

var _ CommandWithFlags = (*testCmdWithFlags)(nil)

func (t *testCmdWithFlags) Usage() string {
	return "testCmdWithFlags"
}

func (t *testCmdWithFlags) Flags() []Flag {
	return []Flag{
		{Long: "flag1", Short: "a", Usage: "flag1 usage", Required: true, Persistent: false, Ptr: &t.flag1},
		{Long: "flag2", Short: "b", Usage: "flag2 usage", Required: false, Persistent: true, Ptr: &t.flag2},
		{Long: "flag3", Short: "c", Usage: "flag3 usage", Required: true, Persistent: false, Ptr: &t.flag3},
		{Long: "flag4", Short: "d", Usage: "flag4 usage", Required: false, Persistent: true, Ptr: &t.flag4},
		{Long: "flag5", Short: "e", Usage: "flag5 usage", Required: true, Persistent: false, Ptr: &t.flag5},
		{Long: "flag6", Short: "f", Usage: "flag6 usage", Required: false, Persistent: true, Ptr: &t.flag6},
		{Long: "flag7", Short: "g", Usage: "flag7 usage", Required: true, Persistent: false, Ptr: &t.flag7},
		{Long: "flag8", Short: "h", Usage: "flag8 usage", Required: false, Persistent: true, Ptr: &t.flag8},
		{Long: "flag9", Short: "i", Usage: "flag9 usage", Required: true, Persistent: false, Ptr: &t.flag9},
		{Long: "flag10", Short: "j", Usage: "flag10 usage", Required: false, Persistent: true, Ptr: &t.flag10},
		{Long: "flag11", Short: "k", Usage: "flag11 usage", Required: true, Persistent: false, Ptr: &t.flag11},
		{Long: "flag12", Short: "l", Usage: "flag12 usage", Required: false, Persistent: true, Ptr: &t.flag12},
		{Long: "flag13", Short: "m", Usage: "flag13 usage", Required: true, Persistent: false, Ptr: &t.flag13},
		{Long: "flag14", Short: "n", Usage: "flag14 usage", Required: false, Persistent: true, Ptr: &t.flag14},
		{Long: "flag15", Short: "o", Usage: "flag15 usage", Required: true, Persistent: false, Ptr: &t.flag15},
		{Long: "flag16", Short: "p", Usage: "flag16 usage", Required: false, Persistent: true, Ptr: &t.flag16},
		{Long: "flag17", Short: "q", Usage: "flag17 usage", Required: true, Persistent: false, Ptr: &t.flag17},
	}
}

func TestBuildCommandWithFlags(t *testing.T) {
	ecdysis := New()
	cmd := &testCmdWithFlags{}

	want := &cobra.Command{Use: "testCmdWithFlags"}
	want.Flags().StringVarP(&cmd.flag1, "flag1", "a", "", "flag1 usage")
	want.PersistentFlags().IntVarP(&cmd.flag2, "flag2", "b", 0, "flag2 usage")
	want.Flags().Int8VarP(&cmd.flag3, "flag3", "c", 0, "flag3 usage")
	want.PersistentFlags().Int16VarP(&cmd.flag4, "flag4", "d", 0, "flag4 usage")
	want.Flags().Int32VarP(&cmd.flag5, "flag5", "e", 0, "flag5 usage")
	want.PersistentFlags().Int64VarP(&cmd.flag6, "flag6", "f", 0, "flag6 usage")
	want.Flags().Float32VarP(&cmd.flag7, "flag7", "g", 0, "flag7 usage")
	want.PersistentFlags().Float64VarP(&cmd.flag8, "flag8", "h", 0, "flag8 usage")
	want.Flags().BoolVarP(&cmd.flag9, "flag9", "i", false, "flag9 usage")
	want.PersistentFlags().DurationVarP(&cmd.flag10, "flag10", "j", 0, "flag10 usage")
	want.Flags().BoolSliceVarP(&cmd.flag11, "flag11", "k", nil, "flag11 usage")
	want.PersistentFlags().Float32SliceVarP(&cmd.flag12, "flag12", "l", nil, "flag12 usage")
	want.Flags().Float64SliceVarP(&cmd.flag13, "flag13", "m", nil, "flag13 usage")
	want.PersistentFlags().Int32SliceVarP(&cmd.flag14, "flag14", "n", nil, "flag14 usage")
	want.Flags().Int64SliceVarP(&cmd.flag15, "flag15", "o", nil, "flag15 usage")
	want.PersistentFlags().IntSliceVarP(&cmd.flag16, "flag16", "p", nil, "flag16 usage")
	want.Flags().StringSliceVarP(&cmd.flag17, "flag17", "q", nil, "flag17 usage")

	for i := 1; i <= 17; i++ {
		if i%2 == 1 {
			_ = want.MarkFlagRequired(fmt.Sprintf("flag%d", i))
		}
	}

	got := ecdysis.MustBuildCobraCommand(cmd)

	// Since we can't compare functions, we ignore RunE (coming from `buildCommandEvent`)
	got.RunE = nil

	// Since we can't compare functions, we ignore PostRunE (coming from `buildCommandAutoUpdate`)
	got.PostRunE = nil

	if v := cmp.Diff(got, want, cmpopts.IgnoreUnexported(cobra.Command{})); v != "" {
		t.Fatal(v)
	}
}
