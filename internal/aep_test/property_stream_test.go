package aep_test

import (
	"testing"

	"github.com/yueli-fx/aep-parser/internal/codec"
)

func TestPropertyStream_InitialStaticMode(t *testing.T) {
	ps := codec.NewPropertyStream[float64]()
	if ps.Mode() != codec.StreamModeStatic {
		t.Fatalf("initial mode = %v, want Static", ps.Mode())
	}
	v, isStatic := ps.StaticValue()
	if !isStatic {
		t.Fatalf("StaticValue() ok = false, want true (initial Static mode)")
	}
	if v != 0.0 {
		t.Fatalf("initial static = %v, want zero", v)
	}
}

func TestPropertyStream_SetStaticValue(t *testing.T) {
	ps := codec.NewPropertyStream[float64]()
	if err := ps.SetStaticValue(3.14); err != nil {
		t.Fatal(err)
	}
	v, _ := ps.StaticValue()
	if v != 3.14 {
		t.Fatalf("after SetStaticValue: %v, want 3.14", v)
	}
}

func TestPropertyStream_AddKeyframe_TransitionsToAnimated(t *testing.T) {
	ps := codec.NewPropertyStream[float64]()
	_ = ps.SetStaticValue(100.0)

	if err := ps.AddKeyframeLinear(0, 0); err != nil {
		t.Fatal(err)
	}
	if ps.Mode() != codec.StreamModeAnimated {
		t.Fatalf("after AddKeyframe: mode = %v, want Animated", ps.Mode())
	}
	if _, isStatic := ps.StaticValue(); isStatic {
		t.Fatalf("StaticValue() ok = true in Animated mode, want false")
	}
}

func TestPropertyStream_SetStaticValue_InAnimatedMode_Error(t *testing.T) {
	ps := codec.NewPropertyStream[float64]()
	_ = ps.AddKeyframeLinear(0, 0)
	if err := ps.SetStaticValue(99); err == nil {
		t.Fatal("SetStaticValue in Animated mode should error")
	}
}

func TestPropertyStream_Clear_RestoresStaticMode(t *testing.T) {
	ps := codec.NewPropertyStream[float64]()
	_ = ps.SetStaticValue(42.0)
	_ = ps.AddKeyframeLinear(0, 0)
	if err := ps.Clear(); err != nil {
		t.Fatal(err)
	}
	if ps.Mode() != codec.StreamModeStatic {
		t.Fatalf("after Clear: mode = %v, want Static", ps.Mode())
	}
	v, isStatic := ps.StaticValue()
	if !isStatic || v != 42.0 {
		t.Fatalf("after Clear: StaticValue = (%v, %v), want (42.0, true)", v, isStatic)
	}
}

func TestPropertyStream_NegativeTime_Rejected(t *testing.T) {
	ps := codec.NewPropertyStream[float64]()
	if err := ps.AddKeyframeLinear(-0.001, 0); err == nil {
		t.Fatal("negative time should error")
	}
}

func TestPropertyStream_DuplicateTime_Rejected(t *testing.T) {
	ps := codec.NewPropertyStream[float64]()
	_ = ps.AddKeyframeLinear(1.0, 0)
	if err := ps.AddKeyframeLinear(1.0, 99); err == nil {
		t.Fatal("duplicate time should error")
	}
}
