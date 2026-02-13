package diff

import "testing"

func TestUnchangedBlock_LineCount(t *testing.T) {
	block := UnchangedBlock{
		Lines: []LinePair{
			{LeftLineNo: 1, RightLineNo: 1},
			{LeftLineNo: 2, RightLineNo: 2},
			{LeftLineNo: 3, RightLineNo: 3},
		},
	}

	if got := block.LineCount(); got != 3 {
		t.Errorf("LineCount() = %d, want 3", got)
	}
}

func TestChangeBlock_LineCount(t *testing.T) {
	block := ChangeBlock{
		Lines: []ChangeLine{
			RemovedLine{LeftLineNo: 1},
			AddedLine{RightLineNo: 1},
			AddedLine{RightLineNo: 2},
		},
	}

	if got := block.LineCount(); got != 3 {
		t.Errorf("LineCount() = %d, want 3", got)
	}
}

func TestFileDiff_TotalLineCount(t *testing.T) {
	fd := &FileDiff{
		Path: "test.go",
		Blocks: []Block{
			UnchangedBlock{Lines: make([]LinePair, 5)},
			ChangeBlock{Lines: make([]ChangeLine, 3)},
			UnchangedBlock{Lines: make([]LinePair, 2)},
		},
	}

	if got := fd.TotalLineCount(); got != 10 {
		t.Errorf("TotalLineCount() = %d, want 10", got)
	}
}

func TestFileDiff_TotalLineCount_Empty(t *testing.T) {
	fd := &FileDiff{Path: "test.go", Blocks: nil}

	if got := fd.TotalLineCount(); got != 0 {
		t.Errorf("TotalLineCount() = %d, want 0", got)
	}
}

func TestBlockInterface_Sealed(t *testing.T) {
	// Verify that Block interface is properly implemented
	var _ Block = UnchangedBlock{}
	var _ Block = ChangeBlock{}
}

func TestChangeLineInterface_Sealed(t *testing.T) {
	// Verify that ChangeLine interface is properly implemented
	var _ ChangeLine = RemovedLine{}
	var _ ChangeLine = AddedLine{}
}
