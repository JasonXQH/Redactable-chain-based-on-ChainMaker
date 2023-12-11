/*
Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/
package scheduler

import (
	"reflect"
	"testing"

	"github.com/holiman/uint256"
)

/*
 * test unit NewConflictsBitWindow func
 */
func TestNewConflictsBitWindow(t *testing.T) {
	type args struct {
		txBatchSize int
	}
	tests := []struct {
		name string
		args args
		want *ConflictsBitWindow
	}{
		{
			name: "test0",
			args: args{
				txBatchSize: 1,
			},
			want: &ConflictsBitWindow{
				bitWindow:         uint256.NewInt(0),
				bitWindowCapacity: AdjustWindowSize,
				maxPoolCapacity:   1,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewConflictsBitWindow(tt.args.txBatchSize); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewConflictsBitWindow() = %v, want %v", got, tt.want)
			}
		})
	}
}

/*
 * test unit ConflictsBitWindow Enqueue func
 */
func TestConflictsBitWindow_Enqueue(t *testing.T) {
	type fields struct {
		bitWindow         *uint256.Int
		bitWindowCapacity int
		maxPoolCapacity   int
		conflictsNum      int
		execCount         int
	}
	type args struct {
		v                TxExecType
		currPoolCapacity int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   int
	}{
		{
			name: "test0",
			fields: fields{
				bitWindow:         uint256.NewInt(0),
				bitWindowCapacity: AdjustWindowSize,
				maxPoolCapacity:   1,
				conflictsNum:      1,
			},
			args: args{
				v: ConflictTx,
			},
			want: -1,
		},
		{
			name: "test1",
			fields: fields{
				bitWindow:         uint256.NewInt(5),
				bitWindowCapacity: 1,
				maxPoolCapacity:   5,
				conflictsNum:      1,
			},
			args: args{
				v: ConflictTx,
			},
			want: 2,
		},
		{
			name: "test2",
			fields: fields{
				bitWindow:         uint256.NewInt(10),
				bitWindowCapacity: AdjustWindowSize,
			},
			args: args{
				v: ConflictTx,
			},
			want: -1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := &ConflictsBitWindow{
				bitWindow:         tt.fields.bitWindow,
				bitWindowCapacity: tt.fields.bitWindowCapacity,
				maxPoolCapacity:   tt.fields.maxPoolCapacity,
				conflictsNum:      tt.fields.conflictsNum,
				execCount:         tt.fields.execCount,
			}
			if got := q.Enqueue(tt.args.v, tt.args.currPoolCapacity); got != tt.want {
				t.Errorf("Enqueue() = %v, want %v", got, tt.want)
			}
		})
	}
}

/*
 * test unit ConflictsBitWindow getNewPoolCapacity func
 */
func TestConflictsBitWindow_getNewPoolCapacity(t *testing.T) {
	type fields struct {
		bitWindow         *uint256.Int
		bitWindowCapacity int
		maxPoolCapacity   int
		conflictsNum      int
		execCount         int
	}

	// float64(q.conflictsNum) / float64(q.bitWindowCapacity)
	type args struct {
		currPoolCapacity int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   int
	}{
		{
			name: "test0", // conflictsRate < BaseConflictRate = 0.05
			fields: fields{
				conflictsNum:      1,
				bitWindowCapacity: 100,
			},
			args: args{
				currPoolCapacity: 1,
			},
			want: 0, // targetCapacity = int(float64(currPoolCapacity) * AscendCoefficient)
		},
		{
			name: "test1", // conflictsRate < BaseConflictRate = 0.05
			fields: fields{
				conflictsNum:      1,
				bitWindowCapacity: 100,
				maxPoolCapacity:   2,
			},
			args: args{
				currPoolCapacity: 1,
			},
			want: 2, // if targetCapacity > q.maxPoolCapacity { return q.maxPoolCapacity}
		},
		{
			name: "test2", // conflictsRate > TopConflictRate = 0.2
			fields: fields{
				conflictsNum:      1,
				bitWindowCapacity: 2,
				maxPoolCapacity:   2,
			},
			args: args{
				currPoolCapacity: 1,
			},
			want: 2, // targetCapacity = int(float64(currPoolCapacity) * DescendCoefficient)
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := &ConflictsBitWindow{
				bitWindow:         tt.fields.bitWindow,
				bitWindowCapacity: tt.fields.bitWindowCapacity,
				maxPoolCapacity:   tt.fields.maxPoolCapacity,
				conflictsNum:      tt.fields.conflictsNum,
				execCount:         tt.fields.execCount,
			}
			if got := q.getNewPoolCapacity(tt.args.currPoolCapacity); got != tt.want {
				t.Errorf("getNewPoolCapacity() = %v, want %v", got, tt.want)
			}
		})
	}
}

/*
 * test unit ConflictsBitWindow getConflictsRate func
 */
func TestConflictsBitWindow_getConflictsRate(t *testing.T) {
	type fields struct {
		bitWindow         *uint256.Int
		bitWindowCapacity int
		maxPoolCapacity   int
		conflictsNum      int
		execCount         int
	}
	tests := []struct {
		name   string
		fields fields
		want   float64
	}{
		{
			name: "test0",
			fields: fields{
				bitWindow:         uint256.NewInt(0),
				bitWindowCapacity: 640,
				maxPoolCapacity:   1,
				conflictsNum:      32,
			},
			want: BaseConflictRate,
		},
		{
			name: "test1",
			fields: fields{
				bitWindow:         uint256.NewInt(0),
				bitWindowCapacity: AdjustWindowSize * 10,
				maxPoolCapacity:   1,
				conflictsNum:      0,
			},
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := &ConflictsBitWindow{
				bitWindow:         tt.fields.bitWindow,
				bitWindowCapacity: tt.fields.bitWindowCapacity,
				maxPoolCapacity:   tt.fields.maxPoolCapacity,
				conflictsNum:      tt.fields.conflictsNum,
				execCount:         tt.fields.execCount,
			}
			if got := q.getConflictsRate(); got != tt.want {
				t.Errorf("getConflictsRate() = %v, want %v", got, tt.want)
			}
		})
	}
}

/*
 * test unit ConflictsBitWindow setMaxPoolCapacity func
 */
func TestConflictsBitWindow_setMaxPoolCapacity(t *testing.T) {
	type fields struct {
		bitWindow         *uint256.Int
		bitWindowCapacity int
		maxPoolCapacity   int
		conflictsNum      int
		execCount         int
	}
	type args struct {
		maxPoolCapacity int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "test0",
			fields: fields{
				bitWindow:         uint256.NewInt(0),
				bitWindowCapacity: AdjustWindowSize,
				maxPoolCapacity:   1,
			},
			args: args{
				maxPoolCapacity: 1,
			},
		},
		{
			name: "test1",
			fields: fields{
				bitWindow:         uint256.NewInt(0),
				bitWindowCapacity: AdjustWindowSize,
				maxPoolCapacity:   5,
			},
			args: args{
				maxPoolCapacity: 5,
			},
		},
		{
			name: "test2",
			fields: fields{
				bitWindow:         uint256.NewInt(0),
				bitWindowCapacity: AdjustWindowSize,
				maxPoolCapacity:   5,
			},
			args: args{
				maxPoolCapacity: MinPoolCapacity,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := &ConflictsBitWindow{
				bitWindow:         tt.fields.bitWindow,
				bitWindowCapacity: tt.fields.bitWindowCapacity,
				maxPoolCapacity:   tt.fields.maxPoolCapacity,
				conflictsNum:      tt.fields.conflictsNum,
				execCount:         tt.fields.execCount,
			}
			q.setMaxPoolCapacity(tt.args.maxPoolCapacity)
		})
	}
}
