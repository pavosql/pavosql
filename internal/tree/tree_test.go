package tree

import "testing"

func TestTree_Get(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		r       pageReader
		k       []byte
		want    []byte
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := New()
			got, gotErr := tr.Get(tt.r, tt.k)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("Get() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("Get() succeeded unexpectedly")
			}
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("Get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTree_Set(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		wr      pageReadWriter
		k       []byte
		v       []byte
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := New()
			gotErr := tr.Set(tt.wr, tt.k, tt.v)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("Set() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("Set() succeeded unexpectedly")
			}
		})
	}
}

func TestTree_Delete(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		k       []byte
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := New()
			gotErr := tr.Delete(tt.k)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("Delete() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("Delete() succeeded unexpectedly")
			}
		})
	}
}
